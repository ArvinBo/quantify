# 开发指南

## 项目结构详解

```
quantify/
├── cmd/qt/main.go                    Go 入口，switch 命令路由
├── internal/
│   ├── cmd/
│   │   ├── python.go                 公共函数：spawn Python 子进程
│   │   ├── init.go                   qt init：验证数据库连接
│   │   ├── download.go               qt download：spawn Python 下载
│   │   ├── backtest.go               qt backtest：spawn Python 回测
│   │   ├── run.go                    qt run：信号生成 → 风控 → 下单
│   │   └── stats.go                  qt stats：行情统计 + 持仓概览
│   ├── config/
│   │   └── config.go                 读 YAML → Config struct
│   ├── db/
│   │   ├── db.go                     GORM 连接管理（Open/Close）
│   │   ├── daily_kline_repo.go       日线仓储
│   │   ├── stock_info_repo.go        股票信息仓储
│   │   ├── signal_repo.go            信号仓储（FindPending / UpdateStatus）
│   │   ├── order_repo.go             订单仓储
│   │   └── position_repo.go          持仓仓储（ApplyFill 按成交更新）
│   ├── risk/
│   │   └── risk.go                   下单前风控检查（资金/仓位/持仓）
│   ├── trader/
│   │   └── trader.go                 下单通道接口 + dry-run 实现
│   └── model/dbmodel/                GORM Model → 表映射
├── python/
│   ├── pyproject.toml                uv 项目定义
│   ├── .venv/                        Python 虚拟环境（uv sync 生成）
│   ├── quantify/
│   │   ├── config.py                 Python 侧配置加载
│   │   ├── data/
│   │   │   ├── schema.py             全部表的建表 DDL（唯一建表入口）
│   │   │   ├── sources.py            数据源适配：akshare / tushare → 标准化
│   │   │   └── downloader.py         下载入口（增量下载 + 写入 SQLite）
│   │   ├── strategy/
│   │   │   └── base.py               策略基类（StrategyBase.generate）
│   │   ├── backtest/
│   │   │   ├── engine.py             模拟撮合（T+1 开盘成交、佣金印花税）
│   │   │   ├── metrics.py            绩效指标（收益/年化/回撤/夏普/胜率）
│   │   │   └── run.py                回测入口
│   │   ├── live/
│   │   │   └── signal_gen.py         实盘信号生成（写 signal_tab）
│   │   └── factors/                  因子计算（规划中）
│   ├── strategies/                   具体策略实现（ma_cross）
│   └── tests/
├── config/default.yaml               全局配置
├── data/quantify.db                  SQLite 数据库（gitignore）
├── docs/                             项目文档
├── Makefile
└── .gitignore
```

---

## 环境配置

Go 1.21+（go.mod 要求）。若系统 Go 过旧，可下载到用户目录并通过 `make build GO=...` 指定：

```bash
curl -sL https://go.dev/dl/go1.25.4.darwin-arm64.tar.gz | tar -xz -C ~/sdk
make build GO=~/sdk/go1.25.4/bin/go
```

Python 环境：

```bash
# 安装 uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# 安装 Python 依赖
cd python && uv sync
```

tushare 数据源需要 token（[tushare.pro](https://tushare.pro) 注册后在个人中心获取）：

```bash
# 方式一：写入 config/default.yaml 的 data.tushare.token
# 方式二：环境变量
export TUSHARE_TOKEN=xxxx
```

---

## 常用命令

```bash
make build           # 编译 Go
make init            # 验证数据库
make download        # 下载全部标的（数据源由 data.source 决定）
make backtest        # 回测 ma_cross on 600519.SH
make run             # 实盘流程（dry-run）
make stats           # 数据与持仓概览
# 手动查库: sqlite3 data/quantify.db "SELECT COUNT(*) FROM daily_kline_tab"
```

---

## 如何添加一个新命令

参考 `qt backtest` 的现成实现，三步：

1. **Go 侧**：在 `internal/cmd/` 新建文件，加载配置后调 `runPython("quantify.xxx.yyy", args...)`
   （见 `internal/cmd/backtest.go`，spawn 逻辑已抽到 `internal/cmd/python.go`）
2. **注册命令**：在 `cmd/qt/main.go` 的 switch 中加 case，用 `flag.NewFlagSet` 解析参数
3. **Python 侧**：在 `python/quantify/` 下新建模块，argparse 收参数（见 `quantify/backtest/run.py`）

---

## 如何添加一个新的数据源

数据源切换由配置 `data.source` 控制，适配层在 `python/quantify/data/sources.py`。

新数据源只需两件事：

```python
# sources.py
def fetch_daily(source, code, start_date):
    if source == "akshare":
        return _fetch_akshare(code, start_date)
    if source == "tushare":
        return _fetch_tushare(code, start_date)
    if source == "xtdata":                  # ← 新增分支
        return _fetch_xtdata(code, start_date)

def _fetch_xtdata(code, start_date) -> pd.DataFrame:
    # 返回标准化 DataFrame：
    # code, trade_date(YYYY-MM-DD), open, high, low, close, volume(股), amount(元)
    # 价格统一为前复权
    ...
```

注意单位换算：A 股接口常以「手」计成交量（×100 → 股）、tushare 以「千元」计成交额（×1000 → 元）。

---

## 如何添加一个新策略

在 `python/strategies/` 下创建文件，继承 `StrategyBase`：

```python
# strategies/my_strategy.py
import pandas as pd
from quantify.strategy.base import StrategyBase

class Strategy(StrategyBase):
    name = "my_strategy"

    def generate(self, df: pd.DataFrame) -> pd.DataFrame:
        """
        输入: 日线 DataFrame
        列: trade_date, open, high, low, close, volume, amount
        输出: 附加 signal 列 (1=买, -1=卖, 0=持有)
        注意: 信号只在触发当日发出（参考 ma_cross 的穿越检测），
              不要持续输出 1/-1，否则实盘会反复下单
        """
        ...
```

类名必须叫 `Strategy`，回测与信号生成都按 `strategies.{name}.Strategy` 动态加载：

```bash
./bin/qt backtest --strategy my_strategy --code 600519.SH
./bin/qt run --strategy my_strategy
```

---

## 回测流程

```
qt backtest --strategy ma_cross --code 600519.SH --start 2020-01-01 --end 2024-12-31

  Go cmd/backtest.go
    └─→ exec: python -m quantify.backtest.run --db x --strategy ma_cross --code 600519.SH ...

      Python quantify/backtest/run.py
        1. 从 SQLite 查日线数据
        2. 动态加载 strategies/ma_cross.py
        3. strategy.generate(df) → 信号
        4. engine.run(df) → 模拟交易（T 日信号 T+1 开盘成交，含佣金/印花税）
        5. metrics 计算：收益率、年化、夏普、回撤、胜率
        6. 结果打印到终端 + 写入 backtest_result_tab
```

---

## 实盘流程

```
qt run [--strategy ma_cross]        # 单次执行，可由 cron 定时调度

  step 1  Go spawn → python -m quantify.live.signal_gen
          最近 120 根 K 线跑策略，最新 bar 有买卖信号 → INSERT signal_tab (PENDING)
          （同 code+日期+方向 去重，不会重复写入）

  step 2  Go 读 PENDING 信号，逐条处理：
          ├─ 定量：买入按 risk.capital × max_position_pct 预算取整百股；卖出全仓
          ├─ 风控 internal/risk：资金够吗？仓位超限吗？持仓够卖吗？
          ├─ 通过 → trader.Place（dry-run 模拟成交 / 未来 qmt 真实下单）
          │         → 写 order_tab + 更新 position_tab → 信号置 EXECUTED
          └─ 拒绝 → 信号置 REJECTED，reason 记录原因
```

切换到真实下单：实现 `internal/trader` 的 `Trader` 接口（QMT xttrader，仅 Windows），
并把配置 `trade.mode` 改为 `qmt`。风控与编排逻辑无需改动。

---

## 数据库表

所有表由 Python `quantify/data/schema.py` 统一建表，Go 侧只做 GORM 映射。

| 表名 | 用途 | 写入方 | 读取方 |
|------|------|--------|--------|
| `daily_kline_tab` | A 股日线行情 | Python | Python / Go |
| `stock_info_tab` | 股票基本信息 | Python | Go |
| `signal_tab` | 策略信号 | Python（产出）/ Go（更新状态） | Go |
| `order_tab` | 订单记录 | Go | Go |
| `position_tab` | 当前持仓 | Go | Go |
| `backtest_result_tab` | 回测结果 | Python | - |

---

## 开发路线图

| 阶段 | 状态 | 内容 |
|------|------|------|
| **一** | 已完成 | 项目骨架、Go CLI、配置系统、数据下载（akshare / tushare） |
| **二** | 已完成 | 回测引擎（engine + metrics + run）、策略示例（ma_cross） |
| **三** | 已完成 | 实盘骨架：signal_gen → 风控 → dry-run 下单 → 订单/持仓记账 |
| **四** | 规划中 | 因子框架、更多策略、参数优化 |
| **五** | 未来 | 接入 QMT xtquant 真实下单、实时行情、日内风控（日亏损上限） |
