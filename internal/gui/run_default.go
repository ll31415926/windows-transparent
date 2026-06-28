//go:build !desktop

package gui

import "fmt"

func Run(_ Config) error {
	return fmt.Errorf(`GUI builds require Wails desktop build tags.

Use one of these commands:
  wails build
  go build -tags "desktop,production" -ldflags "-w -s -H windowsgui" -o wtrans-gui.exe ./cmd/wtrans-gui`)
}
