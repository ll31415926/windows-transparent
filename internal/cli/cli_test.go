package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"windows-transparent/internal/config"
	"windows-transparent/internal/window"
)

func TestParseSetCommand(t *testing.T) {
	cmd, err := Parse([]string{"set", "--process", "notepad.exe", "--opacity", "70"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandSet {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandSet)
	}
	if cmd.Process != "notepad.exe" {
		t.Fatalf("cmd.Process = %q, want notepad.exe", cmd.Process)
	}
	if cmd.Opacity != 70 {
		t.Fatalf("cmd.Opacity = %d, want 70", cmd.Opacity)
	}
}

func TestParseHelpCommand(t *testing.T) {
	cmd, err := Parse([]string{"--help"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandHelp {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandHelp)
	}
}

func TestParseVersionCommand(t *testing.T) {
	cmd, err := Parse([]string{"version"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandVersion {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandVersion)
	}
}

func TestParseListCommandWithProcess(t *testing.T) {
	cmd, err := Parse([]string{"list", "--process=Code.exe"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandList {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandList)
	}
	if cmd.Process != "Code.exe" {
		t.Fatalf("cmd.Process = %q, want Code.exe", cmd.Process)
	}
}

func TestParseApplyCommandWithConfig(t *testing.T) {
	cmd, err := Parse([]string{"apply", "--config", `C:\tmp\wtrans.json`})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandApply {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandApply)
	}
	if cmd.ConfigPath != `C:\tmp\wtrans.json` {
		t.Fatalf("cmd.ConfigPath = %q", cmd.ConfigPath)
	}
}

func TestParseWatchCommand(t *testing.T) {
	cmd, err := Parse([]string{"watch", "--config", "rules.json", "--interval", "500ms"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandWatch {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandWatch)
	}
	if cmd.ConfigPath != "rules.json" {
		t.Fatalf("cmd.ConfigPath = %q, want rules.json", cmd.ConfigPath)
	}
	if cmd.Interval != 500*time.Millisecond {
		t.Fatalf("cmd.Interval = %s, want 500ms", cmd.Interval)
	}
}

func TestParseRememberCommand(t *testing.T) {
	cmd, err := Parse([]string{
		"remember",
		"--process", "Code.exe",
		"--class", "Chrome_WidgetWin_1",
		"--title-contains", "windows-transparent",
		"--opacity", "75",
		"--config", "rules.json",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandRemember {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandRemember)
	}
	if cmd.Process != "Code.exe" || cmd.Class != "Chrome_WidgetWin_1" || cmd.Title != "windows-transparent" || cmd.Opacity != 75 || cmd.ConfigPath != "rules.json" {
		t.Fatalf("parsed command = %#v", cmd)
	}
}

func TestParseDiagnoseCommand(t *testing.T) {
	cmd, err := Parse([]string{"diagnose"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandDiagnose {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandDiagnose)
	}
}

func TestParseGNOMEExtensionInstallCommand(t *testing.T) {
	cmd, err := Parse([]string{"gnome-extension", "install"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandGNOMEExtension {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandGNOMEExtension)
	}
	if cmd.Action != "install" {
		t.Fatalf("cmd.Action = %q, want install", cmd.Action)
	}
}

func TestParseGNOMEExtensionStatusCommand(t *testing.T) {
	cmd, err := Parse([]string{"gnome-extension", "status"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandGNOMEExtension {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandGNOMEExtension)
	}
	if cmd.Action != "status" {
		t.Fatalf("cmd.Action = %q, want status", cmd.Action)
	}
}

func TestParseSetCommandRequiresProcess(t *testing.T) {
	_, err := Parse([]string{"set", "--opacity", "70"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestParseSetCommandRequiresOpacity(t *testing.T) {
	_, err := Parse([]string{"set", "--process", "notepad.exe"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestParseSetCommandRejectsBadOpacity(t *testing.T) {
	_, err := Parse([]string{"set", "--process", "notepad.exe", "--opacity", "10"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestParseWatchCommandRejectsBadInterval(t *testing.T) {
	_, err := Parse([]string{"watch", "--interval", "0s"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestParseRestoreCommandRequiresProcess(t *testing.T) {
	_, err := Parse([]string{"restore"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestRememberRuleAddsAndUpdatesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	var stdout bytes.Buffer

	if err := rememberRule(&stdout, path, "Code.exe", "Chrome_WidgetWin_1", "repo", 75); err != nil {
		t.Fatalf("rememberRule add returned error: %v", err)
	}
	if err := rememberRule(&stdout, path, "code.exe", "chrome_widgetwin_1", "REPO", 85); err != nil {
		t.Fatalf("rememberRule update returned error: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(cfg.Rules) != 1 {
		t.Fatalf("len(cfg.Rules) = %d, want 1", len(cfg.Rules))
	}
	rule := cfg.Rules[0]
	if rule.Opacity != 85 || !rule.EnabledValue() {
		t.Fatalf("rule = %#v", rule)
	}
	if !strings.Contains(stdout.String(), "updated rule") {
		t.Fatalf("stdout did not mention update: %s", stdout.String())
	}
}

func TestWatchConfigAppliesEnabledRulesAndSkipsDisabled(t *testing.T) {
	cfg := config.Config{Rules: []config.Rule{
		{Process: "Code.exe", Class: "Editor", TitleContains: "repo", Opacity: 75},
		{Process: "notepad.exe", Opacity: 50, Enabled: config.Bool(false)},
	}}
	windows := []window.Window{
		{ID: "1", Process: "code.exe", Class: "editor", Title: "repo - main.go"},
		{ID: "2", Process: "notepad.exe", Title: "Untitled"},
	}
	var applied []string
	deps := watchDeps{
		Load: func(string) (config.Config, error) {
			return cfg, nil
		},
		List: func() ([]window.Window, error) {
			return windows, nil
		},
		Set: func(win window.Window, opacity int) error {
			applied = append(applied, win.ID)
			return nil
		},
		Wait: func(context.Context, time.Duration) bool {
			return false
		},
	}

	if err := watchConfig(context.Background(), ioDiscard{}, "", time.Second, deps); err != nil {
		t.Fatalf("watchConfig returned error: %v", err)
	}
	if len(applied) != 1 || applied[0] != "1" {
		t.Fatalf("applied = %#v, want [1]", applied)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func TestApplyRulesUsesRuleSelector(t *testing.T) {
	cfg := config.Config{Rules: []config.Rule{{Process: "Code.exe", Class: "Editor", TitleContains: "repo", Opacity: 75}}}
	windows := []window.Window{
		{ID: "1", Process: "Code.exe", Class: "Editor", Title: "repo - main.go"},
		{ID: "2", Process: "Code.exe", Class: "Editor", Title: "other"},
	}
	var applied []string

	err := applyRules(ioDiscard{}, cfg, windows, func(win window.Window, _ int) error {
		applied = append(applied, win.ID)
		return nil
	}, false)
	if err != nil {
		t.Fatalf("applyRules returned error: %v", err)
	}

	if len(applied) != 1 || applied[0] != "1" {
		t.Fatalf("applied = %#v, want [1]", applied)
	}
}

func TestRememberRuleCreatesConfigDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")

	if err := rememberRule(ioDiscard{}, path, "notepad.exe", "", "", 70); err != nil {
		t.Fatalf("rememberRule returned error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}
}
