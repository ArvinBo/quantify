package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
)

// runPython spawn 一个 Python 子进程执行 quantify 包内模块（协作方式一）。
// 工作目录设为 python/，使 quantify 与 strategies 包可被导入。
func runPython(module string, args ...string) error {
	base, _ := os.Getwd()
	pythonDir := filepath.Join(base, "python")

	exe := filepath.Join(pythonDir, ".venv", "bin", "python")
	if _, err := os.Stat(exe); os.IsNotExist(err) {
		exe = "python3"
		if _, err := exec.LookPath("python3"); err != nil {
			exe = "python"
		}
	}

	scriptArgs := append([]string{"-m", module}, args...)
	c := exec.Command(exe, scriptArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = pythonDir
	return c.Run()
}
