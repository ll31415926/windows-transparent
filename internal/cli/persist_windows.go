//go:build windows

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const (
	windowsCreateNoWindow = 0x08000000
	windowsDetached       = 0x00000008

	windowsProcessQueryLimitedInformation = 0x1000
	windowsStillActive                    = 259
)

func launchPersistentWatcher(configPath string) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("find current executable: %w", err)
	}

	cmd := exec.Command(exe, "watch", "--config", configPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windowsCreateNoWindow | windowsDetached,
		HideWindow:    true,
	}

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

	handle, err := syscall.OpenProcess(windowsProcessQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	defer func() {
		_ = syscall.CloseHandle(handle)
	}()

	var code uint32
	if err := syscall.GetExitCodeProcess(handle, &code); err != nil {
		return false
	}

	return code == windowsStillActive
}

func killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Kill()
}
