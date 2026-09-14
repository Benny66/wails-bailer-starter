// Package apperr 定义管道契约中的「错误协议」。
//
// 契约：绑定方法（app.go 的 App 方法）返回的业务错误 MUST 为 *Error，
// 携带机器可读的 code 与面向用户的中文 message。前端据 code 分流，不解析文案。
//
// 序列化约定（关键，勿改）：
//   - main.go 把 apperr.Format 注入 options.App.ErrorFormatter。
//   - Format 返回【JSON 字符串】而非对象——前端以 new Error(payload) 包装错误，
//     传对象会被强转成 "[object Object]"，结构化字段全部丢失（Wails v2.15.0 实测）。
//   - 前端 frontend/src/lib/invoke.ts 收到后 JSON.parse 还原为 AppError。
package apperr

import (
	"encoding/json"
	"errors"
)

// 错误码常量。新增错误码在此登记（单一真相），前端据其分流。
const (
	// CodeNotFound 记录不存在。
	CodeNotFound = "not_found"
	// CodeValidation 入参校验失败。
	CodeValidation = "validation"
	// CodeConflict 状态冲突（如重复创建）。
	CodeConflict = "conflict"
	// CodeInternal 未归类的系统错误（兜底）。
	CodeInternal = "internal"
)

// Error 是业务错误的统一形态。
// Code 供前端分流，Message 面向用户（中文，可直接展示），Detail 仅排查用不展示。
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// Error 实现 error 接口，返回面向用户的中文消息。
func (e *Error) Error() string { return e.Message }

// New 构造业务错误。
func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// NotFound 记录不存在的业务错误。
func NotFound(message string) *Error {
	return &Error{Code: CodeNotFound, Message: message}
}

// Validation 入参校验失败的业务错误。
func Validation(message string) *Error {
	return &Error{Code: CodeValidation, Message: message}
}

// Conflict 状态冲突的业务错误。
func Conflict(message string) *Error {
	return &Error{Code: CodeConflict, Message: message}
}

// Wrap 把任意错误归一化为 *Error：已是 *Error 原样返回，
// 否则兜底为 internal（Message 为中文兜底文案，Detail 保留原始错误供排查）。
func Wrap(err error) *Error {
	if err == nil {
		return nil
	}
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	return &Error{Code: CodeInternal, Message: "系统错误", Detail: err.Error()}
}

// Format 是 Wails options.App.ErrorFormatter 的实现：把 error 序列化为 JSON 字符串。
//
// 【必须返回字符串，不得返回对象】——前端 new Error(payload) 会把对象转成
// "[object Object]"。返回字符串则前端可 JSON.parse 还原。
func Format(err error) any {
	ae := Wrap(err)
	if ae == nil {
		return ""
	}
	b, mErr := json.Marshal(ae)
	if mErr != nil {
		// 序列化失败（理论上不会）：退回纯文本，前端归一化为 internal。
		return ae.Message
	}
	return string(b)
}
