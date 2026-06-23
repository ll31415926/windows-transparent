**简体中文** | **[English](README_EN.md)**

# wtrans

`wtrans` 是一个命令行窗口透明度工具。它可以列出当前桌面窗口，并按进程名或窗口类名调整窗口不透明度。

它不是窗口管理器，也不会修改目标应用本身。透明度设置通常只作用于当前已打开的窗口；目标应用重启、窗口重建或桌面环境重载后，可能需要重新应用。

## 支持范围

| 平台/后端 | 实现方式 | 依赖 | 说明 |
|---|---|---|---|
| Windows | Win32 layered window API | 无 | 高权限窗口可能需要以管理员权限运行 |
| Linux X11 | `_NET_WM_WINDOW_OPACITY` | `wmctrl`, `xprop` | 适用于 X11 会话或可由 X11 工具管理的窗口 |
| Linux Sway | Sway IPC | `swaymsg` | 按 container id 设置 opacity |
| Linux Hyprland | Hyprland window property | `hyprctl` | 同时设置 `alpha` 和 `alphainactive` |
| Linux GNOME Wayland | D-Bus + GNOME Shell 扩展 | `gdbus`, `gnome-extensions` | 需要先安装本项目提供的扩展 |

不支持 macOS。KDE/KWin Wayland 和其他未列出的 Wayland 合成器目前也不支持，因为 Wayland 没有通用 API 允许外部进程修改其他应用窗口的不透明度。

## Windows 兼容性

如果运行 `wtrans.exe -h` 时看到类似“该版本与正在运行的 Windows 版本不兼容”的系统弹窗或控制台报错，通常不是 `wtrans` 逻辑执行失败，而是 Windows 加载器拒绝了二进制。

优先检查两件事：

1. 下载的 Release 产物是否匹配系统架构：

| 产物名片段 | 适用系统 |
|---|---|
| `windows_amd64` | 64 位 Intel/AMD Windows |
| `windows_386` | 32 位 Windows |
| `windows_arm64` | ARM64 Windows |

2. Windows 版本是否满足 Go 运行时要求。当前项目使用 Go `1.24` 构建；根据 [Go 1.21 release notes](https://go.dev/doc/go1.21#windows)，Go 1.21 起 Windows 目标至少需要 Windows 10 或 Windows Server 2016。更旧系统需要使用旧 Go 工具链自行尝试构建，项目不保证兼容。

下载后可运行：

```powershell
.\wtrans.exe version
```

输出中的 `platform` 应与下载产物架构一致，例如 `windows/amd64`。

## 安装与构建

要求 Go `1.24` 或更高版本。项目当前没有第三方 Go 模块依赖。

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

交叉构建 Windows 64 位：

```bash
GOOS=windows GOARCH=amd64 go build -o wtrans.exe ./cmd/wtrans
```

## 快速使用

以下示例统一写作 `wtrans`。Windows 下将其替换为 `.\wtrans.exe`，Linux 下替换为 `./wtrans`。

```bash
# 查看帮助
wtrans --help

# 查看版本、commit、构建时间和平台架构
wtrans version

# 列出当前可见窗口
wtrans list

# 按进程名或窗口类名过滤
wtrans list --process Code.exe

# 设置匹配窗口为 65% 不透明
wtrans set --process notepad.exe --opacity 65

# 恢复匹配窗口为完全不透明
wtrans restore --process notepad.exe

# 从配置文件批量应用规则
wtrans apply --config config.json

# 输出平台和后端诊断信息
wtrans diagnose
```

## 命令参考

### `help`

```bash
wtrans help
wtrans --help
wtrans -h
```

打印命令用法。`-h` 和 `--help` 是全局帮助入口。

### `version`

```bash
wtrans version
wtrans --version
```

打印版本、commit、构建时间、Go 版本和运行平台。Release 工作流会通过 `ldflags` 写入 tag、commit 和构建时间。

### `list`

```bash
wtrans list [--process name]
```

列出当前后端能发现的可见窗口。输出列为：

```text
HANDLE             PID      PROCESS                  CLASS              BACKEND    TITLE
0x5B0C68           5678     Code.exe                                    windows   main.go - windows-transparent
0x4200003          5678     code                     code               x11       main.go - windows-transparent
```

`--process` 可选。传入后只显示进程名或窗口类名与该值相等的窗口，比较时忽略大小写。

### `set`

```bash
wtrans set --process name --opacity 70
```

设置匹配窗口的不透明度。

| 参数 | 必填 | 说明 |
|---|---:|---|
| `--process` | 是 | 进程名或窗口类名。Windows 通常需要包含 `.exe`，例如 `notepad.exe` |
| `--opacity` | 是 | 不透明度百分比，范围 `20` 到 `100` |

如果没有找到匹配的可见窗口，命令会返回错误。

### `restore`

```bash
wtrans restore --process name
```

恢复匹配窗口为 `100%` 不透明。Windows 上，如果窗口原本不是 layered window，恢复时会尽量移除本工具添加的 `WS_EX_LAYERED` 样式；如果窗口本身已有相关属性，则保留原属性。

### `apply`

```bash
wtrans apply [--config path/to/config.json]
```

从 JSON 配置读取规则，并按规则逐个设置匹配窗口的不透明度。

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

配置解析使用严格 JSON 字段校验；未知字段和尾随额外 JSON 值都会报错。

### `diagnose`

```bash
wtrans diagnose
```

输出当前平台、图形会话、后端选择和工具可用性。Linux 下排查后端选择、PATH 缺失或 GNOME 扩展状态时优先运行该命令。

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

## Release 工作流

仓库包含 `.github/workflows/release.yml`。推送 `v*` 或 `V*` tag 时会自动：

1. 运行 `go test ./...`。
2. 构建以下产物：
   - `windows_amd64`
   - `windows_386`
   - `windows_arm64`
   - `linux_amd64`
   - `linux_arm64`
3. 为二进制注入 tag、commit 和构建时间。
4. 生成 `SHA256SUMS.txt` 和 `MD5SUMS.txt`。
5. 创建 GitHub Release，Release notes 包含：
   - 当前 source commit
   - 相对上一个 tag 的 commit 列表
   - SHA256 校验值
   - MD5 校验值

发布示例：

```bash
git tag v1.2.0
git push origin v1.2.0
```

校验下载文件：

```powershell
Get-FileHash .\wtrans_v1.2.0_windows_amd64.zip -Algorithm SHA256
Get-FileHash .\wtrans_v1.2.0_windows_amd64.zip -Algorithm MD5
```

MD5 仅用于兼容用户的完整性检查，不应作为安全校验依据；安全校验优先使用 SHA256。

## 使用注意

- 透明度值低于 `20` 会被拒绝，避免窗口难以交互。
- 只作用于当前能被后端枚举到的可见窗口。
- 设置不是持久化配置；应用重启、窗口重建或桌面环境重载后可能失效。
- Windows 上控制高权限窗口时，`wtrans` 进程本身可能也需要高权限。
- Linux Wayland 没有跨桌面环境的通用窗口透明度控制 API，因此支持范围取决于具体合成器和可用工具。

## 免责声明

- 仅在你拥有合法访问和操作权限的系统上使用本工具。
- 不建议对关键系统窗口或正在执行重要任务的窗口设置过低透明度。
- 本工具不收集、不上传用户数据；所有操作均在本机完成。
- 窗口透明度修改可能受系统权限、桌面环境、窗口管理器或目标应用行为影响，效果不保证一致。
- 使用本工具造成的误操作、显示异常或工作中断，由使用者自行承担风险。

## 开发验证

```bash
go test ./...
go vet ./...
go build ./cmd/wtrans
```

当前测试主要覆盖 CLI 参数解析、配置解析、透明度转换、窗口匹配、Linux 后端检测、命令构造和输出解析。平台相关行为仍依赖目标桌面环境和系统 API，变更平台代码前应在对应系统做手工验证。

## License

Apache License 2.0. 详见 [LICENSE](LICENSE)。
