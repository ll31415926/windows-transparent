package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"windows-transparent/internal/config"
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

func TestParseSetCommandWithPersist(t *testing.T) {
	cmd, err := Parse([]string{"set", "--process", "notepad.exe", "--opacity", "70", "--persist"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if !cmd.Persist {
		t.Fatal("cmd.Persist = false, want true")
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

func TestParseWatchCommandWithConfig(t *testing.T) {
	cmd, err := Parse([]string{"watch", "--config", `C:\tmp\wtrans.json`})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandWatch {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandWatch)
	}
	if cmd.ConfigPath != `C:\tmp\wtrans.json` {
		t.Fatalf("cmd.ConfigPath = %q", cmd.ConfigPath)
	}
}

func TestParseWatchCommandWithProcess(t *testing.T) {
	cmd, err := Parse([]string{"watch", "--process", "notepad.exe", "--opacity", "70"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandWatch {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandWatch)
	}
	if cmd.Process != "notepad.exe" {
		t.Fatalf("cmd.Process = %q, want notepad.exe", cmd.Process)
	}
	if cmd.Opacity != 70 {
		t.Fatalf("cmd.Opacity = %d, want 70", cmd.Opacity)
	}
}

func TestParseWatchCommandRejectsConfigWithProcess(t *testing.T) {
	_, err := Parse([]string{"watch", "--config", "rules.json", "--process", "notepad.exe", "--opacity", "70"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestParseStatusCommandWithConfig(t *testing.T) {
	cmd, err := Parse([]string{"status", "--config", `C:\tmp\wtrans.json`})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandStatus {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandStatus)
	}
	if cmd.ConfigPath != `C:\tmp\wtrans.json` {
		t.Fatalf("cmd.ConfigPath = %q", cmd.ConfigPath)
	}
}

func TestParseStopCommandWithConfig(t *testing.T) {
	cmd, err := Parse([]string{"stop", "--config", `C:\tmp\wtrans.json`})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandStop {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandStop)
	}
	if cmd.ConfigPath != `C:\tmp\wtrans.json` {
		t.Fatalf("cmd.ConfigPath = %q", cmd.ConfigPath)
	}
}

func TestParseResetCommandWithConfig(t *testing.T) {
	cmd, err := Parse([]string{"reset", "--config", `C:\tmp\wtrans.json`})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandReset {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandReset)
	}
	if cmd.ConfigPath != `C:\tmp\wtrans.json` {
		t.Fatalf("cmd.ConfigPath = %q", cmd.ConfigPath)
	}
}

func TestParseStatusRejectsUnexpectedArgument(t *testing.T) {
	_, err := Parse([]string{"status", "extra"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
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

func TestParseRestoreCommandRequiresProcess(t *testing.T) {
	_, err := Parse([]string{"restore"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestRunPrintsUsageOnHelp(t *testing.T) {
	var stdout strings.Builder
	var stderr strings.Builder

	code := Run([]string{"-h"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run returned exit code %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "wtrans set --process notepad.exe --opacity 70 [--persist]") {
		t.Fatalf("usage output did not include persist flag: %q", stdout.String())
	}
	for _, want := range []string{
		"wtrans status [--config path\\to\\config.json]",
		"wtrans stop [--config path\\to\\config.json]",
		"wtrans reset [--config path\\to\\config.json]",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("usage output did not include %q: %q", want, stdout.String())
		}
	}
}

func TestPersistOpacityRuleCreatesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := persistOpacityRule(path, "notepad.exe", 70); err != nil {
		t.Fatalf("persistOpacityRule returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, `"process": "notepad.exe"`) {
		t.Fatalf("saved config did not include process: %s", text)
	}
	if !strings.Contains(text, `"opacity": 70`) {
		t.Fatalf("saved config did not include opacity: %s", text)
	}
}

func TestPersistOpacityRuleUpdatesExistingRule(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"rules":[{"process":"notepad.exe","opacity":50}]}`), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if err := persistOpacityRule(path, "notepad.exe", 75); err != nil {
		t.Fatalf("persistOpacityRule returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	text := string(data)
	if strings.Count(text, `"process": "notepad.exe"`) != 1 {
		t.Fatalf("expected one persisted rule, got: %s", text)
	}
	if !strings.Contains(text, `"opacity": 75`) {
		t.Fatalf("saved config did not update opacity: %s", text)
	}
}

func TestRemoveOpacityRuleDeletesMatchingRule(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"rules":[{"process":"notepad.exe","opacity":50},{"process":"Code.exe","opacity":80}]}`), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	removed, err := removeOpacityRule(path, "NOTEPAD.EXE")
	if err != nil {
		t.Fatalf("removeOpacityRule returned error: %v", err)
	}
	if !removed {
		t.Fatal("removed = false, want true")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	text := string(data)
	if strings.Contains(text, `"process": "notepad.exe"`) {
		t.Fatalf("saved config still includes removed rule: %s", text)
	}
	if !strings.Contains(text, `"process": "Code.exe"`) {
		t.Fatalf("saved config removed unrelated rule: %s", text)
	}
}

func TestRemoveOpacityRuleIgnoresMissingRule(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"rules":[{"process":"Code.exe","opacity":80}]}`), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	removed, err := removeOpacityRule(path, "notepad.exe")
	if err != nil {
		t.Fatalf("removeOpacityRule returned error: %v", err)
	}
	if removed {
		t.Fatal("removed = true, want false")
	}
}

func TestWatcherStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := writeWatcherState(path, 12345); err != nil {
		t.Fatalf("writeWatcherState returned error: %v", err)
	}

	state, ok, err := readWatcherState(path)
	if err != nil {
		t.Fatalf("readWatcherState returned error: %v", err)
	}
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if state.PID != 12345 {
		t.Fatalf("state.PID = %d, want 12345", state.PID)
	}

	if err := clearWatcherState(path); err != nil {
		t.Fatalf("clearWatcherState returned error: %v", err)
	}
	_, ok, err = readWatcherState(path)
	if err != nil {
		t.Fatalf("readWatcherState after clear returned error: %v", err)
	}
	if ok {
		t.Fatal("ok after clear = true, want false")
	}
}

func TestWatcherStatusTreatsOldStateAsStopped(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	old := time.Now().Add(-watcherStaleAfter - time.Second).Format(time.RFC3339Nano)
	data := `{"pid":12345,"config":"` + strings.ReplaceAll(path, `\`, `\\`) + `","updated_at":"` + old + `"}`
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(watcherStatePath(path), []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	status, err := watcherStatusForConfig(path)
	if err != nil {
		t.Fatalf("watcherStatusForConfig returned error: %v", err)
	}
	if !status.HasState {
		t.Fatal("HasState = false, want true")
	}
	if !status.Stale {
		t.Fatal("Stale = false, want true")
	}
	if status.Running {
		t.Fatal("Running = true, want false")
	}
}

func TestStopRequestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if stopRequested(path) {
		t.Fatal("stopRequested before write = true, want false")
	}
	if err := writeStopRequest(path); err != nil {
		t.Fatalf("writeStopRequest returned error: %v", err)
	}
	if !stopRequested(path) {
		t.Fatal("stopRequested after write = false, want true")
	}
	if err := removeStopRequest(path); err != nil {
		t.Fatalf("removeStopRequest returned error: %v", err)
	}
	if stopRequested(path) {
		t.Fatal("stopRequested after remove = true, want false")
	}
}

func TestShowStatusPrintsRulesAndStoppedState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Config{
		Rules: []config.Rule{{Process: "notepad.exe", Opacity: 70}},
	}
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	var stdout strings.Builder
	if err := showStatus(&stdout, path); err != nil {
		t.Fatalf("showStatus returned error: %v", err)
	}

	text := stdout.String()
	for _, want := range []string{
		"Config: " + path,
		"Background keeper: stopped",
		"notepad.exe -> 70% opacity",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("status output did not include %q: %q", want, text)
		}
	}
}

func TestShowStatusPrintsRunningState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := config.Save(path, config.Config{}); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	original := watcherProcessRunning
	watcherProcessRunning = func(pid int) bool { return pid == 1234 }
	defer func() { watcherProcessRunning = original }()

	if err := writeWatcherState(path, 1234); err != nil {
		t.Fatalf("writeWatcherState returned error: %v", err)
	}

	var stdout strings.Builder
	if err := showStatus(&stdout, path); err != nil {
		t.Fatalf("showStatus returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Background keeper: running (pid 1234)") {
		t.Fatalf("status output did not show running state: %q", stdout.String())
	}
}

func TestStopWatcherUsesInjectedProcessCheck(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := config.Save(path, config.Config{}); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(watcherStatePath(path)), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := writeWatcherState(path, 1234); err != nil {
		t.Fatalf("writeWatcherState returned error: %v", err)
	}

	originalRunning := watcherProcessRunning
	originalKill := watcherKillProcess
	watcherProcessRunning = func(pid int) bool { return pid == 1234 }
	watcherKillProcess = func(pid int) error {
		if pid != 1234 {
			t.Fatalf("kill pid = %d, want 1234", pid)
		}
		return nil
	}
	defer func() {
		watcherProcessRunning = originalRunning
		watcherKillProcess = originalKill
	}()

	stopped, _, err := stopWatcher(path)
	if err != nil {
		t.Fatalf("stopWatcher returned error: %v", err)
	}
	if !stopped {
		t.Fatal("stopped = false, want true")
	}
}
