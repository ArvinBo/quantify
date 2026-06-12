"""双均线策略：快线上穿慢线买入，下穿卖出。"""

from __future__ import annotations

import pandas as pd

from quantify.strategy.base import StrategyBase


class Strategy(StrategyBase):
    name = "ma_cross"

    def __init__(self, fast: int = 5, slow: int = 20):
        self.fast = fast
        self.slow = slow

    def generate(self, df: pd.DataFrame) -> pd.DataFrame:
        df = df.copy()
        df["ma_fast"] = df["close"].rolling(self.fast).mean()
        df["ma_slow"] = df["close"].rolling(self.slow).mean()

        df["signal"] = 0
        above = df["ma_fast"] > df["ma_slow"]
        below = df["ma_fast"] < df["ma_slow"]
        # 仅在穿越当日发出信号，而不是持续发出
        df.loc[above & ~above.shift(1, fill_value=False), "signal"] = 1
        df.loc[below & ~below.shift(1, fill_value=False), "signal"] = -1
        # 均线未形成前不发信号
        df.loc[df["ma_slow"].isna(), "signal"] = 0
        return df
