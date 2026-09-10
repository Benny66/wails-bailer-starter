//go:build darwin

package tray

import "log/slog"

// Start 在 macOS 上是 no-op：Wails v2 不支持 macOS 系统托盘。
//
// 原因：Wails v2 的 darwin frontend 自己持有 NSApplication 的 delegate（跑 WebKit），
// 而所有 systray 库（getlantern 及其 fork）都要 setDelegate 来创建 NSStatusBar，
// 两者互斥——链接期 AppDelegate 符号冲突，运行时 NSInternalInconsistencyException。
//
// 故 macOS 关闭即退出（HideWindowOnClose 在 macOS 上实际也由 Wails 处理为真退出），
// 托盘能力仅在 Windows/Linux 提供。
func Start(opts Options, h Handlers) (stop func(), err error) {
	slog.Warn("macOS 不支持系统托盘，托盘能力已跳过")
	return func() {}, nil
}
