//go:build desktop

package gui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	title            = "wtrans"
	singleInstanceID = "f61e94e7-5bc7-4b0d-b61e-5b5d8b90dd13"
)

func Run(cfg Config) error {
	if cfg.Assets == nil {
		return fmt.Errorf("missing frontend assets")
	}

	app := NewApp()
	return wails.Run(&options.App{
		Title:         title,
		Width:         1060,
		Height:        690,
		Frameless:     true,
		DisableResize: true,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               singleInstanceID,
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		AssetServer:      &assetserver.Options{Assets: cfg.Assets},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			DisableFramelessWindowDecorations: true,
			Theme:                             windows.SystemDefault,
		},
		Mac: &mac.Options{
			WindowIsTranslucent:  true,
			WebviewIsTransparent: true,
			TitleBar:             mac.TitleBarHiddenInset(),
		},
		Linux: &linux.Options{
			WindowIsTranslucent: true,
			ProgramName:         title,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyOnDemand,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			go startTray(
				func() {
					if app.ctx != nil {
						wailsruntime.WindowShow(app.ctx)
					}
				},
				func() {
					if app.ctx != nil {
						wailsruntime.Quit(app.ctx)
					}
				},
			)
		},
		OnDomReady: func(_ context.Context) {
			go applySelfOpacity()
		},
		OnBeforeClose: func(ctx context.Context) bool {
			prefs := loadClosePreference()
			if prefs.Remember && prefs.Action == "exit" {
				stopTray()
				return true
			}
			wailsruntime.WindowHide(ctx)
			return false
		},
		Bind: []interface{}{
			app,
		},
	})
}

func applySelfOpacity() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	processName := filepath.Base(exePath)
	time.Sleep(600 * time.Millisecond)
	_, _ = executeCLI("set", "--process", processName, "--opacity", "75", "--persist")
}
