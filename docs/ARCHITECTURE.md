# 架构设计

## 核心理念

**Go 管流程和安全，Python 管计算和策略。**

| | Go | Python |
|---|---|---|
| 角色 | 调度者 · 守门人 | 执行者 · 研究员 |
| 能做什么 | 启动任务、读数据、做风控、下单 | 下载数据、算因子、跑回测、生成信号 |
| 不能做什么 | 不算因子、不跑策略逻辑 | 不直接下单、不接触风控 |
| 实盘阶段 | **唯一的交易出口** | 只产出信号 |

---

## 两阶段架构

### 研究阶段（当前）

```
                    ┌─────────────────┐
                    │  用户敲 qt 命令   │
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
        qt download                   qt backtest
              │                        (规划中)
              │                             │
     ┌────────┴────────┐           ┌────────┴────────┐
     │  Go spawn       │           │  Go spawn       │
     │  Python 子进程   │           │  Python 子进程   │
     └────────┬────────┘           └────────┬────────┘
              │                             │
              ▼                             ▼
     ┌────────────────┐           ┌────────────────┐
     │  Python         │           │  Python         │
     │  akshare 下载    │           │  SQLite 取行情   │
     │  → 写入 SQLite   │           │  → 策略算信号    │
     │                  │           │  → 模拟撮合      │
     │                  │           │  → 输出指标      │
     └────────┬─────────┘           └────────────────┘
              │                             │
              ▼                             │
     ┌────────────────────────────┐         │
     │      data/quantify.db       │◄────────┘
     │                            │
     │   daily_kline_tab (行情)    │
     │   backtest_results (回测)   │
     └────────────┬───────────────┘
                  │
     ┌────────────┴───────────────┐
     │  Go 读 DB (仅查询/统计)     │
     │  qt list → GetStats()      │
     └────────────────────────────┘
```

Go 现在只读表做展示，Python 承担全部量化工作。

---

### 实盘阶段（未来）

```
     ┌─────────────────────┐
     │  Python (定时/事件)       │
     │  因子计算 → 产生信号   │
     └──────────┬──────────┘
                │
       INSERT INTO signals
                │
                ▼
     ┌─────────────────────┐
     │  Go (定时轮询)          │
     │                       │
     │  读 signals 表        │──── Go 读表的核心意义
     │    ↓                  │
     │  风控检查              │     Python 不能直接下单
     │  · 单票仓位超限？      │     Go 是唯一的下单通道
     │  · 日亏损超限？        │
     │  · 可用资金够吗？      │
     │    ↓                  │
     │  xttrader 下单        │
     └───────────────────────┘
```

Python 只负责"建议买入"，Go 决定"是否执行"。策略代码没有能力亏钱。

---

## Go ↔ Python 协作机制

Go 和 Python 是两个独立的进程，不通过 HTTP、gRPC、管道等方式通信。它们通过三种方式协作，**这三种方式不是互斥的，一次 `qt download --all` 就把三者全走了一遍**。

### 协作全景

以 `qt download --all` 为例：

```
用户: qt download --all

  ① 方式三: 共享 YAML
  ┌──────────────────────────────────┐
  │ Go:  config.LoadConfig()         │  读 config/default.yaml
  │      ．拿 db.path → ./data/quantify.db    │
  │      ．拼成绝对路径传给 Python            │
  │                                  │
  │ Py:  load_config()               │  读同一份 config/default.yaml
  │      ．拿 data.symbols → ['000001.SZ','600519.SH']
  │      ．拿 data.start_date → '2015-01-01'
  └──────────────────────────────────┘
         │                              │
         ▼                              │
  ② 方式一: spawn 子进程                  │
  ┌──────────────────────────────────┐  │
  │ Go:  exec.Command(               │  │
  │        ".venv/bin/python",        │  │
  │        "-m",                      │  │
  │        "quantify.data.downloader", │  │
  │        "--db", dbPath,     ←──────┘  (方式三拿到的路径)
  │        "--all"                     │
  │      )                            │
  │      等待 Python 跑完 ...           │
  └──────────┬───────────────────────┘
             │
             ▼
  ┌──────────────────────────────────┐
  │ Python: akshare 下载 600519.SH    │
  │         → pandas 清洗             │
  │         → INSERT INTO daily_kline_tab
  └──────────┬───────────────────────┘
             │
             ▼
  ③ 方式二: 共享 SQLite
  ┌──────────────────────────────────┐
  │ Python 刚写完 → data/quantify.db  │
  │                 ↑                │
  │ Go 接着读 ←─────┘                │
  │ DailyKlineRepo.GetStats()        │
  │ → "Symbols: 2, Rows: 1176"       │
  └──────────────────────────────────┘
```

---

### 方式一：Go spawn Python 子进程

**做什么**：Go 启动一个 Python 进程，给它传参数，等它跑完。

**什么时候用**：任何需要 Python 干活的命令 —— 下载数据、跑回测、跑策略。

**谁触发谁**：Go → Python（单向，Go 是老板，Python 是打工人）。

**实现文件**：`internal/cmd/download.go`

```go
c := exec.Command(venvPython, "-m", "quantify.data.downloader",
    "--db", dbPath, "--all")
c.Dir = "python/"          // 让 Python 找到 quantify 包
c.Stdout = os.Stdout        // Python 的 print → 终端
c.Run()                     // 阻塞，等 Python 退出
```

**局限**：这是"一次性"的协作。Python 跑完就退出，Go 不能中途和它对话。适合离线任务（下载、回测），不适合实时交互。

---

### 方式二：共享 SQLite 文件

**做什么**：Go 和 Python 读写同一个 `data/quantify.db` 文件，通过数据库交换数据。

**什么时候用**：

- **当下**：Python 下载完写表，Go 用 `list` 命令读统计。
- **将来**：Python 把信号写入 `signals` 表，Go 读出来做风控下单。Python 回测结果写入 `backtest_results` 表，Go 读出来展示。

**谁触发谁**：无主次。双方平等读写，不分先后。

**实现文件**：

| Go | Python |
|----|--------|
| `db/daily_kline_repo.go` （GORM） | `data/downloader.py` （sqlite3） |
| `model/dbmodel/daily_kline.go` （映射） | `data/schema.py` （建表 DDL） |

**为什么不用 HTTP/gRPC**：项目跑在本地单机，SQLite 文件就是最快的 IPC。不加网络层，零配置，零延迟。

---

### 方式三：共享 YAML 配置文件

**做什么**：Go 和 Python 各自读取同一份 `config/default.yaml`，各取自己需要的字段。

**什么时候用**：**每次执行任何命令都在用**。它不是可选的 —— 两边都需要从配置中拿信息才能工作。

**Go 拿什么**：
```go
cfg, _ := config.LoadConfig("config/default.yaml")
cfg.DB.Path    // "./data/quantify.db"  → 传给 Python 子进程
```
Go 只关心 `db.path`，因为它要告诉 Python "数据写到哪儿"。

**Python 拿什么**：
```python
cfg = load_config()
cfg["data"]["symbols"]    # ["000001.SZ", "600519.SH"]
cfg["data"]["start_date"] # "2015-01-01"
```
Python 关心 `symbols` 和 `start_date`，因为它要决定"下载哪些标的、从哪天开始"。

**为什么两边各读各的，不通过 Go 传参**：

- 配置字段会越来越多（数据源、手续费率、风控阈值...），全用命令行参数传太臃肿
- 让 Python 自己读配置，加新字段时不用改 Go 的传参逻辑
- Go 传给 Python 的只有**关键变量**（如 `--db`），其余 Python 自己从 YAML 拿

**路径解析差异**：

| | Go | Python |
|---|---|---|
| 基准 | `os.Getwd()`（项目根） | `ROOT_DIR`（`__file__` 上溯 3 层） |
| 原因 | Go 从项目根启动 | Python 被 Go 在 `python/` 子目录下启动 |
| `./data/quantify.db` 解析为 | `/project/data/quantify.db` | `/project/data/quantify.db` |
| **结果一致** | 殊途同归 |

---

### 三种方式对比

| | 方式一 spawn | 方式二 SQLite | 方式三 YAML |
|---|---|---|---|
| 方向 | Go → Python（单向） | 双向读写 | 各自独立读 |
| 时效 | 一次性（任务跑完结束） | 持久（数据一直存在） | 启动时加载 |
| Go 做的事 | 启动 + 等结果 | 读统计 / 读信号 | 拿 db.path |
| Python 做的事 | 下载 / 回测 / 算因子 | 写数据 / 读行情 | 拿 symbols / 参数 |
| 当下用到 | `qt download` | `qt list` | 所有命令 |
| 将来用到 | `qt backtest`, `qt run` | 信号-下单链路 | 所有命令 |

---

## 技术选型理由

| 组件 | 选型 | 理由 |
|------|------|------|
| CLI 框架 | Go stdlib `flag` | 命令少，够用，零依赖 |
| 数据库 | SQLite + GORM | 零配置部署，Repository 模式可切换 |
| Python 包管理 | uv | 极快，自带 venv 和 lock |
| 行情源（研究） | akshare | 免费，覆盖 A 股日线 |
| 行情+交易（实盘） | QMT xtquant | 券商提供，仅 Windows |
| 日志 | loguru | Python 侧开箱即用 |
