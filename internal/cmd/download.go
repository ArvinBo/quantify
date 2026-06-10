package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"quantify/internal/config"
)

func Download(configPath, code, from string, all bool) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	base, _ := os.Getwd()
	pythonDir := filepath.Join(base, "python")

	exe := filepath.Join(pythonDir, ".venv", "bin", "python")
	if _, err := os.Stat(exe); os.IsNotExist(err) {
		exe = "python3"
		if _, err := exec.LookPath("python3"); err != nil {
			exe = "python"
		}
	}

	scriptArgs := []string{"-m", "quantify.data.downloader", "--db", cfg.DB.Path}

	if all || code == "" {
		scriptArgs = append(scriptArgs, "--all")
	} else {
		scriptArgs = append(scriptArgs, "--code", code)
	}

	if from != "" {
		scriptArgs = append(scriptArgs, "--from", from)
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
}
