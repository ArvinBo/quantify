package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	downloadCode string
	downloadFrom string
	downloadAll  bool
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download daily kline data",
	Long:  `Download A-share daily kline data using the active data source (akshare / xtdata).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		base := projectDir()
		pythonDir := filepath.Join(base, "python")

		exe := filepath.Join(pythonDir, ".venv", "bin", "python")
		if _, err := os.Stat(exe); os.IsNotExist(err) {
			exe = "python3"
			if _, err := exec.LookPath("python3"); err != nil {
				exe = "python"
			}
		}

		scriptArgs := []string{"-m", "quantify.data.downloader", "--db", cfg.DB.Path}

		if downloadAll {
			scriptArgs = append(scriptArgs, "--all")
		} else if downloadCode != "" {
			scriptArgs = append(scriptArgs, "--code", downloadCode)
		} else {
			scriptArgs = append(scriptArgs, "--all")
		}

		if downloadFrom != "" {
			scriptArgs = append(scriptArgs, "--from", downloadFrom)
		}

		fmt.Printf("downloading data via %s...\n", exe)
		c := exec.Command(exe, scriptArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Dir = pythonDir

		if err := c.Run(); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVarP(&downloadCode, "code", "s", "", "single stock code (e.g. 600519.SH)")
	downloadCmd.Flags().StringVarP(&downloadFrom, "from", "f", "", "start date (e.g. 2024-01-01)")
	downloadCmd.Flags().BoolVarP(&downloadAll, "all", "a", false, "download all symbols from config")
}
