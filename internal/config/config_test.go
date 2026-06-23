package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"windows-transparent/internal/window"
)

func TestParseValidConfig(t *testing.T) {
	cfg, err := Parse(strings.NewReader(`{
		"rules": [
			{ "process": "notepad.exe", "class": "Notepad", "title_contains": "Untitled", "opacity": 65, "enabled": true },
			{ "process": "Code.exe", "opacity": 85 }
		]
	}`))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(cfg.Rules) != 2 {
		t.Fatalf("len(cfg.Rules) = %d, want 2", len(cfg.Rules))
	}
	if cfg.Rules[0].Process != "notepad.exe" || cfg.Rules[0].Opacity != 65 {
		t.Fatalf("first rule = %#v", cfg.Rules[0])
	}
	if cfg.Rules[0].Class != "Notepad" || cfg.Rules[0].TitleContains != "Untitled" || !cfg.Rules[0].EnabledValue() {
		t.Fatalf("first rule selector = %#v", cfg.Rules[0])
	}
}

func TestParseRejectsInvalidOpacity(t *testing.T) {
	_, err := Parse(strings.NewReader(`{
		"rules": [
			{ "process": "notepad.exe", "opacity": 10 }
		]
	}`))
	if err == nil {
		t.Fatal("Parse returned nil error, want validation error")
	}
}

func TestParseRejectsMissingProcess(t *testing.T) {
	_, err := Parse(strings.NewReader(`{
		"rules": [
			{ "process": "", "opacity": 70 }
		]
	}`))
	if err == nil {
		t.Fatal("Parse returned nil error, want validation error")
	}
}

func TestParseRejectsTrailingJSON(t *testing.T) {
	_, err := Parse(strings.NewReader(`{"rules": []} {"rules": []}`))
	if err == nil {
		t.Fatal("Parse returned nil error, want trailing JSON error")
	}
}

func TestResolvePathTrimsExplicitPath(t *testing.T) {
	got, err := ResolvePath("  rules.json  ")
	if err != nil {
		t.Fatalf("ResolvePath returned error: %v", err)
	}

	if got != "rules.json" {
		t.Fatalf("ResolvePath = %q, want rules.json", got)
	}
}

func TestRuleMatchesProcessClassAndTitle(t *testing.T) {
	rule := Rule{
		Process:       "Code.exe",
		Class:         "Chrome_WidgetWin_1",
		TitleContains: "windows-transparent",
		Opacity:       75,
	}
	win := window.Window{
		Process: "code.exe",
		Class:   "chrome_widgetwin_1",
		Title:   "main.go - windows-transparent",
	}

	if !rule.Matches(win) {
		t.Fatalf("rule did not match window")
	}
}

func TestRuleMatchesDisabledRuleReturnsFalse(t *testing.T) {
	rule := Rule{Process: "notepad.exe", Opacity: 70, Enabled: Bool(false)}
	win := window.Window{Process: "notepad.exe"}

	if rule.Matches(win) {
		t.Fatalf("disabled rule matched window")
	}
}

func TestRuleMatchesRejectsDifferentTitle(t *testing.T) {
	rule := Rule{Process: "notepad.exe", TitleContains: "draft", Opacity: 70}
	win := window.Window{Process: "notepad.exe", Title: "Untitled - Notepad"}

	if rule.Matches(win) {
		t.Fatalf("rule matched unrelated title")
	}
}

func TestUpsertRuleUpdatesExistingSelector(t *testing.T) {
	cfg := Config{Rules: []Rule{{Process: "Code.exe", Class: "Editor", TitleContains: "repo", Opacity: 60}}}

	updated := cfg.UpsertRule(Rule{Process: "code.exe", Class: "editor", TitleContains: "REPO", Opacity: 85, Enabled: Bool(true)})

	if !updated {
		t.Fatalf("UpsertRule updated = false, want true")
	}
	if len(cfg.Rules) != 1 {
		t.Fatalf("len(cfg.Rules) = %d, want 1", len(cfg.Rules))
	}
	if cfg.Rules[0].Opacity != 85 || !cfg.Rules[0].EnabledValue() {
		t.Fatalf("updated rule = %#v", cfg.Rules[0])
	}
}

func TestSaveCreatesConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wtrans", "config.json")
	cfg := Config{
		Rules: []Rule{{Process: "notepad.exe", Opacity: 70}},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if !strings.Contains(string(data), `"process": "notepad.exe"`) {
		t.Fatalf("saved config did not include rule: %s", data)
	}
}
