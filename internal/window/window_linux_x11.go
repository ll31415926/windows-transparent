//go:build linux

package window

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func listX11Windows(r runner) ([]Window, error) {
	if err := requireTool(r, "wmctrl"); err != nil {
		return nil, err
	}
	if err := requireTool(r, "xprop"); err != nil {
		return nil, err
	}

	output, err := r.Run("wmctrl", "-lp")
	if err != nil {
		return nil, err
	}

	windows := make([]Window, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		win, ok := parseWMCtrlLine(scanner.Text())
		if !ok {
			continue
		}

		class, _ := x11WindowClass(r, win.ID)
		win.Class = class
		if win.Process == "" {
			win.Process = processName(win.PID)
		}
		win.Backend = backendX11
		win.Visible = true
		windows = append(windows, win)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return windows, nil
}

func setX11Opacity(r runner, win Window, value int) error {
	if err := requireTool(r, "xprop"); err != nil {
		return err
	}

	id := linuxWindowID(win)
	if id == "" {
		return fmt.Errorf("missing X11 window id for %s", win.Process)
	}

	_, err := r.Run(
		"xprop",
		"-id", id,
		"-f", "_NET_WM_WINDOW_OPACITY", "32c",
		"-set", "_NET_WM_WINDOW_OPACITY", alphaHex(value),
	)
	return err
}

func restoreX11Opacity(r runner, win Window) error {
	if err := requireTool(r, "xprop"); err != nil {
		return err
	}

	id := linuxWindowID(win)
	if id == "" {
		return fmt.Errorf("missing X11 window id for %s", win.Process)
	}

	_, err := r.Run("xprop", "-id", id, "-remove", "_NET_WM_WINDOW_OPACITY")
	return err
}

func parseWMCtrlLine(line string) (Window, bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return Window{}, false
	}

	pid := uint32(0)
	if fields[2] != "-1" {
		pid64, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			return Window{}, false
		}
		pid = uint32(pid64)
	}

	title := ""
	if len(fields) > 4 {
		title = strings.Join(fields[4:], " ")
	}

	handle, _ := strconv.ParseUint(strings.TrimPrefix(fields[0], "0x"), 16, 64)
	return Window{
		Handle: uintptr(handle),
		ID:     fields[0],
		PID:    pid,
		Title:  title,
	}, true
}

func x11WindowClass(r runner, id string) (string, error) {
	output, err := r.Run("xprop", "-id", id, "WM_CLASS")
	if err != nil {
		return "", err
	}

	return parseWMClass(output), nil
}

func parseWMClass(output string) string {
	parts := strings.Split(output, "=")
	if len(parts) < 2 {
		return ""
	}

	values := strings.Split(parts[1], ",")
	if len(values) == 0 {
		return ""
	}

	return strings.Trim(strings.TrimSpace(values[len(values)-1]), `"`)
}

func processName(pid uint32) string {
	if pid == 0 {
		return ""
	}

	commPath := filepath.Join("/proc", strconv.FormatUint(uint64(pid), 10), "comm")
	data, err := os.ReadFile(commPath)
	if err == nil {
		name := strings.TrimSpace(string(data))
		if name != "" {
			return name
		}
	}

	cmdlinePath := filepath.Join("/proc", strconv.FormatUint(uint64(pid), 10), "cmdline")
	data, err = os.ReadFile(cmdlinePath)
	if err != nil {
		return ""
	}

	parts := strings.Split(string(data), "\x00")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}

	return filepath.Base(parts[0])
}

func linuxWindowID(win Window) string {
	if win.ID != "" {
		return win.ID
	}
	if win.Handle != 0 {
		return fmt.Sprintf("0x%X", win.Handle)
	}

	return ""
}
