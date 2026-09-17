package main

import (
	"embed"
	"log"
	"log/slog"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"__APP_NAME__/internal/apperr"
	"__APP_NAME__/internal/appinfo"
	"__APP_NAME__/internal/config"
	"__APP_NAME__/internal/database"
	"__APP_NAME__/internal/logging"
	"__APP_NAME__/internal/service"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var trayIconBytes []byte

// appName 应用标识，用于定位用户配置目录下的数据文件与日志。
const appName = "__APP_NAME__"

// bootTimer 记录组合根各阶段的耗时，输出启动时间线的 Go 侧半段。
//
// 为什么需要：此前启动只有「应用启动」与「前端已挂载」两点之差，实测 359–733ms
// （同一天两次运行差 2×），中间是黑盒——不知道该优化 webview、前端产物还是组合根。
// 前端那半段由 frontend/src/main.ts 记录，两半靠 app.log 的绝对时间戳拼起来
// （见 openspec/changes/perf-instrumentation/design.md D1）。
//
// 该打点由 internal/guard/wiring_test.go 核对数量：删掉没有任何症状，
// 只是启动优化重新变成盲调。
type bootTimer struct {
	start time.Time // 进程起点
	last  time.Time // 上一阶段结束点
}

// newBootTimer 以调用时刻为进程起点。
func newBootTimer() *bootTimer {
	now := time.Now()
	return &bootTimer{start: now, last: now}
}

// stage 记录一个阶段的耗时与该阶段结束时的累计耗时，并推进计时起点。
func (b *bootTimer) stage(name string) {
	now := time.Now()
	slog.Info("启动阶段",
		"stage", name,
		"ms", now.Sub(b.last).Milliseconds(),
		"total_ms", now.Sub(b.start).Milliseconds(),
	)
	b.last = now
}

func main() {
	boot := newBootTimer()

	// 日志：开发态 Debug，生产态 Info（构建模式判定见 buildmode_*.go）。
	logLevel := slog.LevelDebug
	if !debugBuild {
		logLevel = slog.LevelInfo
	}
	closeLog, err := logging.Init(logging.Options{
		AppName: appName,
		Level:   logLevel,
		Version: appinfo.Resolve(),
	})
	if err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}
	defer closeLog()
	boot.stage("日志初始化")

	// 组合根：数据库 → 服务 → 配置 → App
	db, err := database.Init(appName)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	boot.stage("数据库初始化")

	cfg, err := config.Load(appName)
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	app := NewApp(service.New(db), cfg)
	boot.stage("配置加载与组装")

	// 窗口几何：在【创建期】给定尺寸与最大化状态，首帧即正确。
	// 不能在运行期设——OnStartup 在 goroutine 中执行，紧接着窗口就显示，二者存在竞态，
	// 用户会看到窗口跳一下（见 openspec/changes/runtime-pipeline/design.md D1）。
	winWidth, winHeight := cfg.WindowSize()
	startState := options.Normal
	if cfg.WindowMaximised {
		startState = options.Maximised
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:            "__APP_NAME__",
		Width:            winWidth,
		Height:           winHeight,
		WindowStartState: startState,
		// Wails 的日志接口接到 slog：不接的话，前端的 LogError/LogInfo 与 Wails 自身的
		// 内部错误【只写 stdout】——打包后无人可见，且不接不会有任何症状。
		// 故这条接线由 internal/guard/wiring_test.go 强制。
		Logger:             logging.WailsAdapter{},
		LogLevel:           logger.DEBUG,
		LogLevelProduction: logger.INFO,
		// 错误协议：把 Go 的 error 序列化为 JSON 字符串，前端 JSON.parse 还原为 {code,message}。
		// 注意：ErrorFormatter 签名是 func(error) any，但【必须】返回 string——
		// 返回对象会被前端 new Error(payload) 强转成 "[object Object]"（Wails v2.15.0 实测）。
		ErrorFormatter: apperr.Format,
		// 无边框按平台分叉：Windows/Linux 自绘标题栏（三按钮），macOS 用原生交通灯。
		// macOS 必须 false，否则 Wails 跳过 NSWindowStyleMaskTitled，交通灯消失（见 design D3）。
		Frameless: frameless(),
		// 单实例锁：二次启动唤起已有窗口。
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "__APP_NAME__-single-instance",
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				slog.Info("检测到二次启动，唤起已有窗口", "args", data.Args)
				runtime.Show(app.ctx)
				runtime.WindowUnminimise(app.ctx)
				// 把二次启动的启动参数推给前端（复用事件契约）。
				// 首次启动的参数不经事件——那时前端还没订阅，必然丢；用 GetLaunchArgs 查询。
				app.notifySecondInstance(data.Args, data.WorkingDirectory)
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
