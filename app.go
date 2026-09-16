package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"__APP_NAME__/internal/appdir"
	"__APP_NAME__/internal/apperr"
	"__APP_NAME__/internal/appinfo"
	"__APP_NAME__/internal/config"
	"__APP_NAME__/internal/dialog"
	"__APP_NAME__/internal/event"
	"__APP_NAME__/internal/reveal"
	"__APP_NAME__/internal/service"
	"__APP_NAME__/internal/tray"
	// gen:import
)

// App 应用结构体，承载前端可调用的绑定方法。
// 分层纪律：App 不直接持有 *gorm.DB、不 import database，业务访问一律经 svc。
// 运行时服务（config/tray/dialog）由 App 持有并暴露为绑定方法。
type App struct {
	ctx      context.Context
	svc      *service.Service
	cfg      *config.Config
	trayStop func()
}

// NewApp 创建应用实例，注入服务聚合与配置（由 main.go 完成组合根组装）。
func NewApp(svc *service.Service, cfg *config.Config) *App {
	return &App{svc: svc, cfg: cfg}
}

// startup 在应用启动时调用，保存 ctx 供后续运行时方法使用。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 启动系统托盘（独立 goroutine）
	stop, err := tray.Start(tray.Options{
		Icon:    trayIcon(),
		Tooltip: appName,
	}, tray.Handlers{
		OnShow: func() {
			runtime.Show(a.ctx)
			runtime.WindowUnminimise(a.ctx)
		},
		OnQuit: func() {
			runtime.Quit(a.ctx)
		},
	})
	if err != nil {
		slog.Error("启动托盘失败", "err", err)
	} else {
		a.trayStop = stop
	}
}

// shutdown 在应用关闭前调用，按顺序回收资源。
func (a *App) shutdown(ctx context.Context) {
	// 1. 停托盘
	if a.trayStop != nil {
		a.trayStop()
	}
	// 2. 采集窗口几何并落盘配置（一次 Save 同时写主题与窗口状态）
	a.captureWindowGeometry(ctx)
	if a.cfg != nil {
		if err := a.cfg.Save(); err != nil {
			slog.Error("保存配置失败", "err", err)
		}
	}
	// 3. 关数据库
	if err := a.svc.Close(); err != nil {
		slog.Error("关闭数据库失败", "err", err)
	}
}

// captureWindowGeometry 把当前窗口尺寸与最大化状态写入配置，供下次启动还原。
//
// 注意：窗口【最大化时不去读尺寸】——此时读到的是屏幕尺寸，回填后用户下次取消最大化
// 会得到一个占满屏幕的"还原"尺寸。保持上次的非最大化尺寸才符合直觉。
// 读失败不阻断关闭：窗口几何是体验优化，不该让退出流程失败。
func (a *App) captureWindowGeometry(ctx context.Context) {
	if a.cfg == nil {
		return
	}
	maximised := runtime.WindowIsMaximised(ctx)
	if !maximised {
		if width, height := runtime.WindowGetSize(ctx); width > 0 && height > 0 {
			a.cfg.WindowWidth, a.cfg.WindowHeight = width, height
		}
	}
	a.cfg.WindowMaximised = maximised

	// 落盘前记一条：用户报「窗口大小没记住」时，先看这行是否出现、值是否合理。
	slog.Info("窗口几何已记录",
		"width", a.cfg.WindowWidth, "height", a.cfg.WindowHeight, "maximised", maximised)
}

// ---------------------------------------------------------------------------
// 绑定方法：原生对话框
// ---------------------------------------------------------------------------

// SelectFile 弹出选择文件对话框，返回所选路径。
func (a *App) SelectFile(title string) (string, error) {
	return dialog.SelectFile(a.ctx, title)
}

// SelectDirectory 弹出选择目录对话框，返回所选路径。
func (a *App) SelectDirectory(title string) (string, error) {
	return dialog.SelectDirectory(a.ctx, title)
}

// SelectSaveFile 弹出保存文件对话框，返回目标路径。
func (a *App) SelectSaveFile(title, defaultFilename string) (string, error) {
	return dialog.SelectSaveFile(a.ctx, title, defaultFilename)
}

// Confirm 弹出确认框，返回用户是否确认。
func (a *App) Confirm(title, message string) (bool, error) {
	return dialog.Confirm(a.ctx, title, message)
}

// ---------------------------------------------------------------------------
// 绑定方法：配置
// ---------------------------------------------------------------------------

// GetTheme 返回当前主题设置。
func (a *App) GetTheme() string {
	return a.cfg.Theme
}

// SetTheme 保存主题设置并写回配置文件。
// 入参非法返回 apperr.Validation（契约：业务错误经 apperr 构造，前端据 code 分流）。
func (a *App) SetTheme(theme string) error {
	// 合法值：dark / light / ""（空串 = 未设置，语义为"跟随系统"）。
	if theme != "" && theme != "dark" && theme != "light" {
		return apperr.Validation("主题取值非法，仅支持 dark/light")
	}
	a.cfg.Theme = theme
	if err := a.cfg.Save(); err != nil {
		// 落盘失败是系统错误，经 Wrap 归一化为 internal（原始错误保留在 Detail）。
		return apperr.Wrap(err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// 绑定方法：应用数据目录
// ---------------------------------------------------------------------------

// GetDataDir 返回应用数据目录的绝对路径（数据库、配置、日志同目录）。
// 供前端展示/复制，便于用户自行定位日志。
func (a *App) GetDataDir() (string, error) {
	dir, err := appdir.Dir(appName)
	if err != nil {
		return "", apperr.Wrap(err)
	}
	return dir, nil
}

// OpenDataDir 在系统文件管理器中打开应用数据目录。
func (a *App) OpenDataDir() error {
	dir, err := appdir.Dir(appName)
	if err != nil {
		return apperr.Wrap(err)
	}
	if err := reveal.Open(dir); err != nil {
		return apperr.Wrap(err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// 绑定方法：应用信息与数据库导出
// ---------------------------------------------------------------------------

// GetAppInfo 返回版本、平台与关键路径，供「关于」界面与排障使用。
// 排障时让用户念一次本方法的返回值，版本/日志位置就都齐了。
func (a *App) GetAppInfo() (appinfo.Info, error) {
	info, err := appinfo.Current(appName)
	if err != nil {
		return appinfo.Info{}, apperr.Wrap(err)
	}
	return info, nil
}

// GetLaunchArgs 返回本次启动的启动参数（不含可执行文件路径）。
//
// 二次启动（应用已在运行时再次唤起）的参数【不】经此返回，而是经
// `app:second-instance` 事件推送——那一次前端已经挂载，事件不会丢。
func (a *App) GetLaunchArgs() []string {
	if len(os.Args) <= 1 {
		return []string{}
	}
	// 返回副本：os.Args 是全局切片，直接把内部状态交出去不合适。
	out := make([]string, len(os.Args)-1)
	copy(out, os.Args[1:])
	return out
}

// ExportDatabase 把数据库完整导出到 targetPath（通常来自 SelectSaveFile）。
// 目标已存在会被覆盖（用户已在系统保存对话框确认过）；位于数据目录内会被拒绝。
func (a *App) ExportDatabase(targetPath string) error {
	return a.svc.ExportDatabase(appName, targetPath)
}

// ActionSecondInstance 是「应用被二次启动」的事件动作（单一真相）。
// 前端镜像在 frontend/src/lib/app.ts 的 AppEvent，由 internal/guard/parity_test.go 断言一致。
const ActionSecondInstance = "second-instance"

// SecondInstancePayload 是 `app:second-instance` 事件的载荷。
// 前端侧的类型镜像见 frontend/src/lib/app.ts 的 SecondInstancePayload，字段名需一致。
type SecondInstancePayload struct {
	// Args 二次启动时的命令行参数。
	Args []string `json:"args"`
	// WorkingDirectory 二次启动时的工作目录。
	WorkingDirectory string `json:"working_dir"`
}

// notifySecondInstance 把二次启动的参数推给前端（事件名走事件契约 <domain>:<action>）。
// 应用上下文未就绪时只记日志：这是尽力而为的通知，不该 panic。
func (a *App) notifySecondInstance(args []string, workDir string) {
	if a.ctx == nil {
		slog.Warn("应用上下文尚未就绪，二次启动参数未推送", "args", args)
		return
	}
	runtime.EventsEmit(a.ctx, event.Name("app", ActionSecondInstance), SecondInstancePayload{
		Args:             args,
		WorkingDirectory: workDir,
	})
}

// gen:bind
