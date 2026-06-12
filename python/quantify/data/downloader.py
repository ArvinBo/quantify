"""数据下载入口（由 Go `qt download` spawn）。

    python -m quantify.data.downloader --db <path> [--code 600519.SH] [--from 2024-01-01]
                                       [--all] [--sync-info]

增量逻辑：未指定 --from 时，从库中该标的最新交易日的次日开始下载；
库中无数据则从 config 的 data.start_date 开始。
"""

from __future__ import annotations

import argparse
import sqlite3
from datetime import datetime, timedelta
from pathlib import Path

from loguru import logger

from quantify.config import load_config
from quantify.data import sources
from quantify.data.schema import init_db

INSERT_SQL = (
    "INSERT OR REPLACE INTO daily_kline_tab "
    "(code, trade_date, open, high, low, close, volume, amount) "
    "VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
)


def latest_trade_date(conn: sqlite3.Connection, code: str) -> str | None:
    row = conn.execute(
        "SELECT MAX(trade_date) FROM daily_kline_tab WHERE code = ?", (code,)
    ).fetchone()
    return row[0] if row and row[0] else None


def resolve_start_date(
    conn: sqlite3.Connection, code: str, from_arg: str, default_start: str
) -> str | None:
    """返回下载起始日；已是最新则返回 None。"""
    if from_arg:
        return from_arg

    latest = latest_trade_date(conn, code)
    if latest is None:
        return default_start

    next_day = datetime.strptime(latest, "%Y-%m-%d") + timedelta(days=1)
    if next_day.date() > datetime.now().date():
        return None
    return next_day.strftime("%Y-%m-%d")


def download_single(
    conn: sqlite3.Connection, source: str, code: str, start_date: str
) -> int:
    df = sources.fetch_daily(source, code, start_date)
    if df.empty:
        logger.info(f"{code}: no new data since {start_date}")
        return 0

    rows = df[
        ["code", "trade_date", "open", "high", "low", "close", "volume", "amount"]
    ].values.tolist()
    conn.executemany(INSERT_SQL, rows)
    conn.commit()
    logger.info(f"{code}: {len(rows)} rows ({df['trade_date'].min()} ~ {df['trade_date'].max()})")
    return len(rows)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--db", required=True)
    parser.add_argument("--code", default="")
    parser.add_argument("--from", dest="from_date", default="")
    parser.add_argument("--all", action="store_true")
    parser.add_argument("--sync-info", action="store_true")
    args = parser.parse_args()

    cfg = load_config()
    source = cfg["data"]["source"]
    default_start = cfg["data"]["start_date"]

    Path(args.db).parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(args.db)
    init_db(conn)

    try:
        if args.sync_info:
            n = sources.sync_stock_info(source, conn)
            logger.info(f"stock info synced: {n} rows (source={source})")

        codes = cfg["data"]["symbols"] if args.all or not args.code else [args.code]
        total = 0
        for code in codes:
            start = resolve_start_date(conn, code, args.from_date, default_start)
            if start is None:
                logger.info(f"{code}: already up to date")
                continue
            try:
                total += download_single(conn, source, code, start)
            except Exception as e:  # 单只失败不中断整体下载
                logger.error(f"{code}: download failed: {e}")

        logger.info(f"done. total {total} rows written (source={source})")
    finally:
        conn.close()


if __name__ == "__main__":
    main()
