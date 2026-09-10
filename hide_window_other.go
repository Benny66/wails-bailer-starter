//go:build !darwin

package main

// hideWindowOnClose 返回"关闭时隐藏窗口"。
// Windows/Linux 有托盘常驻，关闭应隐藏到托盘。
func hideWindowOnClose() bool {
	return true
}
