# Quantify

量化交易研究平台 —— 基于 QMT (国金证券) 的自建量化系统。

不做 QMT 官方工具的"二次开发"，而是从零搭建一套**自己的架构**，把 QMT 当作一个可替换的数据/交易驱动来用。

## 设计理念

```
研究阶段 (macOS)                     实盘阶段 (Windows)
──────────────                      ──────────────
akshare → SQLite    ──迁移──→   xtdata → SQLite
本地回测/因子研发                 xttrader 实盘下单
Go CLI 调度                     Go CLI 调度 (同代码)
```

两个阶段共用同一套 Go + Python 代码，只换数据源配置。

## 整体架构

```
┌──────────────────────────────────────────────────┐
│                    Go CLI (qt)                     │
│  入口 · 命令分派 · 定时调度 · 数据库读               │
└─────────────┬────────────────────────────────────┘
              │  spawn subprocess
              ▼
┌──────────────────────────────────────────────────┐
│                 Python (量化引擎)                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │  data/   │ │ factors/ │ │ strategy/        │  │
│  │ 数据下载  │ │ 因子计算  │ │ 策略 · 信号生成    │  │
│  └──────────┘ └──────────┘ └──────────────────┘  │
│  ┌──────────────────────────────────────────┐    │
│  │              backtest/  回测引擎           │    │
│  └──────────────────────────────────────────┘    │
└─────────────┬────────────────────────────────────┘
              │  read / write
              ▼
┌──────────────────────────────────────────────────┐
│              SQLite (data/quantify.db)             │
│  行情数据 · 信号 · 持仓 · 订单 (未来)                │
└──────────────────────────────────────────────────┘
```

### 数据流

```
config/default.yaml
       │
       ▼
  ┌─────────┐  akshare / xtdata   ┌──────────┐
  │ qt      │ ─────────────────→  │ Python   │
  │ download│ ←─── SQLite ─────── │ download │
  └─────────┘                     └──────────┘
       │
       ▼
  ┌─────────┐
  │ qt      │ ←── 直接读 SQLite
  │  list   │
  └─────────┘
```

### 职责划分

| 层 | 语言 | 职责 |
|----|------|------|
| CLI / 调度 | Go | 命令入口、子进程管理、SQLite 查询、定时任务 |
| 策略 / 数据 | Python | 行情下载、因子计算、回测、信号生成 |
| 存储 | SQLite | 所有持久化数据，Go 和 Python 共享读写 |

## 快速开始

### 环境要求

- Go 1.21+
- [uv](https://github.com/astral-sh/uv) (Python 包管理器)
- Python 3.12 (通过 uv 自动安装)

### 安装

```bash
# 安装 Python 依赖
make python-install

# 编译 Go CLI
make build

# 初始化项目（建目录 + 数据库建表）
make init
```

### 下载数据

```bash
# 下载配置文件中全部标的
make download

# 下载单只标的
./bin/qt download -s 600519.SH -f 2024-01-01
```

### 查看数据

```bash
make list

# 输出:
# Database:     /path/to/data/quantify.db
# Symbols:      1
# Total Rows:   588
# Date Range:   2024-01-02 ~ 2026-06-10
```

## 项目结构

```
quantify/
├── cmd/qt/main.go              # Go CLI 入口
├── internal/
│   ├── cmd/
│   │   ├── root.go             # 命令注册
│   │   ├── init.go             # qt init
│   │   ├── download.go         # qt download
│   │   └── list.go             # qt list
│   ├── config/config.go        # YAML 配置解析
│   └── db/sqlite.go            # SQLite 封装 (建表/统计)
├── python/
│   ├── pyproject.toml          # uv 项目定义
│   ├── quantify/
│   │   ├── config.py           # Python 侧配置加载
│   │   ├── data/
│   │   │   ├── schema.py       # 建表 DDL
│   │   │   └── downloader.py   # 数据下载 (akshare)
│   │   ├── factors/            # 因子计算
│   │   ├── strategy/           # 策略基类
│   │   └── backtest/           # 回测引擎
│   └── strategies/             # 具体策略
├── config/default.yaml         # 全局配置
├── data/quantify.db            # SQLite 数据库 (gitignore)
├── logs/                       # 日志 (gitignore)
├── Makefile
└── .gitignore
```

## 配置

```yaml
# config/default.yaml
db:
  path: ./data/quantify.db

data:
  source: akshare          # 数据源: akshare | xtdata
  start_date: "2015-01-01"
  symbols:
    - "000001.SZ"
    - "600519.SH"
```

## 技术栈

- **Go**: CLI、调度、SQLite 操作
- **Python 3.12**: 数据下载、因子计算、回测、策略
- **SQLite**: 本地持久化存储
- **akshare**: 公开行情数据源（研究阶段）
- **QMT/xtquant**: 券商行情+交易（实盘阶段，需 Windows）

## 阶段规划

| 阶段 | 状态 | 内容 |
|------|------|------|
| 一 | 已完成 | 项目骨架、Go CLI、数据下载入库 |
| 二 | 规划中 | 回测引擎、因子框架 |
| 三 | 规划中 | 策略研发、信号生成 |
| 四 | 未来 | 接入 QMT 实盘、风控、订单管理 |

## License

MIT
