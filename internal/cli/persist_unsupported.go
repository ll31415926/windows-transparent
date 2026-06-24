//go:build !windows && !linux

package cli

import "windows-transparent/internal/window"

func launchPersistentWatcher(_ string) (int, error) {
	return 0, window.ErrUnsupported
}

func processRunning(_ int) bool {
	return false
}

func killProcess(_ int) error {
	return window.ErrUnsupported
}
