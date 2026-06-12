"""实盘信号生成（由 Go `qt run` spawn）。

    python -m quantify.live.signal_gen --db <path> --strategy ma_cross

对 config 中每个 symbol：取最近 N 根日线 → 跑策略 → 最新 bar 出现买/卖信号
则写入 signal_tab（状态 PENDING）。Python 只产出信号，下单与风控由 Go 负责。
"""

from __future__ import annotations

import argparse
import importlib
import sqlite3

import pandas as pd
from loguru import logger

from quantify.config import load_config
from quantify.data.schema import init_db

LOOKBACK_BARS = 120  # 取足够覆盖慢均线窗口的历史


def load_recent(conn: sqlite3.Connection, code: str, limit: int) -> pd.DataFrame:
    df = pd.read_sql_query(
        "SELECT trade_date, open, high, low, close, volume, amount "
        "FROM daily_kline_tab WHERE code = ? ORDER BY trade_date DESC LIMIT ?",
        conn,
        params=(code, limit),
    )
    return df.iloc[::-1].reset_index(drop=True)


def signal_exists(conn: sqlite3.Connection, code: str, trade_date: str, action: str) -> bool:
    row = conn.execute(
        "SELECT 1 FROM signal_tab WHERE code = ? AND trade_date = ? AND action = ?",
        (code, trade_date, action),
    ).fetchone()
    return row is not None


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--db", required=True)
    parser.add_argument("--strategy", default="ma_cross")
    args = parser.parse_args()

    cfg = load_config()
    mod = importlib.import_module(f"strategies.{args.strategy}")
    strategy = mod.Strategy()

    conn = sqlite3.connect(args.db)
    init_db(conn)

    created = 0
    try:
        for code in cfg["data"]["symbols"]:
            df = load_recent(conn, code, LOOKBACK_BARS)
            if df.empty:
                logger.warning(f"{code}: no kline data, skip")
                continue

            df = strategy.generate(df)
            last = df.iloc[-1]
            sig = int(last["signal"])
            if sig == 0:
                continue

            action = "BUY" if sig == 1 else "SELL"
            trade_date = str(last["trade_date"])
            if signal_exists(conn, code, trade_date, action):
                logger.info(f"{code}: {action}@{trade_date} already exists, skip")
                continue

            conn.execute(
                "INSERT INTO signal_tab (code, action, price, strategy, trade_date) "
                "VALUES (?, ?, ?, ?, ?)",
                (code, action, float(last["close"]), args.strategy, trade_date),
            )
            created += 1
            logger.info(f"{code}: signal {action} @ {last['close']} ({trade_date})")

        conn.commit()
        logger.info(f"signal generation done: {created} new signals")
    finally:
        conn.close()


if __name__ == "__main__":
    main()
