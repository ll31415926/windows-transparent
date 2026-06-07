//go:build linux

package window

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	gnomeExtensionUUID       = "wtrans-opacity@codex.local"
	gnomeExtensionObjectPath = "/org/gnome/Shell/Extensions/WTrans"
	gnomeExtensionInterface  = "org.gnome.Shell.Extensions.WTrans"
)

type gnomeWindow struct {
	ID      string `json:"id"`
	PID     uint32 `json:"pid"`
	Process string `json:"process"`
	Class   string `json:"class"`
	Title   string `json:"title"`
}

func gnomeExtensionAvailable(r runner) bool {
	if _, err := r.LookPath("gdbus"); err != nil {
		return false
	}

	output, err := r.Run(
		"gdbus",
		"call",
		"--session",
		"--dest", "org.gnome.Shell",
		"--object-path", gnomeExtensionObjectPath,
		"--method", "org.freedesktop.DBus.Introspectable.Introspect",
	)
	return err == nil && strings.Contains(output, gnomeExtensionInterface)
}

func listGNOMEWindows(r runner) ([]Window, error) {
	if err := requireTool(r, "gdbus"); err != nil {
		return nil, err
	}

	output, err := r.Run(
		"gdbus",
		"call",
		"--session",
		"--dest", "org.gnome.Shell",
		"--object-path", gnomeExtensionObjectPath,
		"--method", gnomeExtensionInterface+".ListWindows",
	)
	if err != nil {
		if gnomeBridgeMissing(err) {
			return nil, gnomeBridgeMissingMessage(err)
		}
		return nil, fmt.Errorf("call GNOME extension ListWindows: %w", err)
	}

	payload, err := parseGDBusString(output)
	if err != nil {
		return nil, fmt.Errorf("parse GNOME extension response: %w", err)
	}

	var items []gnomeWindow
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return nil, fmt.Errorf("parse GNOME windows JSON: %w", err)
	}

	windows := make([]Window, 0, len(items))
	for _, item := range items {
		windows = append(windows, Window{
			ID:      item.ID,
			PID:     item.PID,
			Process: item.Process,
			Class:   item.Class,
			Title:   item.Title,
			Visible: true,
			Backend: backendGNOME,
		})
	}

	return windows, nil
}

func setGNOMEOpacity(r runner, win Window, value int) error {
	if err := requireTool(r, "gdbus"); err != nil {
		return err
	}

	id := linuxWindowID(win)
	if id == "" {
		return fmt.Errorf("missing GNOME window id for %s", win.Process)
	}

	_, err := r.Run(
		"gdbus",
		"call",
		"--session",
		"--dest", "org.gnome.Shell",
		"--object-path", gnomeExtensionObjectPath,
		"--method", gnomeExtensionInterface+".SetOpacity",
		id,
		strconv.Itoa(value),
	)
	if err != nil && gnomeBridgeMissing(err) {
		return gnomeBridgeMissingMessage(err)
	}
	return err
}

func InstallGNOMEExtension(w io.Writer) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}

	extensionDir := filepath.Join(home, ".local", "share", "gnome-shell", "extensions", gnomeExtensionUUID)
	if err := os.MkdirAll(extensionDir, 0o755); err != nil {
		return fmt.Errorf("create extension directory: %w", err)
	}

	major, version, versionErr := detectGNOMEShellMajor(defaultRunner)
	files := map[string]string{
		"metadata.json": gnomeExtensionMetadataForMajor(major),
		"extension.js":  gnomeExtensionJSForMajor(major),
	}
	for name, contents := range files {
		path := filepath.Join(extensionDir, name)
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	if versionErr == nil {
		fmt.Fprintf(w, "detected %s\n", version)
	} else {
		fmt.Fprintf(w, "could not detect GNOME Shell version; installed the GNOME 45+ extension template: %v\n", versionErr)
	}
	fmt.Fprintf(w, "installed GNOME Shell extension (%s template) to %s\n", gnomeExtensionTemplateName(major), extensionDir)
	if os.Getenv("XDG_SESSION_TYPE") == "tty" || os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		fmt.Fprintln(w, "not enabling now because this command is not running inside the graphical GNOME session")
	} else if _, err := defaultRunner.LookPath("gnome-extensions"); err == nil {
		if gnomeExtensionEnabled(defaultRunner) {
			if _, disableErr := defaultRunner.Run("gnome-extensions", "disable", gnomeExtensionUUID); disableErr == nil {
				fmt.Fprintf(w, "disabled %s to reload updated extension files\n", gnomeExtensionUUID)
			} else {
				fmt.Fprintf(w, "could not disable existing extension before reload: %v\n", disableErr)
			}
		}
		if _, enableErr := defaultRunner.Run("gnome-extensions", "enable", gnomeExtensionUUID); enableErr == nil {
			fmt.Fprintf(w, "enabled %s\n", gnomeExtensionUUID)
		} else {
			fmt.Fprintf(w, "extension was installed, but GNOME may need a logout/login before it can be enabled: %v\n", enableErr)
		}
	}
	fmt.Fprintln(w, "on GNOME Wayland, log out and back in or restart GNOME Shell, then run `wtrans diagnose` from a terminal inside the graphical session")
	fmt.Fprintf(w, "if `wtrans list` still reports a missing bridge, run `gnome-extensions info %s` and check GNOME Shell logs\n", gnomeExtensionUUID)

	return nil
}

func GNOMEExtensionStatus(w io.Writer) error {
	if _, version, err := detectGNOMEShellMajor(defaultRunner); err == nil {
		fmt.Fprintf(w, "gnome-shell: %s\n", version)
	} else {
		fmt.Fprintf(w, "gnome-shell: unknown (%v)\n", err)
	}

	fmt.Fprintf(w, "installed: %v\n", gnomeExtensionInstalled(defaultRunner))
	fmt.Fprintf(w, "enabled: %v\n", gnomeExtensionEnabled(defaultRunner))
	bridge := gnomeExtensionAvailable(defaultRunner)
	fmt.Fprintf(w, "bridge: %v\n", bridge)
	if !bridge {
		fmt.Fprintf(w, "hint: if installed and enabled are true but bridge is false, log out and back in, then check `gnome-extensions info %s`\n", gnomeExtensionUUID)
	}

	return nil
}

func gnomeExtensionInstalled(r runner) bool {
	return gnomeExtensionsListContains(r)
}

func gnomeExtensionEnabled(r runner) bool {
	return gnomeExtensionsListContains(r, "--enabled")
}

func gnomeExtensionsListContains(r runner, args ...string) bool {
	if _, err := r.LookPath("gnome-extensions"); err != nil {
		return false
	}

	cmdArgs := append([]string{"list"}, args...)
	output, err := r.Run("gnome-extensions", cmdArgs...)
	if err != nil {
		return false
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == gnomeExtensionUUID {
			return true
		}
	}

	return false
}

func detectGNOMEShellMajor(r runner) (int, string, error) {
	if _, err := r.LookPath("gnome-shell"); err != nil {
		return 0, "", fmt.Errorf("find gnome-shell: %w", err)
	}

	output, err := r.Run("gnome-shell", "--version")
	if err != nil {
		return 0, "", fmt.Errorf("run gnome-shell --version: %w", err)
	}

	version := strings.TrimSpace(output)
	major, ok := parseGNOMEShellMajor(version)
	if !ok {
		return 0, version, fmt.Errorf("parse GNOME Shell version from %q", version)
	}

	return major, version, nil
}

func parseGNOMEShellMajor(output string) (int, bool) {
	for _, field := range strings.Fields(output) {
		field = strings.Trim(field, " ,")
		if field == "" {
			continue
		}

		majorText := strings.SplitN(field, ".", 2)[0]
		major, err := strconv.Atoi(majorText)
		if err == nil && major > 0 {
			return major, true
		}
	}

	return 0, false
}

func gnomeExtensionTemplateName(major int) string {
	if major > 0 && major < 45 {
		return "GNOME 42-44 legacy"
	}

	return "GNOME 45+ ESM"
}

func gnomeExtensionJSForMajor(major int) string {
	if major > 0 && major < 45 {
		return gnomeExtensionLegacyJS
	}

	return gnomeExtensionESMJS
}

func gnomeExtensionMetadataForMajor(major int) string {
	shellVersions := []string{"45", "46", "47", "48", "49", "50"}
	if major > 0 && major < 45 {
		shellVersions = []string{"42", "43", "44"}
	} else if major > 50 {
		shellVersions = append(shellVersions, strconv.Itoa(major))
	}

	data, err := json.MarshalIndent(struct {
		UUID         string   `json:"uuid"`
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		ShellVersion []string `json:"shell-version"`
	}{
		UUID:         gnomeExtensionUUID,
		Name:         "wtrans opacity bridge",
		Description:  "D-Bus bridge used by wtrans to change GNOME Wayland window opacity.",
		ShellVersion: shellVersions,
	}, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(data) + "\n"
}

func gnomeBridgeMissing(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "object does not exist") ||
		strings.Contains(text, "unknownmethod") ||
		strings.Contains(text, "unknown method")
}

func gnomeBridgeMissingMessage(err error) error {
	return fmt.Errorf("GNOME Shell extension D-Bus bridge is not loaded at %s. Run `wtrans gnome-extension install`, log out and back in, then retry. If it still fails, run `gnome-extensions info %s` and check GNOME Shell logs: %w", gnomeExtensionObjectPath, gnomeExtensionUUID, err)
}

func parseGDBusString(output string) (string, error) {
	start := strings.Index(output, "'")
	if start < 0 {
		return "", fmt.Errorf("missing opening quote in %q", strings.TrimSpace(output))
	}

	var b strings.Builder
	escaped := false
	for _, r := range output[start+1:] {
		if escaped {
			switch r {
			case 'n':
				b.WriteRune('\n')
			case 't':
				b.WriteRune('\t')
			case 'r':
				b.WriteRune('\r')
			default:
				b.WriteRune(r)
			}
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}
		if r == '\'' {
			return b.String(), nil
		}

		b.WriteRune(r)
	}

	return "", fmt.Errorf("missing closing quote in %q", strings.TrimSpace(output))
}

const gnomeExtensionESMJS = `import Gio from 'gi://Gio';
import {Extension} from 'resource:///org/gnome/shell/extensions/extension.js';

const DBUS_XML = ` + "`" + `<node>
  <interface name="org.gnome.Shell.Extensions.WTrans">
    <method name="ListWindows">
      <arg type="s" name="windows" direction="out"/>
    </method>
    <method name="SetOpacity">
      <arg type="s" name="id" direction="in"/>
      <arg type="u" name="opacity" direction="in"/>
      <arg type="b" name="ok" direction="out"/>
    </method>
  </interface>
</node>` + "`" + `;

export default class WTransOpacityExtension extends Extension {
  enable() {
    this._nextId = 1;
    this._weakIds = new WeakMap();
    this._actorsById = new Map();
    this._dbus = Gio.DBusExportedObject.wrapJSObject(DBUS_XML, this);
    this._dbus.export(Gio.DBus.session, '/org/gnome/Shell/Extensions/WTrans');
  }

  disable() {
    for (const actor of this._actorsById.values()) {
      if (actor)
        actor.opacity = 255;
    }
    this._actorsById.clear();
    if (this._dbus) {
      this._dbus.unexport();
      this._dbus = null;
    }
  }

  ListWindows() {
    this._actorsById.clear();
    const result = [];

    for (const actor of global.get_window_actors()) {
      const win = actor.meta_window || (typeof actor.get_meta_window === 'function' ? actor.get_meta_window() : null);
      if (!win)
        continue;

      const id = this._idForWindow(win);
      this._actorsById.set(id, actor);
      const appId = this._callString(win, 'get_gtk_application_id');
      const wmClass = this._callString(win, 'get_wm_class') || this._callString(win, 'get_wm_class_instance');
      const process = appId || wmClass;

      result.push({
        id,
        pid: this._callNumber(win, 'get_pid'),
        process,
        class: wmClass || appId,
        title: this._callString(win, 'get_title')
      });
    }

    return JSON.stringify(result);
  }

  SetOpacity(id, opacity) {
    const actor = this._actorsById.get(id);
    if (!actor)
      throw new Error(` + "`" + `unknown window id ${id}; run wtrans list again` + "`" + `);

    const clamped = Math.max(20, Math.min(100, opacity));
    actor.opacity = Math.round(clamped * 255 / 100);
    actor.queue_redraw();
    return true;
  }

  _idForWindow(win) {
    if (typeof win.get_stable_sequence === 'function')
      return String(win.get_stable_sequence());

    if (!this._weakIds.has(win))
      this._weakIds.set(win, String(this._nextId++));
    return this._weakIds.get(win);
  }

  _callString(obj, method) {
    if (typeof obj[method] !== 'function')
      return '';
    const value = obj[method]();
    return value === null || value === undefined ? '' : String(value);
  }

  _callNumber(obj, method) {
    if (typeof obj[method] !== 'function')
      return 0;
    const value = obj[method]();
    return Number.isFinite(value) ? value : 0;
  }
}
`

const gnomeExtensionLegacyJS = `const Gio = imports.gi.Gio;

const DBUS_XML = ` + "`" + `<node>
  <interface name="org.gnome.Shell.Extensions.WTrans">
    <method name="ListWindows">
      <arg type="s" name="windows" direction="out"/>
    </method>
    <method name="SetOpacity">
      <arg type="s" name="id" direction="in"/>
      <arg type="u" name="opacity" direction="in"/>
      <arg type="b" name="ok" direction="out"/>
    </method>
  </interface>
</node>` + "`" + `;

class WTransOpacityExtension {
  enable() {
    this._nextId = 1;
    this._weakIds = new WeakMap();
    this._actorsById = new Map();
    this._dbus = Gio.DBusExportedObject.wrapJSObject(DBUS_XML, this);
    this._dbus.export(Gio.DBus.session, '/org/gnome/Shell/Extensions/WTrans');
  }

  disable() {
    for (const actor of this._actorsById.values()) {
      if (actor)
        actor.opacity = 255;
    }
    this._actorsById.clear();
    if (this._dbus) {
      this._dbus.unexport();
      this._dbus = null;
    }
  }

  ListWindows() {
    this._actorsById.clear();
    const result = [];

    for (const actor of global.get_window_actors()) {
      const win = actor.meta_window || (typeof actor.get_meta_window === 'function' ? actor.get_meta_window() : null);
      if (!win)
        continue;

      const id = this._idForWindow(win);
      this._actorsById.set(id, actor);
      const appId = this._callString(win, 'get_gtk_application_id');
      const wmClass = this._callString(win, 'get_wm_class') || this._callString(win, 'get_wm_class_instance');
      const process = appId || wmClass;

      result.push({
        id,
        pid: this._callNumber(win, 'get_pid'),
        process,
        class: wmClass || appId,
        title: this._callString(win, 'get_title')
      });
    }

    return JSON.stringify(result);
  }

  SetOpacity(id, opacity) {
    const actor = this._actorsById.get(id);
    if (!actor)
      throw new Error('unknown window id ' + id + '; run wtrans list again');

    const clamped = Math.max(20, Math.min(100, opacity));
    actor.opacity = Math.round(clamped * 255 / 100);
    actor.queue_redraw();
    return true;
  }

  _idForWindow(win) {
    if (typeof win.get_stable_sequence === 'function')
      return String(win.get_stable_sequence());

    if (!this._weakIds.has(win))
      this._weakIds.set(win, String(this._nextId++));
    return this._weakIds.get(win);
  }

  _callString(obj, method) {
    if (typeof obj[method] !== 'function')
      return '';
    const value = obj[method]();
    return value === null || value === undefined ? '' : String(value);
  }

  _callNumber(obj, method) {
    if (typeof obj[method] !== 'function')
      return 0;
    const value = obj[method]();
    return Number.isFinite(value) ? value : 0;
  }
}

function init() {
  return new WTransOpacityExtension();
}
`
