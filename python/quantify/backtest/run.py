"""回测入口（由 Go `qt backtest` spawn）。

    python -m quantify.backtest.run --db <path> --strategy ma_cross --code 600519.SH \
        --start 2020-01-01 --end 2024-12-31

流程：SQLite 取行情 → 动态加载策略 → 引擎模拟撮合 → 输出指标 → 写 backtest_result_tab
"""

from __future__ import annotations

import argparse
import importlib
import sqlite3

import pandas as pd
from loguru import logger

from quantify.backtest.engine import Engine
from quantify.backtest import metrics
from quantify.config import load_config
from quantify.data.schema import init_db


def load_klines(conn: sqlite3.Connection, code: str, start: str, end: str) -> pd.DataFrame:
    return pd.read_sql_query(
        "SELECT trade_date, open, high, low, close, volume, amount "
        "FROM daily_kline_tab WHERE code = ? AND trade_date >= ? AND trade_date <= ? "
        "ORDER BY trade_date",
        conn,
        params=(code, start, end),
    )


def load_strategy(name: str):
    mod = importlib.import_module(f"strategies.{name}")
    return mod.Strategy()


def save_result(conn: sqlite3.Connection, args, capital: float, m: dict) -> None:
    conn.execute(
        "INSERT INTO backtest_result_tab "
        "(strategy, code, start_date, end_date, capital, total_return, annual_return, "
        " max_drawdown, sharpe, win_rate, num_trades) "
        "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
        (
            args.strategy, args.code, args.start, args.end, capital,
            m["total_return"], m["annual_return"], m["max_drawdown"],
            m["sharpe"], m["win_rate"], m["num_trades"],
        ),
    )
    conn.commit()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--db", required=True)
    parser.add_argument("--strategy", required=True)
    parser.add_argument("--code", required=True)
    parser.add_argument("--start", default="2015-01-01")
    parser.add_argument("--end", default="2099-12-31")
    args = parser.parse_args()

    cfg = load_config()
    bt = cfg.get("backtest", {})
    capital = float(bt.get("capital", 100000))

    conn = sqlite3.connect(args.db)
    init_db(conn)
    try:
        df = load_klines(conn, args.code, args.start, args.end)
        if df.empty:
            logger.error(f"no kline data for {args.code} in [{args.start}, {args.end}], "
                         "run `qt download` first")
            raise SystemExit(1)

        strategy = load_strategy(args.strategy)
        df = strategy.generate(df)

        engine = Engine(
            capital=capital,
            commission_rate=float(bt.get("commission_rate", 0.0003)),
            commission_min=float(bt.get("commission_min", 5)),
            stamp_tax_rate=float(bt.get("stamp_tax_rate", 0.0005)),
        )
        result = engine.run(df)
        m = metrics.compute(result.equity, result.trades, capital)

        print(f"\n===== Backtest: {args.strategy} on {args.code} =====")
        print(f"period       : {df['trade_date'].iloc[0]} ~ {df['trade_date'].iloc[-1]}")
        print(f"capital      : {capital:,.0f}")
        print(f"final equity : {result.equity.iloc[-1]:,.2f}")
        print(f"total return : {m['total_return']:+.2%}")
        print(f"annual return: {m['annual_return']:+.2%}")
        print(f"max drawdown : {m['max_drawdown']:.2%}")
        print(f"sharpe       : {m['sharpe']:.2f}")
        print(f"win rate     : {m['win_rate']:.2%} ({m['num_trades']} trades)")

        save_result(conn, args, capital, m)
        logger.info("result saved to backtest_result_tab")
    finally:
        conn.close()


if __name__ == "__main__":
    main()
