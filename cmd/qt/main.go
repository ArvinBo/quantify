package main

import (
	"flag"
	"fmt"
	"os"

	"quantify/internal/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: qt <command> [options]")
		fmt.Println("commands:")
		fmt.Println("  init       initialize database connection")
		fmt.Println("  download   download daily kline data")
		fmt.Println("  backtest   run backtest for a strategy")
		fmt.Println("  run        generate signals and execute trades (risk-checked)")
		fmt.Println("  stats      show database stats and positions")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		runInit(os.Args[2:])
	case "download":
		runDownload(os.Args[2:])
	case "backtest":
		runBacktest(os.Args[2:])
	case "run":
		runLive(os.Args[2:])
	case "stats":
		runStats(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	configPath := fs.String("c", "config/default.yaml", "config file path")
	fs.Parse(args)

	if err := cmd.InitDB(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runDownload(args []string) {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	configPath := fs.String("c", "config/default.yaml", "config file path")
	code := fs.String("s", "", "stock code (e.g. 600519.SH)")
	from := fs.String("f", "", "start date (e.g. 2024-01-01)")
	var all bool
	fs.BoolVar(&all, "a", false, "download all symbols")
	fs.BoolVar(&all, "all", false, "download all symbols")
	var syncInfo bool
	fs.BoolVar(&syncInfo, "sync-info", false, "sync stock info")
	fs.Parse(args)

	if err := cmd.Download(*configPath, *code, *from, all, syncInfo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runBacktest(args []string) {
	fs := flag.NewFlagSet("backtest", flag.ExitOnError)
	configPath := fs.String("c", "config/default.yaml", "config file path")
	strategy := fs.String("strategy", "ma_cross", "strategy name (in python/strategies/)")
	code := fs.String("code", "", "stock code (e.g. 600519.SH)")
	start := fs.String("start", "", "start date (e.g. 2020-01-01)")
	end := fs.String("end", "", "end date (e.g. 2024-12-31)")
	fs.Parse(args)

	if *code == "" {
		fmt.Fprintln(os.Stderr, "backtest: --code is required")
		os.Exit(1)
	}

	if err := cmd.Backtest(*configPath, *strategy, *code, *start, *end); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runLive(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := fs.String("c", "config/default.yaml", "config file path")
	strategy := fs.String("strategy", "ma_cross", "strategy name (in python/strategies/)")
	fs.Parse(args)

	if err := cmd.Run(*configPath, *strategy); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	configPath := fs.String("c", "config/default.yaml", "config file path")
	fs.Parse(args)

	if err := cmd.Stats(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
