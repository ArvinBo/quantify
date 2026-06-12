"""行情数据源适配层。

所有数据源统一输出标准化 DataFrame，列为：
    code, trade_date(YYYY-MM-DD), open, high, low, close, volume(股), amount(元)

价格统一为前复权（qfq），便于回测。
切换数据源由 config 的 data.source 控制：akshare | tushare
"""

from __future__ import annotations

import os
import sqlite3

import pandas as pd
from loguru import logger

from quantify.config import load_config


def fetch_daily(source: str, code: str, start_date: str) -> pd.DataFrame:
    """按数据源拉取单只标的日线，返回标准化 DataFrame（可能为空）。"""
    if source == "akshare":
        return _fetch_akshare(code, start_date)
    if source == "tushare":
        return _fetch_tushare(code, start_date)
    raise ValueError(f"unknown data source: {source}")


def sync_stock_info(source: str, conn: sqlite3.Connection) -> int:
    """同步股票基本信息到 stock_info_tab，返回写入条数。"""
    if source == "akshare":
        return _sync_info_akshare(conn)
    if source == "tushare":
        return _sync_info_tushare(conn)
    raise ValueError(f"unknown data source: {source}")


# ---------------------------------------------------------------- akshare

def _fetch_akshare(code: str, start_date: str) -> pd.DataFrame:
    import akshare as ak

    symbol = code.split(".")[0]  # akshare 用纯数字代码
    df = ak.stock_zh_a_hist(
        symbol=symbol,
        period="daily",
        start_date=start_date.replace("-", ""),
        adjust="qfq",
    )
    if df is None or df.empty:
        return pd.DataFrame()

    out = pd.DataFrame(
        {
            "code": code,
            "trade_date": pd.to_datetime(df["日期"]).dt.strftime("%Y-%m-%d"),
            "open": df["开盘"].astype(float),
            "high": df["最高"].astype(float),
            "low": df["最低"].astype(float),
            "close": df["收盘"].astype(float),
            "volume": df["成交量"].astype(float) * 100,  # 手 → 股
            "amount": df["成交额"].astype(float),
        }
    )
    return out


def _sync_info_akshare(conn: sqlite3.Connection) -> int:
    import akshare as ak

    df = ak.stock_info_a_code_name()
    rows = []
    for _, r in df.iterrows():
        raw = str(r["code"])
        market = _market_of(raw)
        if not market:
            continue
        rows.append((f"{raw}.{market}", str(r["name"]), market, "", ""))

    conn.executemany(
        "INSERT OR REPLACE INTO stock_info_tab (code, name, market, industry, list_date) "
        "VALUES (?, ?, ?, ?, ?)",
        rows,
    )
    conn.commit()
    return len(rows)


# ---------------------------------------------------------------- tushare

def _tushare_token() -> str:
    cfg = load_config()
    token = cfg.get("data", {}).get("tushare", {}).get("token") or os.environ.get(
        "TUSHARE_TOKEN", ""
    )
    if not token:
        raise RuntimeError(
            "tushare token 未配置：在 config/default.yaml 的 data.tushare.token 填入，"
            "或设置环境变量 TUSHARE_TOKEN（https://tushare.pro 个人中心获取）"
        )
    return token


def _fetch_tushare(code: str, start_date: str) -> pd.DataFrame:
    import tushare as ts

    ts.set_token(_tushare_token())
    # pro_bar 支持前复权；tushare 代码格式与本项目一致（600519.SH）
    df = ts.pro_bar(
        ts_code=code,
        adj="qfq",
        asset="E",
        start_date=start_date.replace("-", ""),
    )
    if df is None or df.empty:
        return pd.DataFrame()

    df = df.sort_values("trade_date")
    out = pd.DataFrame(
        {
            "code": code,
            "trade_date": pd.to_datetime(df["trade_date"]).dt.strftime("%Y-%m-%d"),
            "open": df["open"].astype(float),
            "high": df["high"].astype(float),
            "low": df["low"].astype(float),
            "close": df["close"].astype(float),
            "volume": df["vol"].astype(float) * 100,      # 手 → 股
            "amount": df["amount"].astype(float) * 1000,  # 千元 → 元
        }
    )
    return out


def _sync_info_tushare(conn: sqlite3.Connection) -> int:
    import tushare as ts

    pro = ts.pro_api(_tushare_token())
    df = pro.stock_basic(
        exchange="", list_status="L", fields="ts_code,name,industry,list_date"
    )
    rows = []
    for _, r in df.iterrows():
        code = str(r["ts_code"])
        market = code.split(".")[-1]
        list_date = str(r["list_date"] or "")
        if len(list_date) == 8:
            list_date = f"{list_date[:4]}-{list_date[4:6]}-{list_date[6:]}"
        rows.append((code, str(r["name"]), market, str(r["industry"] or ""), list_date))

    conn.executemany(
        "INSERT OR REPLACE INTO stock_info_tab (code, name, market, industry, list_date) "
        "VALUES (?, ?, ?, ?, ?)",
        rows,
    )
    conn.commit()
    return len(rows)


def _market_of(raw_code: str) -> str:
    if raw_code.startswith(("60", "68")):
        return "SH"
    if raw_code.startswith(("00", "30")):
        return "SZ"
    if raw_code.startswith(("43", "83", "87", "92")):
        return "BJ"
    logger.debug(f"skip unknown market code: {raw_code}")
    return ""
