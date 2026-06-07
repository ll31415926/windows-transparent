//go:build linux

package window

import (
	"encoding/json"
	"fmt"
)

type swayNode struct {
	ID       int64      `json:"id"`
	PID      uint32     `json:"pid"`
	Name     string     `json:"name"`
	AppID    string     `json:"app_id"`
	Type     string     `json:"type"`
	Nodes    []swayNode `json:"nodes"`
	Floating []swayNode `json:"floating_nodes"`
	Props    struct {
		Class string `json:"class"`
	} `json:"window_properties"`
}

func listSwayWindows(r runner) ([]Window, error) {
	if err := requireTool(r, "swaymsg"); err != nil {
		return nil, err
	}

	output, err := r.Run("swaymsg", "-t", "get_tree")
	if err != nil {
		return nil, err
	}

	var root swayNode
	if err := json.Unmarshal([]byte(output), &root); err != nil {
		return nil, fmt.Errorf("parse sway tree: %w", err)
	}

	windows := make([]Window, 0)
	walkSway(root, &windows)
	return windows, nil
}

func setSwayOpacity(r runner, win Window, value int) error {
	if err := requireTool(r, "swaymsg"); err != nil {
		return err
	}

	id := linuxWindowID(win)
	if id == "" {
		return fmt.Errorf("missing Sway container id for %s", win.Process)
	}

	_, err := r.Run("swaymsg", fmt.Sprintf("[con_id=%s]", id), "opacity", "set", opacityFloat(value))
	return err
}

func walkSway(node swayNode, windows *[]Window) {
	if node.Type == "con" && (node.AppID != "" || node.Props.Class != "" || node.PID != 0) {
		class := node.AppID
		if class == "" {
			class = node.Props.Class
		}

		*windows = append(*windows, Window{
			Handle:  uintptr(node.ID),
			ID:      fmt.Sprintf("%d", node.ID),
			PID:     node.PID,
			Process: processName(node.PID),
			Class:   class,
			Title:   node.Name,
			Visible: true,
			Backend: backendSway,
		})
	}

	for _, child := range node.Nodes {
		walkSway(child, windows)
	}
	for _, child := range node.Floating {
		walkSway(child, windows)
	}
}
