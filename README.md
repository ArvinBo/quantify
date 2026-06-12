# Quantify

量化交易研究平台，基于 QMT（国金证券）的自建量化系统。

不做 QMT 官方工具的二次开发，而是从零搭建自己的架构。

---

## 核心理念

**Go 管流程和安全，Python 管计算和策略。**

| | Go | Python |
|---|---|---|
| 角色 | 调度者 · 守门人 | 执行者 · 研究员 |
| 能做什么 | 启动任务、读数据、风控、下单 | 下载数据、算因子、跑回测、生成信号 |
| 不能做什么 | 不算因子、不跑策略逻辑 | 不直接下单、不接触风控 |

---

## 快速开始

```bash
make python-install   # 安装 Python 依赖（uv）
make build            # 编译 Go CLI
make init             # 初始化数据库
make download         # 下载行情数据
```

```bash
# 下载单只标的
./bin/qt download -s 600519.SH -f 2024-01-01

# 跑回测
./bin/qt backtest --strategy ma_cross --code 600519.SH --start 2020-01-01 --end 2024-12-31

# 实盘流程（信号生成 → 风控 → 下单，当前 dry-run 模拟成交）
./bin/qt run

# 查看数据与持仓概览
./bin/qt stats
```

修改 Go 代码后重新编译：

```bash
make build
# 或直接执行
go build -o bin/qt ./cmd/qt
```

---

## 项目结构

```
quantify/
├── cmd/qt/           Go CLI 入口
├── internal/         Go 内部包（cmd / config / db / model / risk / trader）
├── python/           Python 量化包（data / strategy / backtest / live）
│   └── strategies/   具体策略实现
├── config/           全局 YAML 配置（Go 和 Python 共享）
├── data/             SQLite 数据库（gitignore）
├── Makefile
└── docs/             详细文档
```

---

## 配置

```yaml
# config/default.yaml
db:
  path: ./data/quantify.db

data:
  source: akshare               # akshare | tushare
  start_date: "2015-01-01"
  symbols:
    - "000001.SZ"
    - "600519.SH"
  tushare:
    token: ""                   # tushare.pro token，留空读环境变量 TUSHARE_TOKEN
```

### 数据源

| 数据源 | 说明 |
|--------|------|
| akshare | 免费，无需注册；接口走东方财富，海外网络可能不可达 |
| [tushare.pro](https://tushare.pro) | 需注册取 token（个人中心 → 接口 token），稳定，日线 + 复权 + 股票列表 |

切换到 tushare：把 `data.source` 改为 `tushare`，token 填入 `data.tushare.token` 或设置环境变量 `TUSHARE_TOKEN`。

---

## 详细文档

| 文档 | 内容 |
|------|------|
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | 架构设计：两阶段架构、Go/Python 分工、三种协作机制全景 |
| [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) | 开发指南：项目结构详解、如何加命令/策略/数据源、路线图 |

---

## 路线图

| 阶段 | 状态 | 内容 |
|------|------|------|
| 一 | 已完成 | 项目骨架、Go CLI、数据下载（akshare / tushare） |
| 二 | 已完成 | 回测引擎、绩效指标、策略示例（ma_cross） |
| 三 | 已完成 | 实盘骨架：信号生成 → 风控 → 下单通道（dry-run） |
| 四 | 规划中 | 因子框架、参数优化、更多策略 |
| 五 | 未来 | 接入 QMT xtquant 实盘下单、实时行情风控 |
