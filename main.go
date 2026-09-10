package main

import (
	"embed"
	"log"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"wails-bailer-starter/internal/config"
	"wails-bailer-starter/internal/database"
	"wails-bailer-starter/internal/logging"
	"wails-bailer-starter/internal/service"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var trayIconBytes []byte

// appName 应用标识，用于定位用户配置目录下的数据文件与日志。
const appName = "wails-bailer-starter"

func main() {
	// 日志：开发态 Debug，生产态 Info（生产态由 -tags production 或环境变量决定，这里默认 Debug 便于开发）。
	logLevel := slog.LevelDebug
	if os.Getenv("WAILS_PRODUCTION") == "1" {
		logLevel = slog.LevelInfo
	}
	closeLog, err := logging.Init(logging.Options{AppName: appName, Level: logLevel})
	if err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}
	defer closeLog()

	// 组合根：数据库 → 服务 → 配置 → App
	db, err := database.Init(appName)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	cfg, err := config.Load(appName)
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	app := NewApp(service.New(db), cfg)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "wails-bailer-starter",
		Width:  1024,
		Height: 768,
		// 无边框按平台分叉：Windows/Linux 自绘标题栏（三按钮），macOS 用原生交通灯。
		// macOS 必须 false，否则 Wails 跳过 NSWindowStyleMaskTitled，交通灯消失（见 design D3）。
		Frameless: frameless(),
		// 单实例锁：二次启动唤起已有窗口。
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "wails-bailer-starter-single-instance",
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				slog.Info("检测到二次启动，唤起已有窗口", "args", data.Args)
				runtime.Show(app.ctx)
				runtime.WindowUnminimise(app.ctx)
			},
		},
		// 关闭 = 隐藏到托盘（Windows/Linux）；macOS 无托盘，关闭即退出。
		// 值由平台决定，见 hideWindowOnClose()。
		HideWindowOnClose: hideWindowOnClose(),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 14, G: 15, B: 18, A: 1}, // 与 tokens 的 --bg-base 一致，避免启动白闪
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		// macOS：透明标题栏 + 交通灯 inset + 内容顶到顶，不自绘三按钮。
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
		},
		// Windows：保留无边框窗口的圆角/阴影装饰，跟随系统明暗主题。
		Windows: &windows.Options{
			Theme: windows.SystemDefault,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// trayIcon 返回托盘图标字节。
func trayIcon() []byte {
	return trayIconBytes
}
