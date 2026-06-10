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
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		runInit(os.Args[2:])
	case "download":
		runDownload(os.Args[2:])
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
	all := fs.Bool("a", false, "download all symbols")
	fs.Parse(args)

	if err := cmd.Download(*configPath, *code, *from, *all); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
