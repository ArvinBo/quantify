package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"quantify/internal/config"
)

var (
	configPath string
	cfg        *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "qt",
	Short: "quantify - quantitative trading research platform",
	Long:  `qt is a CLI tool for managing quantitative trading research workflows.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config/default.yaml", "config file path")
	rootCmd.AddCommand(initCmd, downloadCmd, listCmd)
}

func loadConfig() error {
	var err error
	cfg, err = config.LoadConfig(configPath)
	return err
}

func projectDir() string {
	dir, _ := os.Getwd()
	return dir
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
