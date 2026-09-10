# runtime — 进程生命周期与平台服务

## Why

B 端桌面工具的核心体验（常驻托盘、单实例唤起、优雅关闭、崩溃可查）依赖一套运行时服务。地基只有数据库，没有这些，应用"能跑"但"不经用"。Wails v2.15.0 已原生内置部分能力，本 change 聚焦"开开关 + 包一层"，只啃两个硬骨头：托盘与日志轮转。

## What Changes

- **单实例锁**：`options.SingleInstanceLock`，二次启动唤起已有窗口（置顶 + 调 `WindowUnminimise`），而非报错。
- **关到托盘语义**：`HideWindowOnClose` + 托盘双击/菜单恢复窗口。
- **系统托盘**：`getlantern/systray`（Wails 官方 tray 已挂起），goroutine 内跑，退出时 `Quit()`；三平台图标适配（.ico/.png 模板）。
- **原生对话框封装**：包一层 `runtime.OpenFileDialog` / `OpenDirectoryDialog` / `SaveFileDialog` / `MessageDialog`，暴露成 App 绑定方法。
- **日志 + 轮转**：`logger.FileLogger` + `lumberjack` 按大小轮转；分级 + 文件/控制台双写。
- **config 读写**：`config.json` 落在用户配置目录，读写封装，App 启动加载。
- **优雅关闭**：`OnShutdown` 释放数据库连接；`OnBeforeClose` 处理"关闭=托盘 or 退出"。
- **崩溃日志保存**：panic recovery + 崩溃信息落盘（独立于常规日志）。

## Capabilities

### New Capabilities

- `app-lifecycle`: 单实例锁、关到托盘、优雅关闭、崩溃捕获与落盘。
- `system-tray`: 系统托盘 + 右键菜单 + 双击恢复（getlantern/systray）。
- `native-dialogs`: 文件/目录/保存/消息对话框的绑定封装。
- `logging`: 文件+控制台双写日志、分级、按大小轮转。
- `app-config`: `config.json` 读写与启动加载。

### Modified Capabilities

<!-- 无 -->

## Impact

- **新增代码**：`internal/tray/`、`internal/logging/`、`internal/config/`、`internal/dialog/`、`app.go` 新增绑定方法与生命周期钩子、`main.go` 增加 options 配置。
- **依赖**：getlantern/systray、lumberjack（Go）。
- **风险**：systray 跨平台图标/生命周期整合是最大不确定点，若受阻可降级为"仅 Windows/macOS 基础托盘"。
- **破坏性**：`app.go` 从纯骨架变为承载绑定方法与生命周期钩子。
