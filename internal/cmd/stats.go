package cmd

import (
	"fmt"

	"quantify/internal/config"
	"quantify/internal/db"
)

// Stats 打印数据库概览：行情统计与当前持仓（协作方式二：Go 直接读 SQLite）。
func Stats(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	database, err := db.Open(cfg.DB.Path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close(database)

	klineRepo := db.NewDailyKlineRepo(database)
	stats, err := klineRepo.GetStats()
	if err != nil {
		return fmt.Errorf("get kline stats: %w", err)
	}

	fmt.Println("===== kline data =====")
	fmt.Printf("symbols   : %d\n", stats.SymbolCount)
	fmt.Printf("total rows: %d\n", stats.TotalRows)
	fmt.Printf("date range: %s ~ %s\n", stats.MinDate, stats.MaxDate)

	positionRepo := db.NewPositionRepo(database)
	positions, err := positionRepo.FindAll()
	if err != nil {
		return fmt.Errorf("get positions: %w", err)
	}

	fmt.Println("===== positions =====")
	if len(positions) == 0 {
		fmt.Println("(empty)")
		return nil
	}
	for _, p := range positions {
		fmt.Printf("%s  qty=%d  avg_cost=%.2f  value=%.2f\n",
			p.Code, p.Qty, p.AvgCost, p.AvgCost*float64(p.Qty))
	}
	return nil
}
