package gui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/options"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"windows-transparent/internal/cli"
	"windows-transparent/internal/config"
	"windows-transparent/internal/opacity"
	"windows-transparent/internal/window"
)

type App struct {
	ctx context.Context
}

type DashboardState struct {
	Windows      []WindowView `json:"windows"`
	Rules        []RuleView   `json:"rules"`
	ConfigPath   string       `json:"configPath"`
	Backend      string       `json:"backend"`
	KeeperStatus string       `json:"keeperStatus"`
	StatusText   string       `json:"statusText"`
}

type WindowView struct {
	ID        string `json:"id"`
	PID       uint32 `json:"pid"`
	Process   string `json:"process"`
	Class     string `json:"class"`
	Title     string `json:"title"`
	Backend   string `json:"backend"`
	Opacity   int    `json:"opacity"`
	HasRule   bool   `json:"hasRule"`
	Icon      string `json:"icon"`
	IconClass string `json:"iconClass"`
}

type RuleView struct {
	Process   string `json:"process"`
	Opacity   int    `json:"opacity"`
	Subtitle  string `json:"subtitle"`
	Icon      string `json:"icon"`
	IconClass string `json:"iconClass"`
}

type OpacityRequest struct {
	Process string `json:"process"`
	Opacity int    `json:"opacity"`
	Persist bool   `json:"persist"`
}

type OperationResult struct {
	Message string         `json:"message"`
	State   DashboardState `json:"state"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) onSecondInstanceLaunch(_ options.SecondInstanceData) {
	if a.ctx == nil {
		return
	}

	wailsruntime.WindowUnminimise(a.ctx)
	wailsruntime.WindowShow(a.ctx)
}

func (a *App) ListWindows() (DashboardState, error) {
	return buildDashboardState("")
}

func (a *App) ApplyOpacity(req OpacityRequest) (OperationResult, error) {
	process := strings.TrimSpace(req.Process)
	if process == "" {
		return OperationResult{}, errors.New("missing process")
	}
	if err := opacity.Validate(req.Opacity); err != nil {
		return OperationResult{}, err
	}

	args := []string{"set", "--process", process, "--opacity", fmt.Sprintf("%d", req.Opacity)}
	if req.Persist {
		args = append(args, "--persist")
	}

	message, err := executeCLI(args...)
	if err != nil {
		return OperationResult{}, err
	}

	state, stateErr := buildDashboardState(message)
	if stateErr != nil {
		return OperationResult{}, stateErr
	}

	return OperationResult{Message: message, State: state}, nil
}

func (a *App) RestoreProcess(process string) (OperationResult, error) {
	process = strings.TrimSpace(process)
	if process == "" {
		return OperationResult{}, errors.New("missing process")
	}

	message, err := executeCLI("restore", "--process", process)
	if err != nil {
		return OperationResult{}, err
	}

	state, stateErr := buildDashboardState(message)
	if stateErr != nil {
		return OperationResult{}, stateErr
	}

	return OperationResult{Message: message, State: state}, nil
}

func (a *App) Diagnose() (string, error) {
	return executeCLI("diagnose")
}

func (a *App) CloseWindow() {
	if a.ctx != nil {
		wailsruntime.Quit(a.ctx)
	}
}

type ClosePreference struct {
	Action   string `json:"action"`
	Remember bool   `json:"remember"`
}

func (a *App) GetClosePreference() ClosePreference {
	prefs := loadClosePreference()
	return ClosePreference{Action: prefs.Action, Remember: prefs.Remember}
}

func (a *App) SaveClosePreference(action string, remember bool) {
	_ = saveClosePreference(closePreferences{Action: action, Remember: remember})
}

func (a *App) QuitApp() {
	if a.ctx != nil {
		stopTray()
		wailsruntime.Quit(a.ctx)
	}
}

func (a *App) HideToTray() {
	if a.ctx != nil {
		wailsruntime.WindowHide(a.ctx)
	}
}

func (a *App) MinimizeWindow() {
	if a.ctx != nil {
		wailsruntime.WindowMinimise(a.ctx)
	}
}

func (a *App) ToggleMaximizeWindow() {
	if a.ctx != nil {
		wailsruntime.WindowToggleMaximise(a.ctx)
	}
}

func (a *App) WindowMinimise() {
	a.MinimizeWindow()
}

func (a *App) WindowToggleMaximise() {
	a.ToggleMaximizeWindow()
}

func executeCLI(args ...string) (string, error) {
	var out bytes.Buffer
	if err := cli.Execute(args, &out); err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

func buildDashboardState(status string) (DashboardState, error) {
	configPath, err := config.ResolvePath("")
	if err != nil {
		return DashboardState{}, err
	}

	rules, err := loadRules(configPath)
	if err != nil {
		return DashboardState{}, err
	}

	windows, err := window.ListVisible()
	if err != nil {
		return DashboardState{}, err
	}

	ruleMap := make(map[string]config.Rule, len(rules))
	for _, rule := range rules {
		ruleMap[strings.ToLower(strings.TrimSpace(rule.Process))] = rule
	}

	views := make([]WindowView, 0, len(windows))
	for _, win := range windows {
		process := strings.TrimSpace(win.Process)
		title := strings.TrimSpace(win.Title)

		rule, hasRule := ruleMap[strings.ToLower(process)]
		opacityValue := 100
		if hasRule {
			opacityValue = rule.Opacity
		}
		icon, iconClass := iconForProcess(process)

		views = append(views, WindowView{
			ID:        windowID(win),
			PID:       win.PID,
			Process:   fallback(process, "unknown"),
			Class:     win.Class,
			Title:     fallback(title, windowID(win)),
			Backend:   fallback(win.Backend, runtime.GOOS),
			Opacity:   opacityValue,
			HasRule:   hasRule,
			Icon:      icon,
			IconClass: iconClass,
		})
	}

	sort.SliceStable(views, func(i, j int) bool {
		left := strings.ToLower(views[i].Process + views[i].Title)
		right := strings.ToLower(views[j].Process + views[j].Title)
		return left < right
	})

	ruleViews := make([]RuleView, 0, len(rules))
	for _, rule := range rules {
		icon, iconClass := iconForProcess(rule.Process)
		ruleViews = append(ruleViews, RuleView{
			Process:   rule.Process,
			Opacity:   rule.Opacity,
			Subtitle:  "persistent rule",
			Icon:      icon,
			IconClass: iconClass,
		})
	}

	backend := runtime.GOOS
	if len(views) > 0 {
		backend = views[0].Backend
	}
	if status == "" {
		status = fmt.Sprintf("loaded %d visible window(s)", len(views))
	}

	return DashboardState{
		Windows:      views,
		Rules:        ruleViews,
		ConfigPath:   configPath,
		Backend:      backend,
		KeeperStatus: watcherStatusText(configPath),
		StatusText:   status,
	}, nil
}

func loadRules(configPath string) ([]config.Rule, error) {
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", configPath, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}

	cfg, err := config.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse config %s: %w", configPath, err)
	}

	return cfg.Rules, nil
}

func watcherStatusText(configPath string) string {
	path := filepath.Join(filepath.Dir(configPath), "watch.pid")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "stopped"
	}
	if err != nil || len(bytes.TrimSpace(data)) == 0 {
		return "unknown"
	}

	return "running"
}

func windowID(win window.Window) string {
	if strings.TrimSpace(win.ID) != "" {
		return win.ID
	}
	if win.Handle != 0 {
		return fmt.Sprintf("0x%X", win.Handle)
	}

	return fmt.Sprintf("%s:%d:%s", win.Process, win.PID, win.Title)
}

func iconForProcess(process string) (string, string) {
	name := strings.ToLower(process)
	switch {
	case strings.Contains(name, "code"), strings.Contains(name, "cursor"), strings.Contains(name, "idea"), strings.Contains(name, "goland"):
		return "code-2", "dev"
	case strings.Contains(name, "notepad"), strings.Contains(name, "word"), strings.Contains(name, "excel"), strings.Contains(name, "wps"):
		return "file-text", "doc"
	case strings.Contains(name, "firefox"), strings.Contains(name, "chrome"), strings.Contains(name, "edge"), strings.Contains(name, "browser"):
		return "globe", "web"
	case strings.Contains(name, "terminal"), strings.Contains(name, "powershell"), strings.Contains(name, "cmd"), strings.Contains(name, "wezterm"), strings.Contains(name, "alacritty"):
		return "terminal", "term"
	default:
		return "app-window", "app"
	}
}

func fallback(value string, replacement string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return replacement
	}

	return value
}
