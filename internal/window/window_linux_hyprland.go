//go:build linux

package window

import (
	"encoding/json"
	"fmt"
)

type hyprlandClient struct {
	Address string `json:"address"`
	Class   string `json:"class"`
	Title   string `json:"title"`
	PID     uint32 `json:"pid"`
	Mapped  bool   `json:"mapped"`
}

func listHyprlandWindows(r runner) ([]Window, error) {
	if err := requireTool(r, "hyprctl"); err != nil {
		return nil, err
	}

	output, err := r.Run("hyprctl", "clients", "-j")
	if err != nil {
		return nil, err
	}

	var clients []hyprlandClient
	if err := json.Unmarshal([]byte(output), &clients); err != nil {
		return nil, fmt.Errorf("parse Hyprland clients: %w", err)
	}

	windows := make([]Window, 0, len(clients))
	for _, client := range clients {
		if !client.Mapped {
			continue
		}

		windows = append(windows, Window{
			ID:      client.Address,
			PID:     client.PID,
			Process: processName(client.PID),
			Class:   client.Class,
			Title:   client.Title,
			Visible: true,
			Backend: backendHyprland,
		})
	}

	return windows, nil
}

func setHyprlandOpacity(r runner, win Window, value int) error {
	if err := requireTool(r, "hyprctl"); err != nil {
		return err
	}

	id := linuxWindowID(win)
	if id == "" {
		return fmt.Errorf("missing Hyprland window address for %s", win.Process)
	}

	opacityValue := opacityFloat(value)
	selector := "address:" + id
	if _, err := r.Run("hyprctl", "setprop", selector, "alpha", opacityValue); err != nil {
		return err
	}
	if _, err := r.Run("hyprctl", "setprop", selector, "alphainactive", opacityValue); err != nil {
		return err
	}

	return nil
}
