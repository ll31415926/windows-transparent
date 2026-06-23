**[简体中文](README.md)** | **English**

# wtrans

`wtrans` is a command-line window opacity tool. It lists desktop windows and changes opacity by process name or window class.

It is not a window manager and it does not modify the target application itself. Opacity changes usually apply only to currently open windows. After an application restart, window recreation, or desktop environment reload, rules may need to be applied again.

## Support Matrix

| Platform/backend | Mechanism | Dependencies | Notes |
|---|---|---|---|
| Windows | Win32 layered window API | None | Elevated windows may require running wtrans as administrator |
| Linux X11 | `_NET_WM_WINDOW_OPACITY` | `wmctrl`, `xprop` | Works for X11 sessions or windows manageable by X11 tools |
| Linux Sway | Sway IPC | `swaymsg` | Sets opacity by container id |
| Linux Hyprland | Hyprland window property | `hyprctl` | Sets both `alpha` and `alphainactive` |
| Linux GNOME Wayland | D-Bus + GNOME Shell extension | `gdbus`, `gnome-extensions` | Requires installing the bundled extension first |

macOS is not supported. KDE/KWin Wayland and other unlisted Wayland compositors are not currently supported because Wayland has no generic API that allows external processes to change the opacity of other applications' windows.

## Windows Compatibility

If running `wtrans.exe -h` shows a Windows loader error such as "This version is not compatible with the version of Windows you're running", the binary was rejected before `wtrans` executed.

Check these first:

1. The downloaded Release asset must match the system architecture:

| Asset name fragment | Target system |
|---|---|
| `windows_amd64` | 64-bit Intel/AMD Windows |
| `windows_386` | 32-bit Windows |
| `windows_arm64` | ARM64 Windows |

2. The Windows version must satisfy the Go runtime requirement. This project is built with Go `1.24`; according to the [Go 1.21 release notes](https://go.dev/doc/go1.21#windows), Go 1.21 and newer require at least Windows 10 or Windows Server 2016 for Windows targets. Older systems require trying an older Go toolchain locally and are not guaranteed to work.

After downloading, run:

```powershell
.\wtrans.exe version
```

The `platform` line should match the downloaded asset, for example `windows/amd64`.

## Install and Build

Requires Go `1.24` or newer. The project currently has no third-party Go module dependencies.

```bash
git clone https://github.com/ll31415926/windows-transparent.git
cd windows-transparent
```

Windows:

```powershell
go build -o wtrans.exe ./cmd/wtrans
.\wtrans.exe --help
```

Linux:

```bash
go build -o wtrans ./cmd/wtrans
./wtrans --help
```

Cross-compile Windows 64-bit:

```bash
GOOS=windows GOARCH=amd64 go build -o wtrans.exe ./cmd/wtrans
```

## Quick Use

The examples below use `wtrans` for brevity. Use `.\wtrans.exe` on Windows and `./wtrans` on Linux.

```bash
# Print help
wtrans --help

# Print version, commit, build date, and platform architecture
wtrans version

# List visible windows
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

### `help`

```bash
wtrans help
wtrans --help
wtrans -h
```

Prints command usage. `-h` and `--help` are global help entry points.

### `version`

```bash
wtrans version
wtrans --version
```

Prints version, commit, build date, Go version, and runtime platform. The Release workflow injects the tag, commit, and build date through `ldflags`.

### `list`

```bash
wtrans list [--process name]
```

Lists visible windows discovered by the active backend. Output columns:

```text
HANDLE             PID      PROCESS                  CLASS              BACKEND    TITLE
0x5B0C68           5678     Code.exe                                    windows   main.go - windows-transparent
0x4200003          5678     code                     code               x11       main.go - windows-transparent
```

`--process` is optional. When provided, only windows whose process name or window class equals the value are shown. Comparison is case-insensitive.

### `set`

```bash
wtrans set --process name --opacity 70
```

Sets opacity for matching windows.

| Argument | Required | Description |
|---|---:|---|
| `--process` | Yes | Process name or window class. On Windows this usually includes `.exe`, for example `notepad.exe` |
| `--opacity` | Yes | Opacity percentage, from `20` to `100` |

If no visible matching window is found, the command returns an error.

### `restore`

```bash
wtrans restore --process name
```

Restores matching windows to `100%` opacity. On Windows, if a window was not already a layered window, restore tries to remove the `WS_EX_LAYERED` style added by this tool. If the window already had layered attributes, existing attributes are preserved.

### `apply`

```bash
wtrans apply [--config path/to/config.json]
```

Reads a JSON config file and applies each process rule in order.

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

Config parsing uses strict JSON field validation. Unknown fields and trailing extra JSON values are rejected.

### `diagnose`

```bash
wtrans diagnose
```

Prints the current platform, graphical session, backend selection, and tool availability. On Linux, run this first when debugging backend selection, missing PATH entries, or GNOME extension state.

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

## Release Workflow

The repository includes `.github/workflows/release.yml`. Pushing a `v*` or `V*` tag automatically:

1. Runs `go test ./...`.
2. Builds these assets:
   - `windows_amd64`
   - `windows_386`
   - `windows_arm64`
   - `linux_amd64`
   - `linux_arm64`
3. Injects tag, commit, and build date into the binary.
4. Generates `SHA256SUMS.txt` and `MD5SUMS.txt`.
5. Creates a GitHub Release whose notes contain:
   - current source commit
   - commits since the previous tag
   - SHA256 checksums
   - MD5 checksums

Release example:

```bash
git tag v1.2.0
git push origin v1.2.0
```

Verify a downloaded file:

```powershell
Get-FileHash .\wtrans_v1.2.0_windows_amd64.zip -Algorithm SHA256
Get-FileHash .\wtrans_v1.2.0_windows_amd64.zip -Algorithm MD5
```

MD5 is provided only for compatibility with integrity-check workflows. Prefer SHA256 for security-sensitive verification.

## Operational Notes

- Opacity below `20` is rejected to avoid making windows difficult to interact with.
- Only currently visible windows discovered by the active backend are affected.
- Changes are not persistent configuration. They may be lost after application restart, window recreation, or desktop environment reload.
- On Windows, controlling elevated windows may require running `wtrans` elevated as well.
- Linux Wayland has no cross-desktop generic window opacity API, so support depends on the compositor and available tools.

## Development and Verification

```bash
go test ./...
go vet ./...
go build ./cmd/wtrans
```

Current tests primarily cover CLI parsing, config parsing, opacity conversion, window matching, Linux backend detection, command construction, and output parsing. Platform-specific behavior still depends on the target desktop environment and system APIs, so manual verification on the relevant platform is recommended before changing platform code.

## License

Apache License 2.0. See [LICENSE](LICENSE).
