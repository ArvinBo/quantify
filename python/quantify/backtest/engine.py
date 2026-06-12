"""回测引擎：基于信号的逐 bar 模拟撮合。

规则：
- 信号在 T 日收盘产生，T+1 日开盘价执行（避免未来函数）
- 全仓买入 / 全仓卖出，按 100 股一手取整
- 买入收佣金，卖出收佣金 + 印花税
"""

from __future__ import annotations

from dataclasses import dataclass, field

import pandas as pd


@dataclass
class Trade:
    entry_date: str
    entry_price: float
    qty: int
    exit_date: str = ""
    exit_price: float = 0.0
    pnl: float = 0.0


@dataclass
class BacktestResult:
    equity: pd.Series = field(default_factory=pd.Series)  # 索引为 trade_date 的净值序列
    trades: list[Trade] = field(default_factory=list)


class Engine:
    def __init__(
        self,
        capital: float = 100000,
        commission_rate: float = 0.0003,
        commission_min: float = 5,
        stamp_tax_rate: float = 0.0005,
    ):
        self.capital = capital
        self.commission_rate = commission_rate
        self.commission_min = commission_min
        self.stamp_tax_rate = stamp_tax_rate

    def _commission(self, turnover: float) -> float:
        return max(turnover * self.commission_rate, self.commission_min)

    def run(self, df: pd.DataFrame) -> BacktestResult:
        """df 需含列：trade_date, open, close, signal。"""
        cash = self.capital
        qty = 0
        open_trade: Trade | None = None
        trades: list[Trade] = []
        equity_values: list[float] = []

        signals = df["signal"].to_numpy()
        opens = df["open"].to_numpy()
        closes = df["close"].to_numpy()
        dates = df["trade_date"].to_numpy()

        for i in range(len(df)):
            if i > 0:
                prev_signal = signals[i - 1]
                price = opens[i]

                if prev_signal == 1 and qty == 0 and price > 0:
                    lots = int(cash / (price * 100 * (1 + self.commission_rate)))
                    if lots > 0:
                        qty = lots * 100
                        turnover = price * qty
                        cash -= turnover + self._commission(turnover)
                        open_trade = Trade(
                            entry_date=str(dates[i]), entry_price=price, qty=qty
                        )

                elif prev_signal == -1 and qty > 0:
                    turnover = price * qty
                    fee = self._commission(turnover) + turnover * self.stamp_tax_rate
                    cash += turnover - fee
                    if open_trade is not None:
                        open_trade.exit_date = str(dates[i])
                        open_trade.exit_price = price
                        open_trade.pnl = (
                            (price - open_trade.entry_price) * open_trade.qty - fee
                        )
                        trades.append(open_trade)
                        open_trade = None
                    qty = 0

            equity_values.append(cash + qty * closes[i])

        # 期末未平仓的持仓按最后收盘价计入浮动盈亏
        if open_trade is not None:
            open_trade.exit_date = str(dates[-1])
            open_trade.exit_price = float(closes[-1])
            open_trade.pnl = (open_trade.exit_price - open_trade.entry_price) * open_trade.qty
            trades.append(open_trade)

        equity = pd.Series(equity_values, index=df["trade_date"].tolist())
        return BacktestResult(equity=equity, trades=trades)
