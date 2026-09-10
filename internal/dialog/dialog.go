// Package dialog 封装 Wails runtime 的原生对话框，暴露为前端可调用的形式。
//
// 运行时方法需要 context.Context（App 启动时保存的 ctx），故这些函数
// 由 App 的绑定方法调用并传入 ctx。
package dialog

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SelectFile 打开"选择文件"对话框，返回所选路径（取消返回空串）。
func SelectFile(ctx context.Context, title string) (string, error) {
	return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{Title: title})
}

// SelectDirectory 打开"选择目录"对话框，返回所选路径（取消返回空串）。
func SelectDirectory(ctx context.Context, title string) (string, error) {
	return runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{Title: title})
}

// SelectSaveFile 打开"保存文件"对话框，返回目标路径（取消返回空串）。
func SelectSaveFile(ctx context.Context, title, defaultFilename string) (string, error) {
	return runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
	})
}

// Message 弹出消息框，返回用户点击的按钮文本（"Yes"/"No"/"OK" 等）。
func Message(ctx context.Context, title, message string) (string, error) {
	return runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
		Buttons: []string{"OK"},
	})
}

// Confirm 弹出确认框，返回用户是否点击了"Yes"。
func Confirm(ctx context.Context, title, message string) (bool, error) {
	result, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.QuestionDialog,
		Title:   title,
		Message: message,
		Buttons: []string{"Yes", "No"},
	})
	if err != nil {
		return false, err
	}
	return result == "Yes", nil
}
