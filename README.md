**[English](README_EN.md)** | **简体中文**

# 🪟 wtrans 窗口透明度工具


## 📸 界面预览

![wtrans GUI 控制台](docs/assets/screenshots/gui-overview.png)

wtrans GUI 控制台：左侧可见窗口列表，右侧透明度滑块和持久化规则
<!-- runtime-screenshots:end -->


<p align="center">
  <img alt="Go版本" src="https://img.shields.io/badge/Go-1.24%2B-blue">
  <img alt="平台支持" src="https://img.shields.io/badge/平台-Windows%20%7C%20Linux-green">
  <img alt="开源协议" src="https://img.shields.io/badge/许可-Apache%202.0-orange">
  <img alt="零依赖" src="https://img.shields.io/badge/依赖-零依赖-brightgreen">
</p>

> 跨平台窗口透明度调节工具，通过 Win32 API 和 X11/Wayland 支持控制任意应用窗口的不透明度

wtrans 是一个纯 Go 实现的命令行窗口透明度管理工具，无需任何第三方依赖。在 Windows 上通过 Win32 API 直接操作，在 Linux 上支持 X11、Sway、Hyprland 和 GNOME Wayland 四种后端，实现对任意可见窗口的透明度设置、恢复和批量管理，适合需要多窗口协同工作的开发者和高级用户。

<br/>

## ✨ 核心功能

- **🪟 窗口透明度控制**：设置任意可见窗口的不透明度（20%-100%），支持精确调节
- **📋 窗口列表查看**：列出所有可见窗口，显示窗口ID、PID、进程名、标题和后端信息
- **🔄 一键恢复**：将窗口恢复为完全不透明，并智能移除分层样式标记
- **📦 批量配置应用**：通过 JSON 配置文件批量设置多个进程的透明度规则
- **🔍 进程名过滤**：支持按进程名和窗口类名筛选，匹配时不区分大小写
- **⚙️ 智能样式管理**：自动处理窗口扩展样式，恢复时保留原有属性
- **🛡️ 非破坏性操作**：恢复透明度时检查原有属性，避免覆盖窗口原始配置
- **🖥️ 多后端支持**（Linux）：自动检测 X11、Sway、Hyprland、GNOME Wayland 桌面环境
- **🔧 诊断命令**：一键查看当前会话信息、后端状态和工具依赖
- **📦 零外部依赖**：完全使用 Go 标准库，无任何第三方模块依赖
- **🧪 完整测试覆盖**：所有核心包均有单元测试，确保功能稳定可靠

<br/>

## 🚀 快速开始

### 安装步骤
```bash
# 克隆仓库
git clone https://github.com/ll31415926/windows-transparent.git

# 进入项目目录
cd windows-transparent

# 编译 CLI
go build -o wtrans ./cmd/wtrans

# 编译 GUI（推荐使用 Wails CLI）
wails build

# 如果本机没有 wails 命令，可使用等价的 Go 构建命令
go build -tags "desktop,production" -ldflags "-w -s -H windowsgui" -o wtrans-gui.exe ./cmd/wtrans-gui

# 验证 CLI 安装
./wtrans -h

# 小白也能直接用的菜单
# Windows：wtrans-setup.bat
# Linux：./wtrans-setup.sh
```

### 基础使用
```bash
# 列出所有可见窗口
./wtrans list

# 按进程名过滤窗口
./wtrans list --process Code.exe

# 设置窗口透明度为 65%
./wtrans set --process notepad.exe --opacity 65

# 设置并让后续新打开的同名窗口保持相同透明度
./wtrans set --process notepad.exe --opacity 65 --persist

# 恢复窗口为完全不透明
./wtrans restore --process notepad.exe

# 查看保存的规则和后台状态
./wtrans status

# 停止后台保持器，但保留规则
./wtrans stop

# 清空所有规则并停止后台保持器
./wtrans reset

# 从配置文件批量应用透明度规则
./wtrans apply --config config.json

# 诊断当前环境（Linux 排查问题时非常有用）
./wtrans diagnose
```

<br/>

## 🛠️ 命令详解

### list 命令
列出当前所有可见窗口，可选按进程名过滤。

| 参数 | 说明 | 示例 |
|------|------|------|
| `--process` | 按进程名过滤（可选，不区分大小写） | `--process explorer.exe` |

```bash
# 列出所有可见窗口
./wtrans list

# 列出所有 Chrome 窗口
./wtrans list --process chrome.exe
```

**Windows 输出示例：**
```
      HWND        PID  PROCESS              TITLE
  0x1A0234       1234  explorer.exe         任务栏
  0x5B0C68       5678  Code.exe             main.go - windows-transparent
  0x3D0F12       9012  notepad.exe          无标题 - 记事本
```

**Linux 输出示例：**
```
         ID        PID  PROCESS       BACKEND   TITLE
  0x3a00006       1234  firefox       x11       Mozilla Firefox
  0x4200003       5678  code          x11       main.go - windows-transparent
  0x5c0001a       9012  gnome-terminal gnome     终端
```

### set 命令
设置指定进程所有可见窗口的透明度。

| 参数 | 说明 | 示例 |
|------|------|------|
| `--process` | 目标进程名（必填，不区分大小写） | `--process Code.exe` |
| `--opacity` | 不透明度百分比，范围 20-100（必填） | `--opacity 75` |
| `--persist` | 保存规则并启动后台监视，让后续新打开的同名窗口保持相同透明度 | `--persist` |

```bash
# 将 VS Code 窗口设置为 85% 不透明度
./wtrans set --process Code.exe --opacity 85

# 将记事本设置为半透明
./wtrans set --process notepad.exe --opacity 50

# 让后续新打开的记事本窗口也保持 65% 不透明度
./wtrans set --process notepad.exe --opacity 65 --persist
```

### restore 命令
恢复指定进程窗口为完全不透明（100%）。如果有保存的持久规则，也会一起取消。

| 参数 | 说明 | 示例 |
|------|------|------|
| `--process` | 目标进程名（必填，不区分大小写） | `--process Code.exe` |

```bash
# 恢复 VS Code 窗口为完全不透明
./wtrans restore --process Code.exe
```

### status 命令
查看配置路径、已保存的规则，以及后台保持器是否正在运行。

```bash
./wtrans status
./wtrans status --config rules.json
```

### stop 命令
停止后台保持器，但保留所有已保存规则。

```bash
./wtrans stop
```

### reset 命令
停止后台保持器，清空所有已保存规则，并尽量恢复当前可见的对应窗口。

```bash
./wtrans reset
```

### apply 命令
从 JSON 配置文件读取规则，批量应用透明度设置。

| 参数 | 说明 | 示例 |
|------|------|------|
| `--config` | 配置文件路径（可选） | `--config rules.json` |

**默认配置路径：**
- **Windows**：`%APPDATA%\wtrans\config.json`
- **Linux**：`~/.config/wtrans/config.json`

```bash
# 使用默认配置文件
./wtrans apply

# 指定配置文件路径
./wtrans apply --config /home/user/my_rules.json
```

### watch 命令
持续应用默认或指定配置文件中的透明度规则。`set --persist` 会自动启动它，也可以手动运行。

| 参数 | 说明 | 示例 |
|------|------|------|
| `--config` | 配置文件路径（可选） | `--config rules.json` |

```bash
./wtrans watch
./wtrans watch --config rules.json
```

### diagnose 命令
打印当前桌面会话的诊断信息，用于排查 Linux 下的兼容性问题。

```bash
./wtrans diagnose
```

**Linux 输出示例：**
```
Session:
  XDG_SESSION_TYPE=wayland
  XDG_CURRENT_DESKTOP=sway
  DISPLAY=
  WAYLAND_DISPLAY=wayland-1
  DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus

Detected backend: sway

Tools:
  wmctrl: not found
  xprop: not found
  swaymsg: /usr/bin/swaymsg
  hyprctl: not found
  gdbus: /usr/bin/gdbus
  gnome-extensions: not found
```

### gnome-extension 命令（仅 Linux GNOME）
管理 GNOME Shell 扩展，用于在 GNOME Wayland 下实现窗口透明度控制。

| 子命令 | 说明 |
|--------|------|
| `install` | 自动安装 wtrans GNOME Shell 扩展 |
| `status` | 查看扩展安装和启用状态 |

```bash
# 安装 GNOME Shell 扩展
./wtrans gnome-extension install

# 查看扩展状态
./wtrans gnome-extension status
```

> **注意：** 安装扩展后需要重新登录或重启 GNOME Shell（按 `Alt+F2` 输入 `r`）才能生效。

<br/>

## 📄 配置文件

### 配置格式
配置文件为 JSON 格式，包含一个 `rules` 数组，每条规则指定进程名和目标透明度。

```json
{
  "rules": [
    { "process": "notepad.exe", "opacity": 65 },
    { "process": "Code.exe", "opacity": 85 },
    { "process": "firefox", "opacity": 90 },
    { "process": "gnome-terminal", "opacity": 80 }
  ]
}
```

### 配置说明
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `rules` | 数组 | 是 | 透明度规则列表 |
| `rules[].process` | 字符串 | 是 | 进程名称，不区分大小写 |
| `rules[].opacity` | 整数 | 是 | 不透明度百分比，范围 20-100 |

<br/>

## 📊 使用示例

### 1. 日常使用示例
```bash
# 查看当前所有窗口
./wtrans list

# 让终端半透明，方便对照其他窗口
./wtrans set --process gnome-terminal --opacity 60

# 工作完成后恢复
./wtrans restore --process gnome-terminal
```

### 2. 多窗口管理示例
```bash
# 同时让多个应用半透明
./wtrans set --process firefox --opacity 80
./wtrans set --process Code.exe --opacity 85
./wtrans set --process notepad.exe --opacity 65

# 恢复所有
./wtrans restore --process firefox
./wtrans restore --process Code.exe
./wtrans restore --process notepad.exe
```

### 3. 配置文件批量管理
```bash
# 创建配置文件
cat > my_rules.json << 'EOF'
{
  "rules": [
    { "process": "Code.exe", "opacity": 85 },
    { "process": "notepad.exe", "opacity": 65 },
    { "process": "firefox", "opacity": 90 }
  ]
}
EOF

# 批量应用
./wtrans apply --config my_rules.json
```

### 4. Linux 排查问题
```bash
# 查看诊断信息，确认后端和工具是否就绪
./wtrans diagnose

# GNOME Wayland 用户：安装扩展
./wtrans gnome-extension install
./wtrans gnome-extension status

# 强制指定后端（覆盖自动检测）
WTRANS_BACKEND=x11 ./wtrans list
```

<br/>

## 🔧 技术特性

### Windows 实现
- **Win32 API 集成**：通过 `user32.dll` / `kernel32.dll` 直接调用 `EnumWindows`、`SetLayeredWindowAttributes` 等 API
- **分层窗口处理**：智能管理 `WS_EX_LAYERED` 扩展样式，恢复时保留原有属性
- **进程解析**：使用 `CreateToolhelp32Snapshot` 快照实现 PID 到进程名的映射

### Linux 后端
| 后端 | 工具依赖 | 说明 |
|------|----------|------|
| **X11** | `wmctrl`, `xprop` | 通过 `_NET_WM_WINDOW_OPACITY` X 属性控制透明度 |
| **Sway** | `swaymsg` | 通过 IPC 协议调用 `swaymsg opacity` 命令 |
| **Hyprland** | `hyprctl` | 设置 `alpha` 和 `alphainactive` 窗口属性 |
| **GNOME Wayland** | `gdbus`, GNOME Shell 扩展 | 通过 D-Bus 接口与自定义 GNOME Shell 扩展通信 |

**后端自动检测优先级：**
1. `WTRANS_BACKEND` 环境变量强制指定
2. 检测 Hyprland（`HYPRLAND_INSTANCE_SIGNATURE` 环境变量）
3. 检测 Sway（`SWAYSOCK` 环境变量）
4. 检测 `XDG_CURRENT_DESKTOP` 中的桌面环境名称
5. 检测 `XDG_SESSION_TYPE`（x11/wayland）
6. 回退检测 PATH 中的工具可用性

### 通用特性
- **零外部依赖**：完全使用 Go 标准库，无任何第三方模块依赖
- **跨平台编译**：使用 Go build tags 实现平台特定代码隔离
- **可测试架构**：通过 `runner` 接口抽象外部命令调用，便于单元测试
- **进程名匹配**：不区分大小写，同时匹配进程名和窗口类名

### 错误处理
- **双错误类型**：`UsageError`（参数错误，退出码 2）和普通错误（退出码 1）
- **详细错误信息**：Linux 下提供环境检测和修复建议
- **跨平台兼容**：不受支持的平台返回 `ErrUnsupported` 错误

### 测试覆盖
| 包 | 测试数 | 覆盖内容 |
|---|---|---|
| `cli` | 7 | 命令解析、参数验证、错误处理 |
| `config` | 4 | 配置解析、验证、保存功能 |
| `opacity` | 4 | Alpha 转换、边界值验证 |
| `window`（通用） | 2 | 进程名匹配、大小写处理 |
| `window`（Linux） | 15+ | 后端检测、X11/Sway/Hyprland/GNOME 命令构造、输出解析 |

<br/>

## 📌 注意事项

### Windows 使用须知
- **管理员权限**：部分系统窗口（如任务管理器）可能需要管理员权限才能修改透明度
- **进程名匹配**：需包含扩展名（如 `notepad.exe`）

### Linux 使用须知
- **X11 会话**：需要安装 `wmctrl` 和 `xprop`（`sudo apt install wmctrl x11-utils`）
- **Sway**：需要 `swaymsg` 已安装（通常随 Sway 一起提供）
- **Hyprland**：需要 `hyprctl` 已安装（通常随 Hyprland 一起提供）
- **GNOME Wayland**：需要安装 GNOME Shell 扩展，运行 `./wtrans gnome-extension install` 后重新登录
- **KDE/KWin**：暂不支持，Wayland 下的 KDE 窗口不支持外部透明度控制
- **排查问题**：运行 `./wtrans diagnose` 查看当前环境和工具依赖状态

### 通用说明
- **透明度范围**：透明度值限制在 20-100 之间，低于 20% 可能导致窗口难以交互
- **即时生效**：透明度设置立即生效，无需重启目标应用
- **持久化**：使用 `set --persist` 可让后续新打开的同名窗口持续应用规则；使用 `restore --process ...` 可删除该持久规则
- **小白模式**：Windows 用 `wtrans-setup.bat`，Linux 用 `./wtrans-setup.sh`
- **环境变量**：可通过 `WTRANS_BACKEND` 强制指定后端（如 `WTRANS_BACKEND=x11`）

<br/>

## ⚠️ 免责声明

**使用本工具前请务必阅读并同意以下条款：**

1. **合法用途**：仅限在您拥有合法访问权限的系统上使用
2. **系统安全**：避免对关键系统进程设置过低透明度，以免影响正常操作
3. **数据安全**：本工具不收集、不传输任何用户数据
4. **风险自担**：因使用本工具导致的任何问题由用户自行承担
5. **备份建议**：首次使用建议先测试单个窗口，确认效果后再批量应用

<br/>

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个项目！


---

**让每一个窗口恰到好处** - 精准控制你的工作空间 🪟
