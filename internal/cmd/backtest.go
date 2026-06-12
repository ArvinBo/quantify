package cmd

import (
	"fmt"

	"quantify/internal/config"
)

// Backtest 执行回测：spawn Python 回测引擎，结果打印到终端并写入 backtest_result_tab。
func Backtest(configPath, strategy, code, startDate, endDate string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	args := []string{
		"--db", cfg.DB.Path,
		"--strategy", strategy,
		"--code", code,
	}
	if startDate != "" {
		args = append(args, "--start", startDate)
	}
	if endDate != "" {
		args = append(args, "--end", endDate)
	}

	if err := runPython("quantify.backtest.run", args...); err != nil {
		return fmt.Errorf("backtest failed: %w", err)
	}
	return nil
}
