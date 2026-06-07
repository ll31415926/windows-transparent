//go:build linux

package window

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"windows-transparent/internal/opacity"
)

const (
	backendX11      = "x11"
	backendSway     = "sway"
	backendHyprland = "hyprland"
	backendGNOME    = "gnome"
)

var defaultRunner runner = osRunner{}

func ListVisible() ([]Window, error) {
	return listVisibleLinux(defaultRunner, envLookup)
}

func SetOpacity(win Window, value int) error {
	return setOpacityLinux(defaultRunner, win, value)
}

func Restore(win Window) error {
	return restoreLinux(defaultRunner, win)
}

type envFunc func(string) string

func envLookup(key string) string {
	return os.Getenv(key)
}

func listVisibleLinux(r runner, env envFunc) ([]Window, error) {
	backend := detectBackend(r, env)
	switch backend {
	case backendX11:
		return listX11Windows(r)
	case backendSway:
		return listSwayWindows(r)
	case backendHyprland:
		return listHyprlandWindows(r)
	case backendGNOME:
		return listGNOMEWindows(r)
	default:
		return nil, unsupportedWaylandError(env)
	}
}

func setOpacityLinux(r runner, win Window, value int) error {
	if err := opacity.Validate(value); err != nil {
		return err
	}

	switch win.Backend {
	case backendX11:
		return setX11Opacity(r, win, value)
	case backendSway:
		return setSwayOpacity(r, win, value)
	case backendHyprland:
		return setHyprlandOpacity(r, win, value)
	case backendGNOME:
		return setGNOMEOpacity(r, win, value)
	default:
		return fmt.Errorf("unsupported Linux window backend %q", win.Backend)
	}
}

func restoreLinux(r runner, win Window) error {
	switch win.Backend {
	case backendX11:
		return restoreX11Opacity(r, win)
	case backendSway:
		return setSwayOpacity(r, win, 100)
	case backendHyprland:
		return setHyprlandOpacity(r, win, 100)
	case backendGNOME:
		return setGNOMEOpacity(r, win, 100)
	default:
		return fmt.Errorf("unsupported Linux window backend %q", win.Backend)
	}
}

func detectBackend(r runner, env envFunc) string {
	sessionType := strings.ToLower(env("XDG_SESSION_TYPE"))
	desktop := strings.ToLower(env("XDG_CURRENT_DESKTOP"))
	override := strings.ToLower(env("WTRANS_BACKEND"))

	switch override {
	case backendX11, backendSway, backendHyprland, backendGNOME:
		return override
	}

	switch {
	case env("HYPRLAND_INSTANCE_SIGNATURE") != "":
		return backendHyprland
	case env("SWAYSOCK") != "":
		return backendSway
	case strings.Contains(desktop, "hyprland"):
		return backendHyprland
	case strings.Contains(desktop, "sway"):
		return backendSway
	case strings.Contains(desktop, "gnome") && gnomeExtensionAvailable(r):
		return backendGNOME
	case sessionType == backendX11:
		return backendX11
	}

	if _, err := r.LookPath("hyprctl"); err == nil && sessionType == "wayland" {
		return backendHyprland
	}
	if _, err := r.LookPath("swaymsg"); err == nil && sessionType == "wayland" {
		return backendSway
	}
	if strings.Contains(desktop, "gnome") && gnomeExtensionAvailable(r) {
		return backendGNOME
	}

	if sessionType != "wayland" && env("DISPLAY") != "" {
		return backendX11
	}

	return ""
}

func unsupportedWaylandError(env envFunc) error {
	sessionType := env("XDG_SESSION_TYPE")
	desktop := env("XDG_CURRENT_DESKTOP")
	if sessionType == "tty" || (desktop == "" && env("DISPLAY") == "" && env("WAYLAND_DISPLAY") == "") {
		return fmt.Errorf("wtrans is running outside a graphical desktop session (XDG_SESSION_TYPE=%q). Open a terminal inside your GNOME/X11/Sway/Hyprland desktop session and run wtrans there", sessionType)
	}
	if strings.Contains(strings.ToLower(desktop), "gnome") {
		return fmt.Errorf("GNOME Wayland requires the wtrans GNOME Shell extension. Run `wtrans gnome-extension install`, log out and back in, then retry. If you only need XWayland windows, install wmctrl/xprop and run with WTRANS_BACKEND=x11")
	}

	if strings.EqualFold(sessionType, "wayland") {
		return fmt.Errorf("Wayland compositor %q is not supported by wtrans; supported Wayland backends are Sway, Hyprland, and GNOME with the wtrans Shell extension. KDE/KWin requires compositor rules or scripts because Wayland has no generic API for changing other applications' opacity", desktop)
	}

	return errors.New("no supported Linux window backend detected; use an X11 session with wmctrl and xprop, or a supported Wayland compositor such as Sway, Hyprland, or GNOME with the wtrans Shell extension")
}

func Diagnose(w io.Writer) error {
	env := envLookup
	backend := detectBackend(defaultRunner, env)
	if backend == "" {
		backend = "unsupported"
	}

	fmt.Fprintf(w, "session: %s\n", env("XDG_SESSION_TYPE"))
	fmt.Fprintf(w, "desktop: %s\n", env("XDG_CURRENT_DESKTOP"))
	fmt.Fprintf(w, "display: %s\n", env("DISPLAY"))
	fmt.Fprintf(w, "wayland-display: %s\n", env("WAYLAND_DISPLAY"))
	fmt.Fprintf(w, "dbus-session-bus: %s\n", env("DBUS_SESSION_BUS_ADDRESS"))
	fmt.Fprintf(w, "backend: %s\n", backend)
	for _, tool := range []string{"wmctrl", "xprop", "swaymsg", "hyprctl", "gdbus", "gnome-extensions"} {
		status := "missing"
		if _, err := defaultRunner.LookPath(tool); err == nil {
			status = "found"
		}
		fmt.Fprintf(w, "tool.%s: %s\n", tool, status)
	}
	if strings.Contains(strings.ToLower(env("XDG_CURRENT_DESKTOP")), "gnome") {
		if _, version, err := detectGNOMEShellMajor(defaultRunner); err == nil {
			fmt.Fprintf(w, "gnome-shell: %s\n", version)
		} else {
			fmt.Fprintf(w, "gnome-shell: unknown (%v)\n", err)
		}
		fmt.Fprintf(w, "gnome-extension-installed: %v\n", gnomeExtensionInstalled(defaultRunner))
		fmt.Fprintf(w, "gnome-extension-enabled: %v\n", gnomeExtensionEnabled(defaultRunner))
		fmt.Fprintf(w, "gnome-extension-bridge: %v\n", gnomeExtensionAvailable(defaultRunner))
	}
	if env("XDG_SESSION_TYPE") == "tty" || (env("DISPLAY") == "" && env("WAYLAND_DISPLAY") == "") {
		fmt.Fprintln(w, "hint: this process is not attached to a graphical desktop session; launch a terminal inside GNOME/X11/Sway/Hyprland and run wtrans there")
	}

	return nil
}

func requireTool(r runner, name string) error {
	if _, err := r.LookPath(name); err != nil {
		return fmt.Errorf("required tool %q was not found in PATH", name)
	}

	return nil
}

func alphaHex(value int) string {
	x11Opacity := uint32(uint64(value) * 0xffffffff / 100)
	return fmt.Sprintf("0x%08x", x11Opacity)
}

func opacityFloat(value int) string {
	if value == 100 {
		return "1"
	}

	text := strconv.FormatFloat(float64(value)/100, 'f', 2, 64)
	return strings.TrimRight(strings.TrimRight(text, "0"), ".")
}
