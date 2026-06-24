//go:build linux

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func launchPersistentWatcher(configPath string) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("find current executable: %w", err)
	}

	cmd := exec.Command(exe, "watch", "--config", configPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	pid := cmd.Process.Pid
	return pid, cmd.Process.Release()
}

func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	return process.Signal(syscall.Signal(0)) == nil
}

func killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Kill()
}
