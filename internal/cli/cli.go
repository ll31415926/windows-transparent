package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"windows-transparent/internal/config"
	"windows-transparent/internal/opacity"
	"windows-transparent/internal/version"
	"windows-transparent/internal/window"
)

type CommandName string

const (
	CommandHelp           CommandName = "help"
	CommandVersion        CommandName = "version"
	CommandList           CommandName = "list"
	CommandSet            CommandName = "set"
	CommandRestore        CommandName = "restore"
	CommandApply          CommandName = "apply"
	CommandDiagnose       CommandName = "diagnose"
	CommandGNOMEExtension CommandName = "gnome-extension"
)

type Command struct {
	Name       CommandName
	Process    string
	Opacity    int
	ConfigPath string
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
	case CommandHelp:
		PrintUsage(stdout)
		return nil
	case CommandVersion:
		version.Print(stdout)
		return nil
	case CommandList:
		return listWindows(stdout, cmd.Process)
	case CommandSet:
		return setProcessOpacity(stdout, cmd.Process, cmd.Opacity)
	case CommandRestore:
		return restoreProcess(stdout, cmd.Process)
	case CommandApply:
		return applyConfig(stdout, cmd.ConfigPath)
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
	case "-h", "--help", string(CommandHelp):
		return Command{Name: CommandHelp}, nil

	case "-v", "--version", string(CommandVersion):
		return Command{Name: CommandVersion}, nil

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

		return Command{Name: CommandSet, Process: strings.TrimSpace(*process), Opacity: *opacityValue}, nil

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
  wtrans help
  wtrans version
  wtrans list [--process notepad.exe]
  wtrans set --process notepad.exe --opacity 70
  wtrans restore --process notepad.exe
  wtrans apply [--config path\to\config.json]
  wtrans diagnose
  wtrans gnome-extension install
  wtrans gnome-extension status

Opacity is a percent from 20 to 100. Process names and window classes are matched case-insensitively.`)
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

func setProcessOpacity(stdout io.Writer, process string, opacityValue int) error {
	windows, err := matchingWindows(process)
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
	windows, err := matchingWindows(process)
	if err != nil {
		return err
	}

	for _, win := range windows {
		if err := window.Restore(win); err != nil {
			return fmt.Errorf("restore %s window %s: %w", win.Process, windowLabel(win), err)
		}
	}

	fmt.Fprintf(stdout, "restored %d window(s) for %s\n", len(windows), process)
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

func matchingWindows(process string) ([]window.Window, error) {
	windows, err := window.ListVisible()
	if err != nil {
		return nil, err
	}

	matches := window.MatchByProcess(windows, process)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no visible windows found for process %q", process)
	}

	return matches, nil
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
