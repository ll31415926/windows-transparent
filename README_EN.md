**[简体中文](README.md)** | **English**

# wtrans

<p align="center">
  <img alt="Go version" src="https://img.shields.io/badge/Go-1.24%2B-blue">
  <img alt="Platform" src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-green">
  <img alt="License" src="https://img.shields.io/badge/License-Apache%202.0-orange">
  <img alt="Dependencies" src="https://img.shields.io/badge/Dependencies-Zero-brightgreen">
</p>

`wtrans` is a command-line tool for listing desktop windows and changing window opacity by process name or window class. It does not modify the target application and it is not a window manager. Opacity changes usually apply only to currently open windows and must be applied again after the target application restarts.

## Scope

- List currently visible windows with handle/ID, PID, process name, window class, backend, and title.
- Match windows by process name or window class; matching is case-insensitive, but not fuzzy.
- Set matching windows to an integer opacity from `20` to `100`.
- Restore matching windows to fully opaque.
- Apply multiple process rules from a JSON config file.
- Diagnose the current Linux graphical session, selected backend, and required tools.
- Install the built-in GNOME Shell extension bridge for GNOME Wayland.

## Platform Support

| Platform/backend | Mechanism | External tools | Notes |
|---|---|---|---|
| Windows | Win32 layered window API | None | Some system windows may require an elevated process |
| Linux X11 | `_NET_WM_WINDOW_OPACITY` property | `wmctrl`, `xprop` | Works for X11 sessions or windows manageable by X11 tools |
| Linux Sway | Sway IPC | `swaymsg` | Sets opacity by container id |
| Linux Hyprland | Hyprland window properties | `hyprctl` | Sets both `alpha` and `alphainactive` |
| Linux GNOME Wayland | D-Bus + GNOME Shell extension | `gdbus`, `gnome-extensions` | Requires installing and enabling the extension first |

macOS is not supported. KDE/KWin Wayland and other unlisted Wayland compositors are also not supported at the moment because Wayland has no generic API for external processes to change the opacity of other applications' windows.

## Build

Requires Go `1.24` or newer. The project currently has no third-party Go module dependencies.

```bash
git clone https://github.com/ll31415926/windows-transparent.git
cd windows-transparent
```

Windows:

```powershell
go build -o wtrans.exe ./cmd/wtrans
.\wtrans.exe -h
```

Linux:

```bash
go build -o wtrans ./cmd/wtrans
./wtrans -h
```

## Quick Use

Windows examples use `.\wtrans.exe`; Linux examples use `./wtrans`. The commands below use `wtrans` for brevity.

```bash
# List all visible windows
wtrans list

# Filter by process name or window class
wtrans list --process Code.exe

# Set matching windows to 65% opacity
wtrans set --process notepad.exe --opacity 65

# Restore matching windows to fully opaque
wtrans restore --process notepad.exe

# Apply rules from a config file
wtrans apply --config config.json

# Print platform and backend diagnostics
wtrans diagnose
```

## Command Reference

### `list`

Lists visible windows discovered by the active backend.

```bash
wtrans list [--process name]
```

Output columns:

```text
HANDLE             PID      PROCESS                  CLASS              BACKEND    TITLE
0x5B0C68           5678     Code.exe                                    windows   main.go - windows-transparent
0x4200003          5678     code                     code               x11       main.go - windows-transparent
```

`--process` is optional. When provided, only windows whose process name or window class equals the value are shown. Comparison is case-insensitive.

### `set`

Sets opacity for matching windows.

```bash
wtrans set --process name --opacity 70
```

Arguments:

| Argument | Required | Description |
|---|---:|---|
| `--process` | Yes | Process name or window class. On Windows this usually includes `.exe`, for example `notepad.exe` |
| `--opacity` | Yes | Opacity percentage, from `20` to `100` |

If no visible matching window is found, the command returns an error.

### `restore`

Restores matching windows to `100%` opacity.

```bash
wtrans restore --process name
```

On Windows, if a window was not already a layered window, restore tries to remove the `WS_EX_LAYERED` style added by this tool. If the window already had layered attributes, existing attributes are preserved.

### `apply`

Reads a JSON config file and applies each process rule in order.

```bash
wtrans apply [--config path/to/config.json]
```

When `--config` is omitted, the system user config directory is used:

| Platform | Default path |
|---|---|
| Windows | `%APPDATA%\wtrans\config.json` |
| Linux | `~/.config/wtrans/config.json` |

Example config:

```json
{
  "rules": [
    { "process": "Code.exe", "opacity": 85 },
    { "process": "notepad.exe", "opacity": 65 },
    { "process": "firefox", "opacity": 90 },
    { "process": "gnome-terminal", "opacity": 80 }
  ]
}
```

Field constraints:

| Field | Type | Required | Description |
|---|---|---:|---|
| `rules` | array | Yes | List of opacity rules |
| `rules[].process` | string | Yes | Process name or window class |
| `rules[].opacity` | integer | Yes | Opacity percentage, from `20` to `100` |

### `diagnose`

Prints the current platform, graphical session, backend selection, and tool availability. On Linux, run this first when debugging backend selection, missing PATH entries, or GNOME extension state.

```bash
wtrans diagnose
```

Linux output example:

```text
session: wayland
desktop: sway
display:
wayland-display: wayland-1
dbus-session-bus: unix:path=/run/user/1000/bus
backend: sway
tool.wmctrl: missing
tool.xprop: missing
tool.swaymsg: found
tool.hyprctl: missing
tool.gdbus: found
tool.gnome-extensions: missing
```

### `gnome-extension`

Linux GNOME Wayland only.

```bash
wtrans gnome-extension install
wtrans gnome-extension status
```

`install` writes extension files to:

```text
~/.local/share/gnome-shell/extensions/wtrans-opacity@codex.local
```

After installing, log out and back into the graphical session, or restart GNOME Shell where supported. Then run `wtrans diagnose` and confirm `gnome-extension-bridge: true`.

## Linux Backend Selection

Auto-detection priority:

1. `WTRANS_BACKEND`, with valid values `x11`, `sway`, `hyprland`, and `gnome`.
2. `HYPRLAND_INSTANCE_SIGNATURE`.
3. `SWAYSOCK`.
4. Desktop name in `XDG_CURRENT_DESKTOP`.
5. `XDG_SESSION_TYPE`.
6. Fallback checks for available tools in PATH.

To override detection:

```bash
WTRANS_BACKEND=x11 wtrans list
```

## Operational Notes

- Opacity below `20` is rejected to avoid making windows difficult to interact with.
- Only currently visible windows discovered by the active backend are affected.
- Changes are not persistent configuration. They may be lost after application restart, window recreation, or desktop environment reload.
- On Windows, controlling elevated windows may require running `wtrans` elevated as well.
- Linux Wayland has no cross-desktop generic window opacity API, so support depends on the compositor and available tools.

## Development and Verification

```bash
go test ./...
go build ./cmd/wtrans
```

Current tests primarily cover CLI parsing, window matching, Linux backend detection, command construction, and output parsing. Platform-specific behavior still depends on the target desktop environment and system APIs, so manual verification on the relevant platform is recommended before shipping platform changes.

## License

Apache License 2.0. See [LICENSE](LICENSE).
