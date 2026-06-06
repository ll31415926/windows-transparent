# 🪟 wtrans 窗口透明度工具

<p align="center">
  <img alt="Go版本" src="https://img.shields.io/badge/Go-1.24%2B-blue">
  <img alt="平台支持" src="https://img.shields.io/badge/平台-Windows-green">
  <img alt="开源协议" src="https://img.shields.io/badge/许可-Apache%202.0-orange">
  <img alt="零依赖" src="https://img.shields.io/badge/依赖-零依赖-brightgreen">
</p>

> 轻量级 Windows 窗口透明度调节工具，通过 Win32 API 直接控制任意应用窗口的不透明度

wtrans 是一个纯 Go 实现的命令行窗口透明度管理工具，无需任何第三方依赖。通过直接调用 `user32.dll` 和 `kernel32.dll` 的 Win32 API，实现对任意可见窗口的透明度设置、恢复和批量管理，适合需要多窗口协同工作的开发者和高级用户。

<br/>

## ✨ 核心功能

- **🪟 窗口透明度控制**：设置任意可见窗口的不透明度（20%-100%），支持精确调节
- **📋 窗口列表查看**：列出所有可见窗口，显示 HWND、PID、进程名和标题信息
- **🔄 一键恢复**：将窗口恢复为完全不透明，并智能移除分层样式标记
- **📦 批量配置应用**：通过 JSON 配置文件批量设置多个进程的透明度规则
- **🔍 进程名过滤**：支持按进程名筛选窗口，匹配时不区分大小写
- **⚙️ 智能样式管理**：自动处理 `WS_EX_LAYERED` 扩展样式，恢复时保留原有属性
- **🛡️ 非破坏性操作**：恢复透明度时检查原有分层属性，避免覆盖窗口原始配置
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

# 编译项目
go build -o wtrans.exe ./cmd/wtrans

# 验证安装
./wtrans.exe -h
```

### 基础使用
```bash
# 列出所有可见窗口
./wtrans.exe list

# 按进程名过滤窗口
./wtrans.exe list --process Code.exe

# 设置记事本窗口透明度为 65%
./wtrans.exe set --process notepad.exe --opacity 65

# 恢复记事本窗口为完全不透明
./wtrans.exe restore --process notepad.exe

# 从配置文件批量应用透明度规则
./wtrans.exe apply --config config.json
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
./wtrans.exe list

# 列出所有 Chrome 窗口
./wtrans.exe list --process chrome.exe
```

**输出示例：**
```
      HWND        PID  PROCESS              TITLE
  0x1A0234       1234  explorer.exe         任务栏
  0x5B0C68       5678  Code.exe             main.go - windows-transparent
  0x3D0F12       9012  notepad.exe          无标题 - 记事本
```

### set 命令
设置指定进程所有可见窗口的透明度。

| 参数 | 说明 | 示例 |
|------|------|------|
| `--process` | 目标进程名（必填，不区分大小写） | `--process Code.exe` |
| `--opacity` | 不透明度百分比，范围 20-100（必填） | `--opacity 75` |

```bash
# 将 VS Code 窗口设置为 85% 不透明度
./wtrans.exe set --process Code.exe --opacity 85

# 将记事本设置为半透明
./wtrans.exe set --process notepad.exe --opacity 50
```

### restore 命令
恢复指定进程窗口为完全不透明（100%），并智能移除分层窗口样式。

| 参数 | 说明 | 示例 |
|------|------|------|
| `--process` | 目标进程名（必填，不区分大小写） | `--process Code.exe` |

```bash
# 恢复 VS Code 窗口为完全不透明
./wtrans.exe restore --process Code.exe
```

### apply 命令
从 JSON 配置文件读取规则，批量应用透明度设置。

| 参数 | 说明 | 示例 |
|------|------|------|
| `--config` | 配置文件路径（可选，默认 `%APPDATA%\wtrans\config.json`） | `--config rules.json` |

```bash
# 使用默认配置文件
./wtrans.exe apply

# 指定配置文件路径
./wtrans.exe apply --config C:\Users\me\wtrans.json
```

<br/>

## 📄 配置文件

### 配置格式
配置文件为 JSON 格式，包含一个 `rules` 数组，每条规则指定进程名和目标透明度。

```json
{
  "rules": [
    { "process": "notepad.exe", "opacity": 65 },
    { "process": "Code.exe", "opacity": 85 },
    { "process": "chrome.exe", "opacity": 90 },
    { "process": "explorer.exe", "opacity": 95 }
  ]
}
```

### 配置说明
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `rules` | 数组 | 是 | 透明度规则列表 |
| `rules[].process` | 字符串 | 是 | 进程名称，不区分大小写 |
| `rules[].opacity` | 整数 | 是 | 不透明度百分比，范围 20-100 |

### 默认配置路径
- **Windows**：`%APPDATA%\wtrans\config.json`
- 使用 `apply` 命令时，若未指定 `--config` 参数，将自动读取默认路径

<br/>

## 📊 使用示例

### 1. 日常使用示例
```bash
# 查看当前所有窗口
./wtrans.exe list

# 让记事本半透明，方便对照其他窗口
./wtrans.exe set --process notepad.exe --opacity 60

# 工作完成后恢复
./wtrans.exe restore --process notepad.exe
```

### 2. 多窗口管理示例
```bash
# 同时让多个应用半透明
./wtrans.exe set --process chrome.exe --opacity 80
./wtrans.exe set --process Code.exe --opacity 85
./wtrans.exe set --process notepad.exe --opacity 65

# 一键恢复所有
./wtrans.exe restore --process chrome.exe
./wtrans.exe restore --process Code.exe
./wtrans.exe restore --process notepad.exe
```

### 3. 配置文件批量管理
```bash
# 创建配置文件
echo '{
  "rules": [
    { "process": "Code.exe", "opacity": 85 },
    { "process": "notepad.exe", "opacity": 65 },
    { "process": "chrome.exe", "opacity": 90 }
  ]
}' > my_rules.json

# 批量应用
./wtrans.exe apply --config my_rules.json
```

<br/>

## 🔧 技术特性

### Win32 API 集成
- **窗口枚举**：通过 `EnumWindows` 遍历所有顶层窗口
- **可见性检测**：使用 `IsWindowVisible` 过滤不可见窗口
- **透明度控制**：调用 `SetLayeredWindowAttributes` 设置窗口 alpha 值
- **样式管理**：通过 `GetWindowLongPtrW` / `SetWindowLongPtrW` 管理 `WS_EX_LAYERED` 扩展样式
- **进程解析**：使用 `CreateToolhelp32Snapshot` 快照实现 PID 到进程名的映射

### 分层窗口处理
- **智能添加样式**：设置透明度前检查窗口是否已有 `WS_EX_LAYERED` 样式，避免重复设置
- **非破坏性恢复**：恢复时读取当前分层属性，仅在窗口无其他分层特性时移除样式标记
- **Alpha 转换公式**：`alpha = (opacity × 255 + 50) / 100`，确保精度和四舍五入

### 错误处理
- **Win32 错误检测**：调用 `SetLastError(0)` 后检查 `errno`，区分真实错误和零值返回
- **双错误类型**：`UsageError`（参数错误，退出码 2）和普通错误（退出码 1）
- **跨平台兼容**：非 Windows 系统返回 `ErrUnsupported` 错误

### 测试覆盖
| 包 | 测试数 | 覆盖内容 |
|---|---|---|
| `cli` | 7 | 命令解析、参数验证、错误处理 |
| `config` | 4 | 配置解析、验证、保存功能 |
| `opacity` | 4 | Alpha 转换、边界值验证 |
| `window` | 2 | 进程名匹配、大小写处理 |

<br/>

## 📌 注意事项

### 使用须知
- **仅限 Windows**：工具依赖 Win32 API，仅支持 Windows 系统
- **管理员权限**：部分系统窗口（如任务管理器）可能需要管理员权限才能修改透明度
- **进程名匹配**：进程名匹配不区分大小写，但需包含扩展名（如 `notepad.exe`）
- **透明度范围**：透明度值限制在 20-100 之间，低于 20% 可能导致窗口难以交互

### 性能说明
- **即时生效**：透明度设置立即生效，无需重启目标应用
- **低资源占用**：纯 Go 实现，无外部依赖，内存占用极低
- **非持久化**：透明度设置在目标应用重启后会恢复默认，需重新应用

### 兼容性
- **系统要求**：Windows 10 及以上版本
- **Go 版本**：Go 1.24+
- **应用兼容**：大多数标准 Windows 应用均支持透明度调节

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


**让每一个窗口恰到好处** - 精准控制你的工作空间 🪟
