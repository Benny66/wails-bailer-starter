//go:build darwin

package main

// hideWindowOnClose 返回"关闭时隐藏窗口"。
// macOS 无托盘常驻，关闭应真正退出（否则窗口消失且无法恢复）。
func hideWindowOnClose() bool {
	return false
}
