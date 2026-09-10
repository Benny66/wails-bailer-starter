//go:build !darwin

package tray

import (
	"github.com/energye/systray"
)

// Start 启动托盘，返回停止函数（优雅关闭时调用）。
// start 应在 Wails 启动后调用；返回的 stop 在关闭前调用。
func Start(opts Options, h Handlers) (stop func(), err error) {
	if opts.Icon == nil {
		opts.Icon = placeholderIcon()
	}

	onReady := func() {
		systray.SetIcon(opts.Icon)
		systray.SetTooltip(opts.Tooltip)
		systray.SetTitle(opts.Tooltip)
		// 双击托盘图标恢复窗口
		systray.SetOnDClick(func(menu systray.IMenu) {
			if h.OnShow != nil {
				h.OnShow()
			}
		})
		buildMenu(h)
	}
	onExit := func() {}

	start, end := systray.RunWithExternalLoop(onReady, onExit)
	start()

	return func() {
		end()
	}, nil
}

// buildMenu 构建右键菜单。
func buildMenu(h Handlers) {
	show := systray.AddMenuItem("显示", "恢复窗口")
	quit := systray.AddMenuItem("退出", "退出应用")
	show.Click(func() {
		if h.OnShow != nil {
			h.OnShow()
		}
	})
	quit.Click(func() {
		if h.OnQuit != nil {
			h.OnQuit()
		}
	})
}

// placeholderIcon 返回 1x1 透明 PNG，仅作无图标时的兜底。
func placeholderIcon() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
		0x0d, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x62, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
