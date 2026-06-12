"""策略基类。所有策略实现 generate()，输入日线 DataFrame，输出附加 signal 列的 DataFrame。"""

from __future__ import annotations

import pandas as pd


class StrategyBase:
    """策略统一接口。

    generate() 输入列：trade_date, open, high, low, close, volume, amount
    输出：原 DataFrame 附加 signal 列（1=买入, -1=卖出, 0=持有/观望）
    """

    name = "base"

    def generate(self, df: pd.DataFrame) -> pd.DataFrame:
        raise NotImplementedError
