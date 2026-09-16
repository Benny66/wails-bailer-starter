package main

import (
	"context"
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"__APP_NAME__/internal/appdir"
	"__APP_NAME__/internal/apperr"
	"__APP_NAME__/internal/config"
	"__APP_NAME__/internal/dialog"
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
	// 2. 落盘配置
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

// gen:bind
