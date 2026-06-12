"""全部表的建表 DDL。Python 是建表的唯一入口，Go 侧只做 ORM 映射。"""

from __future__ import annotations

import sqlite3

DDL = [
    """
    CREATE TABLE IF NOT EXISTS daily_kline_tab (
        code       TEXT NOT NULL,
        trade_date TEXT NOT NULL,
        open       REAL NOT NULL,
        high       REAL NOT NULL,
        low        REAL NOT NULL,
        close      REAL NOT NULL,
        volume     REAL NOT NULL,
        amount     REAL NOT NULL,
        PRIMARY KEY (code, trade_date)
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS stock_info_tab (
        code      TEXT PRIMARY KEY,
        name      TEXT NOT NULL DEFAULT '',
        market    TEXT NOT NULL DEFAULT '',
        industry  TEXT NOT NULL DEFAULT '',
        list_date TEXT NOT NULL DEFAULT ''
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS signal_tab (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        code       TEXT NOT NULL,
        action     TEXT NOT NULL,                  -- BUY / SELL
        price      REAL NOT NULL,                  -- 参考价（信号日收盘价）
        strategy   TEXT NOT NULL,
        trade_date TEXT NOT NULL,                  -- 信号产生的交易日
        status     TEXT NOT NULL DEFAULT 'PENDING',-- PENDING / EXECUTED / REJECTED
        reason     TEXT NOT NULL DEFAULT '',
        created_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime'))
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS order_tab (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        signal_id  INTEGER NOT NULL,
        code       TEXT NOT NULL,
        action     TEXT NOT NULL,
        price      REAL NOT NULL,
        qty        INTEGER NOT NULL,
        status     TEXT NOT NULL DEFAULT 'FILLED', -- dry-run 模式即时成交
        created_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime'))
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS position_tab (
        code       TEXT PRIMARY KEY,
        qty        INTEGER NOT NULL,
        avg_cost   REAL NOT NULL,
        updated_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime'))
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS backtest_result_tab (
        id            INTEGER PRIMARY KEY AUTOINCREMENT,
        strategy      TEXT NOT NULL,
        code          TEXT NOT NULL,
        start_date    TEXT NOT NULL,
        end_date      TEXT NOT NULL,
        capital       REAL NOT NULL,
        total_return  REAL NOT NULL,
        annual_return REAL NOT NULL,
        max_drawdown  REAL NOT NULL,
        sharpe        REAL NOT NULL,
        win_rate      REAL NOT NULL,
        num_trades    INTEGER NOT NULL,
        created_at    TEXT NOT NULL DEFAULT (datetime('now', 'localtime'))
    )
    """,
]


def init_db(conn: sqlite3.Connection) -> None:
    for ddl in DDL:
        conn.execute(ddl)
    conn.commit()
