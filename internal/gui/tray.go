//go:build desktop

package gui

import (
	"github.com/energye/systray"
)

var (
	trayOnShow func()
	trayOnQuit func()
)

func startTray(onShow, onQuit func()) {
	trayOnShow = onShow
	trayOnQuit = onQuit

	systray.Run(func() {
		systray.SetIcon(trayIconBytes())
		systray.SetTitle("wtrans")
		systray.SetTooltip("wtrans - 窗口透明度控制台")

		mShow := systray.AddMenuItem("显示窗口", "显示 wtrans 窗口")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("退出", "退出 wtrans")

		mShow.Click(func() {
			if trayOnShow != nil {
				trayOnShow()
			}
		})
		mQuit.Click(func() {
			systray.Quit()
		})
	}, func() {
		if trayOnQuit != nil {
			trayOnQuit()
		}
	})
}

func stopTray() {
	systray.Quit()
}
