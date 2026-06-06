package window

import "strings"

type Window struct {
	Handle  uintptr
	PID     uint32
	Process string
	Title   string
	Visible bool
}

func MatchByProcess(windows []Window, process string) []Window {
	process = strings.TrimSpace(process)
	if process == "" {
		return append([]Window(nil), windows...)
	}

	matches := make([]Window, 0, len(windows))
	for _, win := range windows {
		if strings.EqualFold(win.Process, process) {
			matches = append(matches, win)
		}
	}

	return matches
}
