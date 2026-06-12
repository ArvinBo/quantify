package cmd

import (
	"fmt"

	"quantify/internal/config"
	"quantify/internal/db"
	"quantify/internal/model/dbmodel"
	"quantify/internal/risk"
	"quantify/internal/trader"
)

// Run 实盘流程（单次执行，可由 cron 定时调度）：
//  1. spawn Python 生成信号 → 写 signal_tab
//  2. Go 读取 PENDING 信号 → 风控检查 → 下单通道 → 更新订单/持仓/信号状态
//
// Python 只负责"建议买卖"，Go 是唯一的下单出口。
func Run(configPath, strategy string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	fmt.Println("step 1: generating signals via Python...")
	if err := runPython("quantify.live.signal_gen", "--db", cfg.DB.Path, "--strategy", strategy); err != nil {
		return fmt.Errorf("signal generation failed: %w", err)
	}

	fmt.Println("step 2: processing pending signals...")
	database, err := db.Open(cfg.DB.Path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close(database)

	signalRepo := db.NewSignalRepo(database)
	orderRepo := db.NewOrderRepo(database)
	positionRepo := db.NewPositionRepo(database)
	checker := risk.NewChecker(cfg.Risk, positionRepo)

	tr, err := trader.New(cfg.Trade.Mode)
	if err != nil {
		return fmt.Errorf("init trader: %w", err)
	}

	signals, err := signalRepo.FindPending()
	if err != nil {
		return fmt.Errorf("find pending signals: %w", err)
	}
	if len(signals) == 0 {
		fmt.Println("no pending signals")
		return nil
	}

	executed, rejected := 0, 0
	for _, sig := range signals {
		if err := processSignal(sig, cfg, checker, tr, signalRepo, orderRepo, positionRepo); err != nil {
			rejected++
			fmt.Printf("signal #%d %s %s rejected: %v\n", sig.ID, sig.Action, sig.Code, err)
			if uerr := signalRepo.UpdateStatus(sig.ID, dbmodel.SignalStatusRejected, err.Error()); uerr != nil {
				return fmt.Errorf("update signal status: %w", uerr)
			}
			continue
		}
		executed++
		if uerr := signalRepo.UpdateStatus(sig.ID, dbmodel.SignalStatusExecuted, ""); uerr != nil {
			return fmt.Errorf("update signal status: %w", uerr)
		}
	}

	fmt.Printf("done: %d executed, %d rejected (mode=%s)\n", executed, rejected, cfg.Trade.Mode)
	return nil
}

// processSignal 处理单条信号：定量 → 风控 → 下单 → 记录订单与持仓。
func processSignal(
	sig *dbmodel.Signal,
	cfg *config.Config,
	checker *risk.Checker,
	tr trader.Trader,
	signalRepo db.SignalRepo,
	orderRepo db.OrderRepo,
	positionRepo db.PositionRepo,
) error {
	if sig.Price <= 0 {
		return fmt.Errorf("invalid signal price: %.2f", sig.Price)
	}

	var qty int
	switch sig.Action {
	case dbmodel.ActionBuy:
		// 单笔买入金额 = 账户资金 × 单票仓位上限，按 100 股取整
		budget := cfg.Risk.Capital * cfg.Risk.MaxPositionPct
		qty = int(budget/(sig.Price*100)) * 100
		if qty <= 0 {
			return fmt.Errorf("budget %.2f too small for price %.2f", budget, sig.Price)
		}
		if err := checker.CheckBuy(sig.Code, sig.Price, qty); err != nil {
			return err
		}
	case dbmodel.ActionSell:
		pos, err := positionRepo.Get(sig.Code)
		if err != nil {
			return err
		}
		if pos == nil || pos.Qty == 0 {
			return fmt.Errorf("no position to sell")
		}
		qty = pos.Qty // 全仓卖出
		if err := checker.CheckSell(sig.Code, qty); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown action: %s", sig.Action)
	}

	status, err := tr.Place(&trader.Order{
		SignalID: sig.ID,
		Code:     sig.Code,
		Action:   sig.Action,
		Price:    sig.Price,
		Qty:      qty,
	})
	if err != nil {
		return fmt.Errorf("place order: %w", err)
	}

	if err := orderRepo.Create(&dbmodel.Order{
		SignalID: sig.ID,
		Code:     sig.Code,
		Action:   sig.Action,
		Price:    sig.Price,
		Qty:      qty,
		Status:   status,
	}); err != nil {
		return fmt.Errorf("record order: %w", err)
	}

	if err := positionRepo.ApplyFill(sig.Code, sig.Action, sig.Price, qty); err != nil {
		return fmt.Errorf("update position: %w", err)
	}
	return nil
}
