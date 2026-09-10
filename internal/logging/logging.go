// Package logging 负责文件 + 控制台双写日志，按大小轮转。
//
// 使用 lumberjack 做文件轮转（按大小 + 保留份数），标准库 log/slog 做分级。
// 开发态 Debug 级，生产态 Info 级（由调用方在 Init 时指定）。
package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Options 是日志初始化参数。
type Options struct {
	AppName string
	// Level 为最低输出级别（Debug/Info/Warn/Error）。
	Level slog.Level
}

// Init 初始化全局 logger，返回关闭函数（优雅关闭时调用以 flush 文件）。
// 同时输出到控制台（stderr）与轮转文件。
func Init(opts Options) (closeFn func(), err error) {
	logPath, err := logPath(opts.AppName)
	if err != nil {
		return nil, err
	}

	// lumberjack：单文件 5MB，保留 3 份，最多 7 天。
	fileWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    5,  // MB
		MaxBackups: 3,
		MaxAge:     7, // days
		Compress:   false,
	}

	// 控制台 + 文件双写
	multi := io.MultiWriter(os.Stderr, fileWriter)
	handler := slog.NewTextHandler(multi, &slog.HandlerOptions{Level: opts.Level})
	slog.SetDefault(slog.New(handler))

	return func() { _ = fileWriter.Close() }, nil
}

// logPath 返回日志文件路径，并确保目录存在。
func logPath(appName string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(dir, appName)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "app.log"), nil
}
