// Package tray 封装系统托盘。
//
// 平台支持：
//   - Windows / Linux：使用 energye/systray（getlantern 的 fork，提供
//     RunWithExternalLoop 让托盘与 Wails 主事件循环共存）。
//   - macOS：不启动托盘（Wails v2 的 NSApplication delegate 与所有 systray
//     库冲突——库都想当 delegate 来创建 NSStatusBar，而 Wails 自己占着 delegate，
//     导致链接期符号冲突与运行时崩溃）。见 start_darwin.go 的 no-op 实现。
//
// 本文件只定义共享类型，平台实现见 start_windows_linux.go / start_darwin.go。
package tray

// Handlers 是托盘交互回调，由调用方注入。
type Handlers struct {
	// OnShow 双击/单击托盘图标或点"显示"时触发（恢复窗口）。
	OnShow func()
	// OnQuit 点"退出"时触发（优雅退出应用）。
	OnQuit func()
}

// Options 是托盘启动参数。
type Options struct {
	// Icon 是托盘图标二进制（.ico 或 .png）。macOS 建议用模板图（黑白）。
	Icon []byte
	// Tooltip 悬停提示。
	Tooltip string
}
