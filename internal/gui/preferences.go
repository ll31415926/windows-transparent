//go:build desktop

package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type closePreferences struct {
	Action   string `json:"action"`   // "exit" or "tray"
	Remember bool   `json:"remember"` // whether to skip the dialog
}

func preferencesPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "wtrans")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "preferences.json"), nil
}

func loadClosePreference() closePreferences {
	path, err := preferencesPath()
	if err != nil {
		return closePreferences{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return closePreferences{}
	}
	var prefs closePreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return closePreferences{}
	}
	return prefs
}

func saveClosePreference(prefs closePreferences) error {
	path, err := preferencesPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
