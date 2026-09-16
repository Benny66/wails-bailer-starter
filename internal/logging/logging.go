// Package logging 负责文件 + 控制台双写日志，按大小轮转。
//
// 使用 lumberjack 做文件轮转（按大小 + 保留份数），标准库 log/slog 做分级。
// 开发态 Debug 级，生产态 Info 级（由调用方在 Init 时指定）。
// 日志文件写在应用数据目录下（见 internal/appdir），与数据库、配置同目录。
package logging

import (
	"io"
	"log/slog"
	"os"
	"runtime"

	"gopkg.in/natefinch/lumberjack.v2"

	"__APP_NAME__/internal/appdir"
)

// Options 是日志初始化参数。
type Options struct {
	AppName string
	// Level 为最低输出级别（Debug/Info/Warn/Error）。
	Level slog.Level
	// Version 版本号，写入启动首行——使【每个】日志文件自带版本上下文，
	// 用户报障时不必再问「你用的是哪个版本」。
	Version string
}

// Init 初始化全局 logger，返回关闭函数（优雅关闭时调用以 flush 文件）。
// 同时输出到控制台（stderr）与轮转文件。
func Init(opts Options) (closeFn func(), err error) {
	logPath, err := appdir.File(opts.AppName, "app.log")
	if err != nil {
		return nil, err
	}

	// lumberjack：单文件 5MB，保留 3 份，最多 7 天。
	fileWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    5, // MB
		MaxBackups: 3,
		MaxAge:     7, // days
		Compress:   false,
	}

	// 控制台 + 文件双写
	multi := io.MultiWriter(os.Stderr, fileWriter)
	handler := slog.NewTextHandler(multi, &slog.HandlerOptions{Level: opts.Level})
	slog.SetDefault(slog.New(handler))

	// 启动首行：版本 + 平台 + 日志自身所在路径。
	// 记日志路径是为了「从日志本身找到日志在哪」——排查时第一个要问的问题。
	slog.Info("应用启动",
		"version", opts.Version,
		"platform", runtime.GOOS+"/"+runtime.GOARCH,
		"log_file", logPath,
	)

	return func() { _ = fileWriter.Close() }, nil
}
