package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"windows-transparent/internal/config"
	"windows-transparent/internal/opacity"
	"windows-transparent/internal/window"
)

type CommandName string

const (
	CommandList           CommandName = "list"
	CommandSet            CommandName = "set"
	CommandRestore        CommandName = "restore"
	CommandApply          CommandName = "apply"
	CommandWatch          CommandName = "watch"
	CommandStatus         CommandName = "status"
	CommandStop           CommandName = "stop"
	CommandReset          CommandName = "reset"
	CommandDiagnose       CommandName = "diagnose"
	CommandGNOMEExtension CommandName = "gnome-extension"
)

const (
	watcherStateFile    = "watch.pid"
	watcherStopFile     = "watch.stop"
	watcherTickInterval = 1500 * time.Millisecond
	watcherStaleAfter   = 10 * time.Second
	watcherStopTimeout  = 3 * time.Second
)

var (
	watcherProcessRunning = processRunning
	watcherKillProcess    = killProcess
)

type Command struct {
	Name       CommandName
	Process    string
	Opacity    int
	ConfigPath string
	Persist    bool
	Action     string
}

type UsageError struct {
	Message string
}

func (e UsageError) Error() string {
	return e.Message
}

func IsUsageError(err error) bool {
	var usage UsageError
	return errors.As(err, &usage)
}

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if hasHelpRequest(args) {
		PrintUsage(stdout)
		return 0
	}

	if err := Execute(args, stdout); err != nil {
		if IsUsageError(err) {
			fmt.Fprintf(stderr, "error: %v\n\n", err)
			PrintUsage(stderr)
			return 2
		}

		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	return 0
}

func Execute(args []string, stdout io.Writer) error {
	cmd, err := Parse(args)
	if err != nil {
		return err
	}

	switch cmd.Name {
	case CommandList:
		return listWindows(stdout, cmd.Process)
	case CommandSet:
		return setProcessOpacity(stdout, cmd.Process, cmd.Opacity, cmd.Persist)
	case CommandRestore:
		return restoreProcess(stdout, cmd.Process)
	case CommandApply:
		return applyConfig(stdout, cmd.ConfigPath)
	case CommandWatch:
		if cmd.Process != "" {
			return watchProcessOpacity(stdout, cmd.Process, cmd.Opacity)
		}
		return watchConfigOpacity(stdout, cmd.ConfigPath)
	case CommandStatus:
		return showStatus(stdout, cmd.ConfigPath)
	case CommandStop:
		return stopPersistentWatcher(stdout, cmd.ConfigPath)
	case CommandReset:
		return resetPersistentRules(stdout, cmd.ConfigPath)
	case CommandDiagnose:
		return window.Diagnose(stdout)
	case CommandGNOMEExtension:
		return runGNOMEExtension(stdout, cmd.Action)
	default:
		return UsageError{Message: fmt.Sprintf("unknown command %q", cmd.Name)}
	}
}

func Parse(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, UsageError{Message: "missing command"}
	}

	switch args[0] {
	case string(CommandList):
		fs := newFlagSet("list")
		process := fs.String("process", "", "process name")
		if err := fs.Parse(args[1:]); err != nil {
			return Command{}, UsageError{Message: err.Error()}
		}
		if fs.NArg() != 0 {
			return Command{}, UsageError{Message: fmt.Sprintf("unexpected argument %q", fs.Arg(0))}
		}

		return Command{Name: CommandList, Process: strings.TrimSpace(*process)}, nil

	case string(CommandSet):
		fs := newFlagSet("set")
		process := fs.String("process", "", "process name")
		opacityValue := fs.Int("opacity", 0, "opacity percent")
		persist := fs.Bool("persist", false, "keep applying this opacity to new matching windows")
		if err := fs.Parse(args[1:]); err != nil {
			return Command{}, UsageError{Message: err.Error()}
		}
		if fs.NArg() != 0 {
			return Command{}, UsageError{Message: fmt.Sprintf("unexpected argument %q", fs.Arg(0))}
		}
		if strings.TrimSpace(*process) == "" {
			return Command{}, UsageError{Message: "missing --process"}
		}
		if !hasFlag(args[1:], "opacity") {
			return Command{}, UsageError{Message: "missing --opacity"}
		}
		if err := opacity.Validate(*opacityValue); err != nil {
			return Command{}, UsageError{Message: err.Error()}
		}

		return Command{Name: CommandSet, Process: strings.TrimSpace(*process), Opacity: *opacityValue, Persist: *persist}, nil

	case string(CommandWatch):
		fs := newFlagSet("watch")
		configPath := fs.String("config", "", "config path")
		process := fs.String("process", "", "process name")
		opacityValue := fs.Int("opacity", 0, "opacity percent")
		if err := fs.Parse(args[1:]); err != nil {
			return Command{}, UsageError{Message: err.Error()}
		}
		if fs.NArg() != 0 {
			return Command{}, UsageError{Message: fmt.Sprintf("unexpected argument %q", fs.Arg(0))}
		}
		if strings.TrimSpace(*process) != "" || hasFlag(args[1:], "opacity") {
			if strings.TrimSpace(*configPath) != "" {
				return Command{}, UsageError{Message: "--config cannot be combined with --process or --opacity"}
			}
			if strings.TrimSpace(*process) == "" {
				return Command{}, UsageError{Message: "missing --process"}
			}
			if !hasFlag(args[1:], "opacity") {
				return Command{}, UsageError{Message: "missing --opacity"}
			}
			if err := opacity.Validate(*opacityValue); err != nil {
				return Command{}, UsageError{Message: err.Error()}
			}

			return Command{Name: CommandWatch, Process: strings.TrimSpace(*process), Opacity: *opacityValue}, nil
		}

		return Command{Name: CommandWatch, ConfigPath: strings.TrimSpace(*configPath)}, nil

	case string(CommandRestore):
		fs := newFlagSet("restore")
		process := fs.String("process", "", "process name")
		if err := fs.Parse(args[1:]); err != nil {
			return Command{}, UsageError{Message: err.Error()}
		}
		if fs.NArg() != 0 {
			return Command{}, UsageError{Message: fmt.Sprintf("unexpected argument %q", fs.Arg(0))}
		}
		if strings.TrimSpace(*process) == "" {
			return Command{}, UsageError{Message: "missing --process"}
		}

		return Command{Name: CommandRestore, Process: strings.TrimSpace(*process)}, nil

	case string(CommandApply):
		fs := newFlagSet("apply")
		configPath := fs.String("config", "", "config path")
		if err := fs.Parse(args[1:]); err != nil {
			return Command{}, UsageError{Message: err.Error()}
		}
		if fs.NArg() != 0 {
			return Command{}, UsageError{Message: fmt.Sprintf("unexpected argument %q", fs.Arg(0))}
		}

		return Command{Name: CommandApply, ConfigPath: strings.TrimSpace(*configPath)}, nil

	case string(CommandStatus):
		configPath, err := parseOptionalConfigFlag("status", args[1:])
		if err != nil {
			return Command{}, err
		}

		return Command{Name: CommandStatus, ConfigPath: configPath}, nil

	case string(CommandStop):
		configPath, err := parseOptionalConfigFlag("stop", args[1:])
		if err != nil {
			return Command{}, err
		}

		return Command{Name: CommandStop, ConfigPath: configPath}, nil

	case string(CommandReset):
		configPath, err := parseOptionalConfigFlag("reset", args[1:])
		if err != nil {
			return Command{}, err
		}

		return Command{Name: CommandReset, ConfigPath: configPath}, nil

	case string(CommandDiagnose):
		fs := newFlagSet("diagnose")
		if err := fs.Parse(args[1:]); err != nil {
			return Command{}, UsageError{Message: err.Error()}
		}
		if fs.NArg() != 0 {
			return Command{}, UsageError{Message: fmt.Sprintf("unexpected argument %q", fs.Arg(0))}
		}

		return Command{Name: CommandDiagnose}, nil

	case string(CommandGNOMEExtension):
		if len(args) < 2 {
			return Command{}, UsageError{Message: "missing gnome-extension action"}
		}
		action := strings.TrimSpace(args[1])
		if action != "install" && action != "status" {
			return Command{}, UsageError{Message: fmt.Sprintf("unknown gnome-extension action %q", action)}
		}
		fs := newFlagSet("gnome-extension")
		if err := fs.Parse(args[2:]); err != nil {
			return Command{}, UsageError{Message: err.Error()}
		}
		if fs.NArg() != 0 {
			return Command{}, UsageError{Message: fmt.Sprintf("unexpected argument %q", fs.Arg(0))}
		}

		return Command{Name: CommandGNOMEExtension, Action: action}, nil

	default:
		return Command{}, UsageError{Message: fmt.Sprintf("unknown command %q", args[0])}
	}
}

func PrintUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  wtrans -h | --help
  wtrans list [--process notepad.exe]
  wtrans set --process notepad.exe --opacity 70 [--persist]
  wtrans restore --process notepad.exe
  wtrans apply [--config path\to\config.json]
  wtrans watch [--config path\to\config.json]
  wtrans status [--config path\to\config.json]
  wtrans stop [--config path\to\config.json]
  wtrans reset [--config path\to\config.json]
  wtrans diagnose
  wtrans gnome-extension install
  wtrans gnome-extension status

Opacity is a percent from 20 to 100. Process names are matched case-insensitively.
Use --persist with set to keep newly opened matching windows at the same opacity.`)
}

func listWindows(stdout io.Writer, process string) error {
	windows, err := window.ListVisible()
	if err != nil {
		return err
	}

	windows = window.MatchByProcess(windows, process)
	printWindows(stdout, windows)
	return nil
}

func setProcessOpacity(stdout io.Writer, process string, opacityValue int, persist bool) error {
	if persist {
		if err := persistOpacityRule("", process, opacityValue); err != nil {
			return fmt.Errorf("persist rule for %s: %w", process, err)
		}

		configPath, err := config.ResolvePath("")
		if err != nil {
			return err
		}

		applied := applyOpacityBestEffort(process, opacityValue)
		started, err := ensurePersistentWatcher(configPath)
		if err != nil {
			return fmt.Errorf("start persistent watcher for %s: %w", process, err)
		}

		state := "background keeper is already running"
		if started {
			state = "background keeper started"
		}
		if applied == 0 {
			fmt.Fprintf(stdout, "saved persistent rule for %s at %d%% opacity; %s and will apply it when a matching window appears\n", process, opacityValue, state)
		} else {
			fmt.Fprintf(stdout, "set %d window(s) for %s to %d%% opacity and saved persistent rule; %s\n", applied, process, opacityValue, state)
		}
		return nil
	}

	windows, err := matchingWindows(process, false)
	if err != nil {
		return err
	}

	for _, win := range windows {
		if err := window.SetOpacity(win, opacityValue); err != nil {
			return fmt.Errorf("set %s window %s: %w", win.Process, windowLabel(win), err)
		}
	}

	fmt.Fprintf(stdout, "set %d window(s) for %s to %d%% opacity\n", len(windows), process, opacityValue)
	return nil
}

func restoreProcess(stdout io.Writer, process string) error {
	windows, err := matchingWindows(process, true)
	if err != nil {
		return err
	}

	for _, win := range windows {
		if err := window.Restore(win); err != nil {
			return fmt.Errorf("restore %s window %s: %w", win.Process, windowLabel(win), err)
		}
	}

	removed, err := removeOpacityRule("", process)
	if err != nil {
		return fmt.Errorf("remove persistent rule for %s: %w", process, err)
	}
	if len(windows) == 0 {
		if removed {
			fmt.Fprintf(stdout, "All done. %s is back to normal, and its keep-transparent rule has been turned off.\n", process)
			return nil
		}

		fmt.Fprintf(stdout, "I couldn't find a visible %s window, and there was no saved rule for it.\n", process)
		return nil
	}
	if removed {
		fmt.Fprintf(stdout, "Restored %d visible %s window(s) and turned off keep-transparent for this app.\n", len(windows), process)
		return nil
	}

	fmt.Fprintf(stdout, "Restored %d visible %s window(s).\n", len(windows), process)
	return nil
}

func applyConfig(stdout io.Writer, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	windows, err := window.ListVisible()
	if err != nil {
		return err
	}

	if len(cfg.Rules) == 0 {
		fmt.Fprintln(stdout, "no rules found in config")
		return nil
	}

	for _, rule := range cfg.Rules {
		matches := window.MatchByProcess(windows, rule.Process)
		if len(matches) == 0 {
			fmt.Fprintf(stdout, "skipped %s: no visible windows found\n", rule.Process)
			continue
		}

		for _, win := range matches {
			if err := window.SetOpacity(win, rule.Opacity); err != nil {
				return fmt.Errorf("apply rule for %s window %s: %w", win.Process, windowLabel(win), err)
			}
		}

		fmt.Fprintf(stdout, "set %d window(s) for %s to %d%% opacity\n", len(matches), rule.Process, rule.Opacity)
	}

	return nil
}

func showStatus(stdout io.Writer, configPath string) error {
	resolved, err := config.ResolvePath(configPath)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Config: %s\n", resolved)

	status, err := watcherStatusForConfig(resolved)
	if err != nil {
		return err
	}
	if status.Running {
		fmt.Fprintf(stdout, "Background keeper: running (pid %d)\n", status.State.PID)
	} else if status.HasState && status.Stale {
		fmt.Fprintf(stdout, "Background keeper: stopped (last seen %s)\n", status.LastSeen.Format(time.RFC3339))
	} else {
		fmt.Fprintln(stdout, "Background keeper: stopped")
	}

	cfg, exists, err := loadOpacityConfig(resolved)
	if err != nil {
		return err
	}
	if !exists || len(cfg.Rules) == 0 {
		fmt.Fprintln(stdout, "Saved rules: none yet")
		return nil
	}

	fmt.Fprintln(stdout, "Saved rules:")
	for _, rule := range cfg.Rules {
		fmt.Fprintf(stdout, "  %s -> %d%% opacity\n", rule.Process, rule.Opacity)
	}

	return nil
}

func stopPersistentWatcher(stdout io.Writer, configPath string) error {
	resolved, err := config.ResolvePath(configPath)
	if err != nil {
		return err
	}

	stopped, _, err := stopWatcher(resolved)
	if err != nil {
		return err
	}
	if stopped {
		fmt.Fprintln(stdout, "Stopped background keeper. Saved rules are still there; run wtrans set --process APP.exe --opacity 85 --persist to start it again.")
		return nil
	}

	fmt.Fprintln(stdout, "Background keeper is already stopped. Saved rules are still there.")
	return nil
}

func resetPersistentRules(stdout io.Writer, configPath string) error {
	resolved, err := config.ResolvePath(configPath)
	if err != nil {
		return err
	}

	cfg, exists, err := loadOpacityConfig(resolved)
	if err != nil {
		return err
	}

	stopped, _, err := stopWatcher(resolved)
	if err != nil {
		return err
	}

	restored, restoreErr := restoreRulesBestEffort(cfg.Rules)
	if exists {
		if err := config.Save(resolved, config.Config{}); err != nil {
			return err
		}
	}

	if !exists || len(cfg.Rules) == 0 {
		if stopped {
			fmt.Fprintln(stdout, "Reset complete. Background keeper is stopped and there were no saved rules to clear.")
			return nil
		}
		fmt.Fprintln(stdout, "Reset complete. There were no saved rules to clear.")
		return nil
	}

	fmt.Fprintf(stdout, "Reset complete. Cleared %d saved rule(s)", len(cfg.Rules))
	if stopped {
		fmt.Fprint(stdout, " and stopped the background keeper")
	}
	fmt.Fprintf(stdout, ". Restored %d visible window(s).\n", restored)
	if restoreErr != nil {
		fmt.Fprintf(stdout, "Note: saved rules were cleared, but visible windows could not be checked: %v\n", restoreErr)
	}

	return nil
}

func runGNOMEExtension(stdout io.Writer, action string) error {
	switch action {
	case "install":
		return window.InstallGNOMEExtension(stdout)
	case "status":
		return window.GNOMEExtensionStatus(stdout)
	default:
		return UsageError{Message: fmt.Sprintf("unknown gnome-extension action %q", action)}
	}
}

func parseOptionalConfigFlag(name string, args []string) (string, error) {
	fs := newFlagSet(name)
	configPath := fs.String("config", "", "config path")
	if err := fs.Parse(args); err != nil {
		return "", UsageError{Message: err.Error()}
	}
	if fs.NArg() != 0 {
		return "", UsageError{Message: fmt.Sprintf("unexpected argument %q", fs.Arg(0))}
	}

	return strings.TrimSpace(*configPath), nil
}

func matchingWindows(process string, allowEmpty bool) ([]window.Window, error) {
	windows, err := window.ListVisible()
	if err != nil {
		return nil, err
	}

	matches := window.MatchByProcess(windows, process)
	if len(matches) == 0 && !allowEmpty {
		return nil, fmt.Errorf("no visible windows found for process %q", process)
	}

	return matches, nil
}

func watchProcessOpacity(stdout io.Writer, process string, opacityValue int) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	applied := applyOpacityBestEffort(process, opacityValue)
	if applied == 0 {
		fmt.Fprintf(stdout, "watching %s at %d%% opacity; waiting for matching windows\n", process, opacityValue)
	} else {
		fmt.Fprintf(stdout, "watching %s at %d%% opacity\n", process, opacityValue)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			_ = applyOpacityBestEffort(process, opacityValue)
		}
	}
}

func watchConfigOpacity(stdout io.Writer, configPath string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	resolved, err := config.ResolvePath(configPath)
	if err != nil {
		return err
	}
	pid := os.Getpid()
	if err := removeStopRequest(resolved); err != nil {
		return err
	}
	if err := writeWatcherState(resolved, pid); err != nil {
		return err
	}
	defer clearWatcherStateIfCurrent(resolved, pid)

	ticker := time.NewTicker(watcherTickInterval)
	defer ticker.Stop()

	result := applyConfigBestEffort(resolved)
	if result.loaded && result.rules == 0 {
		fmt.Fprintf(stdout, "no persistent rules found in %s\n", resolved)
		return nil
	}
	if result.applied == 0 {
		fmt.Fprintf(stdout, "watching persistent rules from %s; waiting for matching windows\n", resolved)
	} else {
		fmt.Fprintf(stdout, "watching persistent rules from %s\n", resolved)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if stopRequested(resolved) {
				return nil
			}
			if err := writeWatcherState(resolved, pid); err != nil {
				return err
			}
			result := applyConfigBestEffort(resolved)
			if result.loaded && result.rules == 0 {
				return nil
			}
		}
	}
}

func applyOpacityBestEffort(process string, opacityValue int) int {
	windows, err := window.ListVisible()
	if err != nil {
		return 0
	}

	applied := 0
	for _, win := range window.MatchByProcess(windows, process) {
		if err := window.SetOpacity(win, opacityValue); err == nil {
			applied++
		}
	}

	return applied
}

type configApplyResult struct {
	applied int
	rules   int
	loaded  bool
}

func applyConfigBestEffort(configPath string) configApplyResult {
	cfg, err := config.Load(configPath)
	if err != nil {
		return configApplyResult{}
	}

	windows, err := window.ListVisible()
	if err != nil {
		return configApplyResult{rules: len(cfg.Rules), loaded: true}
	}

	applied := 0
	for _, rule := range cfg.Rules {
		for _, win := range window.MatchByProcess(windows, rule.Process) {
			if err := window.SetOpacity(win, rule.Opacity); err == nil {
				applied++
			}
		}
	}

	return configApplyResult{applied: applied, rules: len(cfg.Rules), loaded: true}
}

type watcherState struct {
	PID       int       `json:"pid"`
	Config    string    `json:"config"`
	UpdatedAt time.Time `json:"updated_at"`
}

type watcherStatus struct {
	HasState bool
	Running  bool
	Stale    bool
	LastSeen time.Time
	State    watcherState
}

func ensurePersistentWatcher(configPath string) (bool, error) {
	status, err := watcherStatusForConfig(configPath)
	if err != nil {
		return false, err
	}
	if status.Running {
		return false, nil
	}

	if err := removeStopRequest(configPath); err != nil {
		return false, err
	}
	if err := clearWatcherState(configPath); err != nil {
		return false, err
	}

	pid, err := launchPersistentWatcher(configPath)
	if err != nil {
		return false, err
	}
	if pid > 0 {
		if err := writeWatcherState(configPath, pid); err != nil {
			return false, err
		}
	}

	return true, nil
}

func watcherStatusForConfig(configPath string) (watcherStatus, error) {
	state, ok, err := readWatcherState(configPath)
	if err != nil {
		return watcherStatus{}, err
	}
	if !ok {
		return watcherStatus{}, nil
	}

	status := watcherStatus{
		HasState: true,
		State:    state,
		LastSeen: state.UpdatedAt,
	}
	status.Running = watcherProcessRunning(state.PID)
	status.Stale = state.PID <= 0 || state.UpdatedAt.IsZero() || time.Since(state.UpdatedAt) > watcherStaleAfter

	return status, nil
}

func stopWatcher(configPath string) (bool, watcherStatus, error) {
	status, err := watcherStatusForConfig(configPath)
	if err != nil {
		return false, watcherStatus{}, err
	}
	if !status.Running {
		if err := removeStopRequest(configPath); err != nil {
			return false, status, err
		}
		if status.HasState {
			if err := clearWatcherState(configPath); err != nil {
				return false, status, err
			}
		}
		return false, status, nil
	}

	if err := writeStopRequest(configPath); err != nil {
		return false, status, err
	}

	deadline := time.Now().Add(watcherStopTimeout)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		next, err := watcherStatusForConfig(configPath)
		if err != nil {
			return false, status, err
		}
		if !next.Running {
			if err := removeStopRequest(configPath); err != nil {
				return true, status, err
			}
			if err := clearWatcherState(configPath); err != nil {
				return true, status, err
			}
			return true, status, nil
		}
	}

	if err := watcherKillProcess(status.State.PID); err != nil {
		return false, status, err
	}
	if err := removeStopRequest(configPath); err != nil {
		return true, status, err
	}
	if err := clearWatcherState(configPath); err != nil {
		return true, status, err
	}

	return true, status, nil
}

func writeWatcherState(configPath string, pid int) error {
	state := watcherState{
		PID:       pid,
		Config:    configPath,
		UpdatedAt: time.Now(),
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	path := watcherStatePath(configPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create watcher state directory: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write watcher state %s: %w", path, err)
	}

	return nil
}

func readWatcherState(configPath string) (watcherState, bool, error) {
	path := watcherStatePath(configPath)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return watcherState{}, false, nil
	}
	if err != nil {
		return watcherState{}, false, fmt.Errorf("read watcher state %s: %w", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return watcherState{}, false, nil
	}

	var state watcherState
	if err := json.Unmarshal(data, &state); err != nil {
		return watcherState{}, false, fmt.Errorf("parse watcher state %s: %w", path, err)
	}

	return state, true, nil
}

func clearWatcherStateIfCurrent(configPath string, pid int) {
	state, ok, err := readWatcherState(configPath)
	if err != nil || !ok || state.PID != pid {
		return
	}

	_ = clearWatcherState(configPath)
	_ = removeStopRequest(configPath)
}

func clearWatcherState(configPath string) error {
	if err := os.Remove(watcherStatePath(configPath)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove watcher state: %w", err)
	}

	return nil
}

func writeStopRequest(configPath string) error {
	path := watcherStopPath(configPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create watcher stop directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(time.Now().Format(time.RFC3339Nano)+"\n"), 0o644); err != nil {
		return fmt.Errorf("write watcher stop request %s: %w", path, err)
	}

	return nil
}

func removeStopRequest(configPath string) error {
	if err := os.Remove(watcherStopPath(configPath)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove watcher stop request: %w", err)
	}

	return nil
}

func stopRequested(configPath string) bool {
	_, err := os.Stat(watcherStopPath(configPath))
	return err == nil
}

func watcherStatePath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), watcherStateFile)
}

func watcherStopPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), watcherStopFile)
}

func loadOpacityConfig(configPath string) (config.Config, bool, error) {
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return config.Config{}, false, nil
	}
	if err != nil {
		return config.Config{}, false, fmt.Errorf("read config %s: %w", configPath, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return config.Config{}, true, nil
	}

	cfg, err := config.Parse(bytes.NewReader(data))
	if err != nil {
		return config.Config{}, true, fmt.Errorf("parse config %s: %w", configPath, err)
	}

	return cfg, true, nil
}

func restoreRulesBestEffort(rules []config.Rule) (int, error) {
	if len(rules) == 0 {
		return 0, nil
	}

	windows, err := window.ListVisible()
	if err != nil {
		return 0, err
	}

	restored := 0
	var firstErr error
	seen := make(map[string]struct{})
	for _, rule := range rules {
		for _, win := range window.MatchByProcess(windows, rule.Process) {
			key := windowLabel(win)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			if err := window.Restore(win); err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("restore %s window %s: %w", win.Process, windowLabel(win), err)
				}
				continue
			}
			restored++
		}
	}

	return restored, firstErr
}

func persistOpacityRule(configPath, process string, opacityValue int) error {
	resolved, err := config.ResolvePath(configPath)
	if err != nil {
		return err
	}

	cfg := config.Config{}
	data, err := os.ReadFile(resolved)
	switch {
	case err == nil:
		if len(bytes.TrimSpace(data)) != 0 {
			cfg, err = config.Parse(bytes.NewReader(data))
			if err != nil {
				return fmt.Errorf("parse config %s: %w", resolved, err)
			}
		}
	case os.IsNotExist(err):
		// Start with an empty config.
	default:
		return fmt.Errorf("read config %s: %w", resolved, err)
	}

	cfg.Rules = upsertRule(cfg.Rules, process, opacityValue)
	if err := config.Save(resolved, cfg); err != nil {
		return err
	}

	return nil
}

func upsertRule(rules []config.Rule, process string, opacityValue int) []config.Rule {
	process = strings.TrimSpace(process)
	for i := range rules {
		if strings.EqualFold(strings.TrimSpace(rules[i].Process), process) {
			rules[i].Process = process
			rules[i].Opacity = opacityValue
			return rules
		}
	}

	return append(rules, config.Rule{Process: process, Opacity: opacityValue})
}

func removeOpacityRule(configPath, process string) (bool, error) {
	resolved, err := config.ResolvePath(configPath)
	if err != nil {
		return false, err
	}

	data, err := os.ReadFile(resolved)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read config %s: %w", resolved, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return false, nil
	}

	cfg, err := config.Parse(bytes.NewReader(data))
	if err != nil {
		return false, fmt.Errorf("parse config %s: %w", resolved, err)
	}

	process = strings.TrimSpace(process)
	rules := cfg.Rules[:0]
	removed := false
	for _, rule := range cfg.Rules {
		if strings.EqualFold(strings.TrimSpace(rule.Process), process) {
			removed = true
			continue
		}
		rules = append(rules, rule)
	}
	if !removed {
		return false, nil
	}

	cfg.Rules = rules
	if err := config.Save(resolved, cfg); err != nil {
		return false, err
	}

	return true, nil
}

func printWindows(stdout io.Writer, windows []window.Window) {
	if len(windows) == 0 {
		fmt.Fprintln(stdout, "no visible windows found")
		return
	}

	fmt.Fprintf(stdout, "%-18s %-8s %-24s %-18s %-10s %s\n", "HANDLE", "PID", "PROCESS", "CLASS", "BACKEND", "TITLE")
	for _, win := range windows {
		fmt.Fprintf(stdout, "%-18s %-8d %-24s %-18s %-10s %s\n", windowLabel(win), win.PID, win.Process, win.Class, win.Backend, win.Title)
	}
}

func windowLabel(win window.Window) string {
	if win.ID != "" {
		return win.ID
	}
	if win.Handle != 0 {
		return fmt.Sprintf("0x%X", win.Handle)
	}

	return "<unknown>"
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func hasFlag(args []string, name string) bool {
	long := "--" + name
	short := "-" + name
	for _, arg := range args {
		if arg == long || arg == short || strings.HasPrefix(arg, long+"=") || strings.HasPrefix(arg, short+"=") {
			return true
		}
	}

	return false
}

func hasHelpRequest(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}

	return false
}
