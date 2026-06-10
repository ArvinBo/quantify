# 开发指南

## 项目结构详解

```
quantify/
├── cmd/qt/main.go                    Go 入口，switch 命令路由
├── internal/
│   ├── cmd/
│   │   ├── init.go                   qt init：验证数据库连接
│   │   └── download.go               qt download：spawn Python 下载
│   ├── config/
│   │   └── config.go                 读 YAML → Config struct
│   ├── db/
│   │   ├── db.go                     GORM 连接管理（Open/Close）
│   │   └── daily_kline_repo.go       日线仓储（CURD 接口 + GORM 实现）
│   └── model/
│       └── dbmodel/
│           └── daily_kline.go        GORM Model → daily_kline_tab 映射
├── python/
│   ├── pyproject.toml                uv 项目定义
│   ├── .venv/                        Python 3.12 虚拟环境
│   ├── quantify/
│   │   ├── config.py                 Python 侧配置加载
│   │   ├── data/
│   │   │   ├── schema.py             建表 DDL
│   │   │   └── downloader.py         akshare 下载 + 写入 SQLite
│   │   ├── factors/                  因子计算
│   │   ├── strategy/                 策略基类
│   │   └── backtest/                 回测引擎（engine / metrics / run）
│   ├── strategies/                   具体策略实现
│   └── tests/
├── config/default.yaml               全局配置
├── data/quantify.db                  SQLite 数据库（gitignore）
├── docs/                             项目文档
├── Makefile
└── .gitignore
```

---

## 环境配置

Go SDK 路径：

```bash
export PATH="/Users/arvinz/sdk/go1.25.4/bin:$PATH"
```

Python 环境：

```bash
# 安装 uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# 安装 Python 依赖
cd python && uv sync
```

---

## 常用命令

```bash
make build           # 编译 Go
make init            # 验证数据库
make download        # 下载全部标的
make list            # 数据概览
```

---

## 如何添加一个新命令

以 `qt list` 为例（展示数据库概览）：

### 1. Go 侧：创建 `internal/cmd/list.go`

```go
package cmd

import (
    "fmt"
    "quantify/internal/config"
    "quantify/internal/db"
)

func ListDB(configPath string) error {
    cfg, _ := config.LoadConfig(configPath)
    database, _ := db.Open(cfg.DB.Path)
    defer db.Close(database)

    repo := db.NewDailyKlineRepo(database)
    stats, _ := repo.GetStats()

    fmt.Printf("Symbols: %d\n", stats.SymbolCount)
    fmt.Printf("Rows:    %d\n", stats.TotalRows)
    fmt.Printf("Date:    %s ~ %s\n", stats.MinDate, stats.MaxDate)
    return nil
}
```

### 2. 注册命令：修改 `cmd/qt/main.go`

```go
case "list":
    runList(os.Args[2:])

func runList(args []string) {
    fs := flag.NewFlagSet("list", flag.ExitOnError)
    configPath := fs.String("c", "config/default.yaml", "config file path")
    fs.Parse(args)
    if err := cmd.ListDB(*configPath); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

### 3. 如果有 Python 任务（如 `qt backtest`）

Go 侧函数和 `internal/cmd/download.go` 结构一样：`exec.Command(venvPython, "-m", "quantify.xxx", ...)`。

Python 侧创建对应的入口文件，如 `quantify/backtest/run.py`：

```python
def main():
    cfg = load_config()
    parser = argparse.ArgumentParser()
    parser.add_argument("--db")
    parser.add_argument("--strategy")
    parser.add_argument("--code")
    # ...
    args = parser.parse_args()
    # 干活...

if __name__ == "__main__":
    main()
```

---

## 如何添加一个新的数据源

数据源切换通过配置 `data.source` 控制，不硬编码。

在 `python/quantify/data/downloader.py` 中添加分支：

```python
def download_single(conn, code, start_date):
    cfg = load_config()
    source = cfg["data"]["source"]

    if source == "akshare":
        return _download_akshare(conn, code, start_date)
    elif source == "xtdata":
        return _download_xtdata(conn, code, start_date)
    else:
        raise ValueError(f"unknown source: {source}")

def _download_akshare(conn, code, start_date):
    # 当前逻辑

def _download_xtdata(conn, code, start_date):
    from xtquant import xtdata
    # xtdata 下载逻辑
```

---

## 如何添加一个新策略

在 `python/strategies/` 下创建文件，遵循统一接口：

```python
# strategies/ma_cross.py
import pandas as pd

class Strategy:
    def __init__(self, fast=5, slow=20):
        self.fast = fast
        self.slow = slow

    def generate(self, df: pd.DataFrame) -> pd.DataFrame:
        """
        输入: 日线 DataFrame
        列: trade_date, open, high, low, close, volume, amount
        输出: 附加 signal 列 (1=买, -1=卖, 0=持有)
        """
        df = df.copy()
        df["ma_fast"] = df["close"].rolling(self.fast).mean()
        df["ma_slow"] = df["close"].rolling(self.slow).mean()
        df["signal"] = 0
        df.loc[df["ma_fast"] > df["ma_slow"], "signal"] = 1
        df.loc[df["ma_fast"] < df["ma_slow"], "signal"] = -1
        return df
```

回测引擎动态加载策略：

```python
import importlib
strategy_mod = importlib.import_module(f"strategies.{args.strategy}")
strategy = strategy_mod.Strategy(fast=5, slow=20)
signals = strategy.generate(df)
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
        4. engine.run(df, signals, capital=100000) → 模拟交易
        5. metrics 计算：收益率、夏普、回撤、胜率
        6. 结果打印到终端
        7. 可选：写回 backtest_results 表
```

---

## 实盘流程（未来）

```
qt run

  Python: 因子计算 → 产生信号 → INSERT INTO signals

  Go: 轮询 signals 表
      ├─ 风控检查（仓位/亏损/资金）
      ├─ 通过 → xttrader 下单
      └─ 拒绝 → 日志/报警
```

---

## 数据库表

| 表名 | 用途 | 阶段 |
|------|------|------|
| `daily_kline_tab` | A 股日线行情 | 已实现 |
| `signals` | 策略信号 | 规划中 |
| `backtest_results` | 回测结果 | 规划中 |
| `orders` | 实盘订单记录 | 规划中 |
| `positions` | 当前持仓 | 规划中 |

---

## 开发路线图

| 阶段 | 状态 | 内容 |
|------|------|------|
| **一** | 已完成 | 项目骨架、Go CLI、配置系统、akshare 数据下载 |
| **二** | 待开始 | 回测引擎（engine + metrics + run）、策略示例（ma_cross） |
| **三** | 规划中 | 因子框架、更多策略、参数优化 |
| **四** | 未来 | 接入 QMT xtquant、信号-下单通道、风控层 |
