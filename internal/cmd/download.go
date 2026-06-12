package cmd

import (
	"fmt"

	"quantify/internal/config"
)

// Download 下载行情数据：spawn Python 下载器，数据源由 config 的 data.source 决定。
func Download(configPath, code, from string, all, syncInfo bool) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	args := []string{"--db", cfg.DB.Path}

	if syncInfo {
		args = append(args, "--sync-info")
	}

	if all || code == "" {
		args = append(args, "--all")
	} else {
		args = append(args, "--code", code)
	}

	if from != "" {
		args = append(args, "--from", from)
	}

	fmt.Printf("downloading data (source=%s)...\n", cfg.Data.Source)
	if err := runPython("quantify.data.downloader", args...); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	return nil
}
