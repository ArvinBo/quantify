"""回测绩效指标：收益率、年化、最大回撤、夏普、胜率。"""

from __future__ import annotations

import math

import pandas as pd

from quantify.backtest.engine import Trade

TRADING_DAYS_PER_YEAR = 252


def compute(equity: pd.Series, trades: list[Trade], capital: float) -> dict:
    if equity.empty:
        return {
            "total_return": 0.0,
            "annual_return": 0.0,
            "max_drawdown": 0.0,
            "sharpe": 0.0,
            "win_rate": 0.0,
            "num_trades": 0,
        }

    final = float(equity.iloc[-1])
    total_return = final / capital - 1

    n_days = len(equity)
    if n_days > 1 and final > 0:
        annual_return = (final / capital) ** (TRADING_DAYS_PER_YEAR / n_days) - 1
    else:
        annual_return = 0.0

    running_max = equity.cummax()
    drawdown = equity / running_max - 1
    max_drawdown = float(drawdown.min())

    daily_returns = equity.pct_change().dropna()
    std = float(daily_returns.std())
    if std > 0:
        sharpe = float(daily_returns.mean()) / std * math.sqrt(TRADING_DAYS_PER_YEAR)
    else:
        sharpe = 0.0

    closed = [t for t in trades if t.exit_date]
    wins = sum(1 for t in closed if t.pnl > 0)
    win_rate = wins / len(closed) if closed else 0.0

    return {
        "total_return": total_return,
        "annual_return": annual_return,
        "max_drawdown": max_drawdown,
        "sharpe": sharpe,
        "win_rate": win_rate,
        "num_trades": len(closed),
    }
