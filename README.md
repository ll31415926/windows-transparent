**简体中文** | **[English](README_EN.md)**

# wtrans

<p align="center">
  <img alt="Go 版本" src="https://img.shields.io/badge/Go-1.24%2B-blue">
  <img alt="平台" src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-green">
  <img alt="许可证" src="https://img.shields.io/badge/License-Apache%202.0-orange">
  <img alt="依赖" src="https://img.shields.io/badge/Dependencies-Zero-brightgreen">
</p>

`wtrans` 是一个命令行工具，用于列出桌面窗口，并按进程名或窗口类名调整窗口不透明度。它不修改目标应用本身，也不是窗口管理器；透明度设置通常只对当前已打开的窗口生效，目标应用重启后需要重新应用。

## 能力范围

- 列出当前可见窗口，输出窗口句柄/ID、PID、进程名、窗口类名、后端和标题。
- 按进程名或窗口类名匹配窗口；匹配大小写不敏感，但不是模糊搜索。
- 将匹配窗口的不透明度设置为 `20` 到 `100` 之间的整数百分比。
- 将匹配窗口恢复为完全不透明。
- 从 JSON 配置文件批量应用多条进程规则。
- 在 Linux 上诊断当前图形会话、检测到的后端和必要工具状态。
- 在 GNOME Wayland 上安装本项目内置的 GNOME Shell 扩展桥接层。

## 支持平台

| 平台/后端 | 实现方式 | 外部工具 | 说明 |
|---|---|---|---|
| Windows | Win32 layered window API | 无 | 部分系统窗口可能需要以管理员权限运行 |
| Linux X11 | `_NET_WM_WINDOW_OPACITY` 属性 | `wmctrl`, `xprop` | 适用于 X11 会话或可由 X11 工具管理的窗口 |
| Linux Sway | Sway IPC | `swaymsg` | 通过 container id 设置 opacity |
| Linux Hyprland | Hyprland window properties | `hyprctl` | 同时设置 `alpha` 和 `alphainactive` |
| Linux GNOME Wayland | D-Bus + GNOME Shell 扩展 | `gdbus`, `gnome-extensions` | 需要先安装并启用扩展 |

不支持 macOS。KDE/KWin Wayland 和其他未列出的 Wayland 合成器目前也不支持，因为 Wayland 没有通用 API 允许外部进程修改其他应用窗口的不透明度。

## 构建

要求 Go `1.24` 或更高版本。项目当前没有第三方 Go 模块依赖。

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

## 快速使用

Windows 示例使用 `.\wtrans.exe`，Linux 示例使用 `./wtrans`。下面统一写作 `wtrans`。

```bash
# 列出所有可见窗口
wtrans list

# 按进程名或窗口类名过滤
wtrans list --process Code.exe

# 将匹配窗口设置为 65% 不透明
wtrans set --process notepad.exe --opacity 65

# 恢复匹配窗口为完全不透明
wtrans restore --process notepad.exe

# 从配置文件批量应用规则
wtrans apply --config config.json

# 输出平台和后端诊断信息
wtrans diagnose
```

## 命令参考

### `list`

列出当前后端能发现的可见窗口。

```bash
wtrans list [--process name]
```

输出列：

```text
HANDLE             PID      PROCESS                  CLASS              BACKEND    TITLE
0x5B0C68           5678     Code.exe                                    windows   main.go - windows-transparent
0x4200003          5678     code                     code               x11       main.go - windows-transparent
```

`--process` 可选。传入后只显示进程名或窗口类名与该值相等的窗口，比较时忽略大小写。

### `set`

设置匹配窗口的不透明度。

```bash
wtrans set --process name --opacity 70
```

参数：

| 参数 | 必填 | 说明 |
|---|---:|---|
| `--process` | 是 | 进程名或窗口类名。Windows 通常需要包含 `.exe`，例如 `notepad.exe` |
| `--opacity` | 是 | 不透明度百分比，范围 `20` 到 `100` |

如果没有找到匹配的可见窗口，命令会返回错误。

### `restore`

恢复匹配窗口为 `100%` 不透明。

```bash
wtrans restore --process name
```

Windows 上，如果窗口原本不是 layered window，恢复时会尽量移除本工具添加的 `WS_EX_LAYERED` 样式；如果窗口本身已有相关属性，则保留原属性。

### `apply`

从 JSON 配置读取规则，并按规则逐个设置匹配窗口的不透明度。

```bash
wtrans apply [--config path/to/config.json]
```

未传 `--config` 时使用系统用户配置目录：

| 平台 | 默认路径 |
|---|---|
| Windows | `%APPDATA%\wtrans\config.json` |
| Linux | `~/.config/wtrans/config.json` |

配置示例：

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

字段约束：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| `rules` | array | 是 | 透明度规则列表 |
| `rules[].process` | string | 是 | 进程名或窗口类名 |
| `rules[].opacity` | integer | 是 | 不透明度百分比，范围 `20` 到 `100` |

### `diagnose`

输出当前平台、图形会话、后端选择和工具可用性。Linux 下排查后端选择、PATH 缺失或 GNOME 扩展状态时优先运行该命令。

```bash
wtrans diagnose
```

Linux 输出示例：

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

仅用于 Linux GNOME Wayland。

```bash
wtrans gnome-extension install
wtrans gnome-extension status
```

`install` 会把扩展文件写入：

```text
~/.local/share/gnome-shell/extensions/wtrans-opacity@codex.local
```

安装后需要重新登录图形会话，或在支持的 GNOME Shell 会话中重启 Shell。之后运行 `wtrans diagnose` 确认 `gnome-extension-bridge: true`。

## Linux 后端选择

自动检测优先级：

1. `WTRANS_BACKEND` 环境变量，合法值为 `x11`、`sway`、`hyprland`、`gnome`。
2. `HYPRLAND_INSTANCE_SIGNATURE`。
3. `SWAYSOCK`。
4. `XDG_CURRENT_DESKTOP` 中的桌面名称。
5. `XDG_SESSION_TYPE`。
6. PATH 中可用工具的回退检测。

需要覆盖自动检测时：

```bash
WTRANS_BACKEND=x11 wtrans list
```

## 使用注意

- 透明度值低于 `20` 会被拒绝，避免窗口难以交互。
- 只作用于当前能被后端枚举到的可见窗口。
- 设置不是持久化配置；应用重启、窗口重建或桌面环境重载后可能失效。
- Windows 上控制高权限窗口时，`wtrans` 进程本身可能也需要高权限。
- Linux Wayland 没有跨桌面环境的通用窗口透明度控制 API，因此支持范围取决于具体合成器和可用工具。

## 开发与验证

```bash
go test ./...
go build ./cmd/wtrans
```

当前测试主要覆盖 CLI 参数解析、窗口匹配、Linux 后端检测、命令构造和输出解析。平台相关行为仍依赖目标桌面环境和系统 API，提交变更前建议在对应平台做一次手工验证。

## License

Apache License 2.0. 详见 [LICENSE](LICENSE)。
