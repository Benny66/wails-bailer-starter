package logging

import "log/slog"

// WailsAdapter 把 Wails 的日志接口接到 slog，使前端与框架内部的日志落到同一个文件。
//
// 为什么必须有它：Wails 的日志链路是
//
//	JS LogError(msg) → WailsInvoke → Go 内部 logger → options.App.Logger
//
// options.App.Logger 为空时 Wails 用默认 logger【只写 stdout】——打包后的 GUI 应用
// stdout 无人可见，于是前端的 LogError 与 Wails 自身的内部错误（资源加载失败、
// 前端异常）全部石沉大海。这是一处【没有任何症状】的缺失：不接也不会报错，
// 直到用户丢日志来报障时才发现日志里什么都没有。故 main.go 的接线由
// internal/guard/wiring_test.go 强制。
//
// 本类型不 import wails：Go 接口是结构化满足的，main.go 直接赋值即可
// （避免 logging 包为一个日志适配器拖进整个 wails 依赖）。
type WailsAdapter struct{}

// sourceWails 标记日志来源，排查时一眼区分「前端/框架」与「Go 业务」。
const sourceWails = "wails"

// Print Wails 的无级别输出（多为启动细节），按 Debug 处理。
func (WailsAdapter) Print(message string) { slog.Debug(message, "source", sourceWails) }

// Trace Wails 的 trace 级输出（slog 无 trace，归入 Debug）。
func (WailsAdapter) Trace(message string) { slog.Debug(message, "source", sourceWails) }

// Debug 调试信息。
func (WailsAdapter) Debug(message string) { slog.Debug(message, "source", sourceWails) }

// Info 常规信息。
func (WailsAdapter) Info(message string) { slog.Info(message, "source", sourceWails) }

// Warning 警告。
func (WailsAdapter) Warning(message string) { slog.Warn(message, "source", sourceWails) }

// Error 错误。
func (WailsAdapter) Error(message string) { slog.Error(message, "source", sourceWails) }

// Fatal 致命错误。
//
// 【刻意不调用 os.Exit】：Wails 在调用本方法后会自行 os.Exit(1)，
// 这里再退一次会让 defer（日志 flush、数据库连接关闭）来不及执行。
func (WailsAdapter) Fatal(message string) { slog.Error(message, "source", sourceWails) }
