package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseValidConfig(t *testing.T) {
	cfg, err := Parse(strings.NewReader(`{
		"rules": [
			{ "process": "notepad.exe", "opacity": 65 },
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
