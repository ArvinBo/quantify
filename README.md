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
├── internal/         Go 内部包（cmd / config / db / model）
├── python/           Python 量化包（data / factors / strategy / backtest）
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
  source: akshare               # 研究阶段
  start_date: "2015-01-01"
  symbols:
    - "000001.SZ"
    - "600519.SH"
```

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
| 一 | 已完成 | 项目骨架、Go CLI、akshare 数据下载 |
| 二 | 待开始 | 回测引擎、策略示例 |
| 三 | 规划中 | 因子框架、参数优化 |
| 四 | 未来 | QMT 实盘、风控、下单 |
