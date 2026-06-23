package window

import "strings"

type Window struct {
	Handle  uintptr
	ID      string
	PID     uint32
	Process string
	Class   string
	Title   string
	Visible bool
	Backend string
}

func MatchByProcess(windows []Window, process string) []Window {
	process = strings.TrimSpace(process)
	if process == "" {
		return append([]Window(nil), windows...)
	}

	matches := make([]Window, 0, len(windows))
	for _, win := range windows {
		if strings.EqualFold(win.Process, process) || strings.EqualFold(win.Class, process) {
			matches = append(matches, win)
		}
	}

	return matches
}
