//go:build windows

package window

import (
	"errors"
	"fmt"
	"io"
	"syscall"
	"unsafe"

	"windows-transparent/internal/opacity"
)

const (
	wsExLayered = 0x00080000

	lwaColorKey = 0x00000001
	lwaAlpha    = 0x00000002
)

var (
	gwlExStyle = ^uintptr(19)

	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procEnumWindows                = user32.NewProc("EnumWindows")
	procGetLayeredWindowAttributes = user32.NewProc("GetLayeredWindowAttributes")
	procGetWindowLongPtrW          = user32.NewProc("GetWindowLongPtrW")
	procGetWindowTextLengthW       = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW             = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessID   = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible            = user32.NewProc("IsWindowVisible")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procSetLastError               = kernel32.NewProc("SetLastError")
	procSetWindowLongPtrW          = user32.NewProc("SetWindowLongPtrW")
)

type layeredAttributes struct {
	colorKey uint32
	alpha    byte
	flags    uint32
}

func ListVisible() ([]Window, error) {
	processes, err := processNames()
	if err != nil {
		return nil, err
	}

	found := make([]Window, 0, 32)
	callback := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		if !isWindowVisible(hwnd) {
			return 1
		}

		pid := windowProcessID(hwnd)
		found = append(found, Window{
			Handle:  hwnd,
			ID:      fmt.Sprintf("0x%X", hwnd),
			PID:     pid,
			Process: processes[pid],
			Title:   windowText(hwnd),
			Visible: true,
			Backend: "windows",
		})

		return 1
	})

	ret, _, callErr := procEnumWindows.Call(callback, 0)
	if ret == 0 {
		if err := cleanLastError(callErr); err != nil {
			return nil, fmt.Errorf("enumerate windows: %w", err)
		}
	}

	return found, nil
}

func SetOpacity(win Window, value int) error {
	if err := opacity.Validate(value); err != nil {
		return err
	}

	hwnd := win.Handle
	style, err := extendedStyle(hwnd)
	if err != nil {
		return fmt.Errorf("get extended style for HWND 0x%X: %w", hwnd, err)
	}

	if style&wsExLayered == 0 {
		if err := setExtendedStyle(hwnd, style|wsExLayered); err != nil {
			return fmt.Errorf("enable layered style for HWND 0x%X: %w", hwnd, err)
		}
	}

	if err := setLayeredAttributes(hwnd, 0, opacity.ToAlpha(value), lwaAlpha); err != nil {
		return fmt.Errorf("set opacity for HWND 0x%X: %w", hwnd, err)
	}

	return nil
}

func Restore(win Window) error {
	hwnd := win.Handle
	style, err := extendedStyle(hwnd)
	if err != nil {
		return fmt.Errorf("get extended style for HWND 0x%X: %w", hwnd, err)
	}

	if style&wsExLayered == 0 {
		return nil
	}

	attrs, attrsErr := getLayeredAttributes(hwnd)
	colorKey := uint32(0)
	flags := uint32(lwaAlpha)
	removeLayeredStyle := false

	if attrsErr == nil {
		colorKey = attrs.colorKey
		flags = attrs.flags | lwaAlpha
		removeLayeredStyle = attrs.flags == 0 || attrs.flags == lwaAlpha
	}

	if err := setLayeredAttributes(hwnd, colorKey, 255, flags); err != nil {
		return fmt.Errorf("restore opacity for HWND 0x%X: %w", hwnd, err)
	}

	if removeLayeredStyle {
		if err := setExtendedStyle(hwnd, style&^uintptr(wsExLayered)); err != nil {
			return fmt.Errorf("remove layered style for HWND 0x%X: %w", hwnd, err)
		}
	}

	return nil
}

func Diagnose(w io.Writer) error {
	fmt.Fprintln(w, "backend: windows")
	fmt.Fprintln(w, "support: Win32 layered windows")
	return nil
}

func InstallGNOMEExtension(_ io.Writer) error {
	return errors.New("GNOME Shell extension installation is only supported on Linux")
}

func GNOMEExtensionStatus(_ io.Writer) error {
	return errors.New("GNOME Shell extension status is only supported on Linux")
}

func processNames() (map[uint32]string, error) {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("create process snapshot: %w", err)
	}
	defer func() {
		_ = syscall.CloseHandle(snapshot)
	}()

	processes := make(map[uint32]string)
	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := syscall.Process32First(snapshot, &entry); err != nil {
		if errors.Is(err, syscall.ERROR_NO_MORE_FILES) {
			return processes, nil
		}

		return nil, fmt.Errorf("read first process entry: %w", err)
	}

	for {
		processes[entry.ProcessID] = syscall.UTF16ToString(entry.ExeFile[:])

		entry.Size = uint32(unsafe.Sizeof(entry))
		if err := syscall.Process32Next(snapshot, &entry); err != nil {
			if errors.Is(err, syscall.ERROR_NO_MORE_FILES) {
				break
			}

			return nil, fmt.Errorf("read next process entry: %w", err)
		}
	}

	return processes, nil
}

func isWindowVisible(hwnd uintptr) bool {
	ret, _, _ := procIsWindowVisible.Call(hwnd)
	return ret != 0
}

func windowProcessID(hwnd uintptr) uint32 {
	var pid uint32
	procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	return pid
}

func windowText(hwnd uintptr) string {
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}

	buffer := make([]uint16, int(length)+1)
	ret, _, _ := procGetWindowTextW.Call(
		hwnd,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
	)
	if ret == 0 {
		return ""
	}

	return syscall.UTF16ToString(buffer[:ret])
}

func extendedStyle(hwnd uintptr) (uintptr, error) {
	procSetLastError.Call(0)
	ret, _, callErr := procGetWindowLongPtrW.Call(hwnd, gwlExStyle)
	if ret == 0 {
		if err := cleanLastError(callErr); err != nil {
			return 0, err
		}
	}

	return ret, nil
}

func setExtendedStyle(hwnd uintptr, style uintptr) error {
	procSetLastError.Call(0)
	ret, _, callErr := procSetWindowLongPtrW.Call(hwnd, gwlExStyle, style)
	if ret == 0 {
		if err := cleanLastError(callErr); err != nil {
			return err
		}
	}

	return nil
}

func getLayeredAttributes(hwnd uintptr) (layeredAttributes, error) {
	var attrs layeredAttributes
	ret, _, callErr := procGetLayeredWindowAttributes.Call(
		hwnd,
		uintptr(unsafe.Pointer(&attrs.colorKey)),
		uintptr(unsafe.Pointer(&attrs.alpha)),
		uintptr(unsafe.Pointer(&attrs.flags)),
	)
	if ret == 0 {
		return layeredAttributes{}, failedCallError(callErr)
	}

	return attrs, nil
}

func setLayeredAttributes(hwnd uintptr, colorKey uint32, alpha byte, flags uint32) error {
	ret, _, callErr := procSetLayeredWindowAttributes.Call(
		hwnd,
		uintptr(colorKey),
		uintptr(alpha),
		uintptr(flags),
	)
	if ret == 0 {
		return failedCallError(callErr)
	}

	return nil
}

func failedCallError(err error) error {
	if clean := cleanLastError(err); clean != nil {
		return clean
	}

	return errors.New("win32 call failed without an error code")
}

func cleanLastError(err error) error {
	if err == nil {
		return nil
	}

	var errno syscall.Errno
	if errors.As(err, &errno) && errno == 0 {
		return nil
	}

	return err
}
