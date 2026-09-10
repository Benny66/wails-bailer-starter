# runtime — 技术设计

## Context

Wails v2.15.0 已原生内置：单实例锁（`options.SingleInstanceLock`）、关到托盘（`HideWindowOnClose`）、原生对话框（`runtime.*Dialog`）、优雅关闭（`OnShutdown`/`OnBeforeClose`）、panic 恢复、窗口控制（`runtime.Window*`）。托盘官方挂起（`buildassets/onhold/tray`），日志 `FileLogger` 无轮转。

本 change 聚焦"开开关 + 包一层"，硬骨头只有托盘与日志轮转。

## Goals / Non-Goals

**Goals:**
- 单实例二次启动唤起窗口。
- 托盘 + 右键菜单 + 双击恢复。
- 对话框绑定封装。
- 日志轮转。
- config 读写。
- 优雅关闭 + 崩溃落盘。

**Non-Goals:**
- 自动更新（已从主线移除）。
- 全局快捷键（v1.1）。
- 复杂差分更新。

## Decisions

### D1: 托盘用 `energye/systray`，且 macOS 降级为 no-op
Wails 官方 tray 挂起，getlantern/systray 是主流选择，但实测在 macOS 上**架构冲突**（三层坑，逐层深入）：

1. **链接期**：getlantern 的 darwin 后端与 Wails 都定义 `AppDelegate` → duplicate symbol。
2. **运行时**：getlantern 及 efeenesc fork 的 `nativeLoop` 调 `[NSApp run]`，与 Wails 抢 NSApplication 事件循环 → SIGTRAP。
3. **运行时**：energye fork 提供 `RunWithExternalLoop` 不抢循环，但 `nativeStart` 仍 `setDelegate:owner` 且 `NSStatusBar` 只能主线程创建，Wails 的 `OnStartup` 在 goroutine → `NSInternalInconsistencyException`。

**结论**：macOS 上 `NSApplication` 只能有一个 delegate，Wails 自己占着，所有 systray 库都想当 delegate 创建 `NSStatusBar`，本质互斥——这正是 Wails 官方把 tray 挂 `onhold/` 的原因。

**最终方案**：
- Windows/Linux：`energye/systray` + `RunWithExternalLoop`（不抢主循环），build tag 隔离。
- macOS：`Start()` 为 no-op（打 WARN 日志），托盘能力跳过；`HideWindowOnClose` 在 macOS 为 `false`（关闭即退出），其余平台 `true`（隐藏到托盘）。

**备选**：手写 macOS 的 NSStatusBar（ObjC + cgo，不碰 delegate）——否决，工作量最大、最脆，后续维护成本高，收益（仅 macOS 托盘）不划算。

### D2: 托盘生命周期用 channel 协调，避免 goroutine 泄漏
`internal/tray` 提供 `Start(ctx, onQuit)` / `Stop()`，内部用 channel 通知退出。经验文档强调"优雅关闭避免 goroutine 残留"，托盘是最大残留源。

### D3: 日志轮转用 lumberjack，FileLogger 作为分级前端
`lumberjack.Logger`（按大小轮转 + 保留份数）作为 writer，`logger.FileLogger` 或自封装 logger 写入。`make dev` 用 Debug 级，生产用 Info 级（`LogLevel`/`LogLevelProduction` 分开配）。

### D4: config 读写用标准库 encoding/json + 用户配置目录
`config.json` 落 `os.UserConfigDir()/appName/`，与数据库同目录。启动时不存在则写默认值；字段用结构体 + json tag。禁用硬编码路径。

### D5: 优雅关闭顺序：托盘停 → 保存 config → 关 db
`OnShutdown` 里按依赖顺序回收：先停托盘 goroutine，再落盘 config，最后 `db.Close()`。`OnBeforeClose` 决定"关闭=托盘（HideWindowOnClose 已处理）"。

### D6: 崩溃落盘独立于常规日志
`panic` 被 Wails 默认恢复（`DisablePanicRecovery: false`），但默认只打日志。补一层：`OnStartup` 里 `defer recover` 包住业务启动逻辑，崩溃信息写独立 `crash-<timestamp>.log`。

## Risks / Trade-offs

- **[Risk] macOS 托盘与 Wails 架构冲突** → 已决策：macOS 托盘降级为 no-op，仅 Windows/Linux 提供。见 D1 完整分析。
- **[Risk] systray 与 Wails 主循环冲突** → `energye/systray` 的 `RunWithExternalLoop` 不抢主循环，托盘作为外挂共存。
- **[Risk] macOS 托盘图标需模板图（黑白）** → 已不适用（macOS 无托盘）；Windows/Linux 用普通图标。
- **[Risk] lumberjack 引入第三方依赖** → 已在 deps.yaml 登记。
