package apperr

import (
	"encoding/json"
	"errors"
	"testing"
)

// TestFormatReturnsJSONString 是本包最关键的契约：Format MUST 返回字符串而非对象，
// 否则前端 new Error(payload) 会退化成 "[object Object]"。
func TestFormatReturnsJSONString(t *testing.T) {
	out := Format(NotFound("记录不存在"))

	s, ok := out.(string)
	if !ok {
		t.Fatalf("Format MUST 返回 string，实际 %T —— 返回对象会让前端拿到 [object Object]", out)
	}

	var got Error
	if err := json.Unmarshal([]byte(s), &got); err != nil {
		t.Fatalf("Format 返回的字符串不是合法 JSON: %v（原值 %q）", err, s)
	}
	if got.Code != CodeNotFound {
		t.Errorf("期望 code=%q，实际 %q", CodeNotFound, got.Code)
	}
	if got.Message != "记录不存在" {
		t.Errorf("期望 message=记录不存在，实际 %q", got.Message)
	}
}

// TestWrapUnknownError 未知错误 MUST 兜底为 internal，且不把英文堆栈当 message 展示。
func TestWrapUnknownError(t *testing.T) {
	raw := errors.New("sqlite: no such table")
	got := Wrap(raw)

	if got.Code != CodeInternal {
		t.Errorf("未知错误期望 code=%q，实际 %q", CodeInternal, got.Code)
	}
	if got.Message != "系统错误" {
		t.Errorf("未知错误 message 应为中文兜底，实际 %q", got.Message)
	}
	if got.Detail != "sqlite: no such table" {
		t.Errorf("原始错误应保留在 Detail，实际 %q", got.Detail)
	}
}

// TestWrapPreservesBusinessError 已是 *Error 的错误 MUST 原样返回，不被兜底覆盖。
func TestWrapPreservesBusinessError(t *testing.T) {
	orig := NotFound("资产不存在")
	got := Wrap(orig)

	if got != orig {
		t.Errorf("业务错误应原样返回，实际被替换为 %+v", got)
	}
}

// TestWrapWrappedError 用 %w 包裹的业务错误 MUST 经 errors.As 解出。
func TestWrapWrappedError(t *testing.T) {
	orig := NotFound("资产不存在")
	got := Wrap(errors.Join(errors.New("上下文"), orig))

	if got.Code != CodeNotFound {
		t.Errorf("包裹的业务错误应被解出，期望 code=%q，实际 %q", CodeNotFound, got.Code)
	}
}

// TestWrapNil nil 输入返回 nil（不 panic）。
func TestWrapNil(t *testing.T) {
	if got := Wrap(nil); got != nil {
		t.Errorf("Wrap(nil) 应返回 nil，实际 %+v", got)
	}
}
