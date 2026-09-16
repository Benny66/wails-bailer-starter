package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// captureSlog 把默认 logger 临时换成写入内存的 handler，返回缓冲与还原函数。
func captureSlog(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()
	buf := &bytes.Buffer{}
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return buf, func() { slog.SetDefault(old) }
}

// TestWailsAdapterLevelRouting 断言 Wails 的各级别落到对应的 slog 级别，
// 并带 source=wails 标记——排查时要能一眼区分「前端/框架」与「Go 业务」。
func TestWailsAdapterLevelRouting(t *testing.T) {
	cases := []struct {
		name      string
		call      func(WailsAdapter)
		wantLevel string
	}{
		{"Print 归入 DEBUG", func(a WailsAdapter) { a.Print("打印") }, "level=DEBUG"},
		{"Trace 归入 DEBUG", func(a WailsAdapter) { a.Trace("跟踪") }, "level=DEBUG"},
		{"Debug 归入 DEBUG", func(a WailsAdapter) { a.Debug("调试") }, "level=DEBUG"},
		{"Info 归入 INFO", func(a WailsAdapter) { a.Info("信息") }, "level=INFO"},
		{"Warning 归入 WARN", func(a WailsAdapter) { a.Warning("警告") }, "level=WARN"},
		{"Error 归入 ERROR", func(a WailsAdapter) { a.Error("错误") }, "level=ERROR"},
		{"Fatal 归入 ERROR", func(a WailsAdapter) { a.Fatal("致命") }, "level=ERROR"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf, restore := captureSlog(t)
			defer restore()

			c.call(WailsAdapter{})
			out := buf.String()

			if !strings.Contains(out, c.wantLevel) {
				t.Errorf("输出 %q 未包含 %s", out, c.wantLevel)
			}
			if !strings.Contains(out, "source=wails") {
				t.Errorf("输出 %q 缺少 source=wails 标记", out)
			}
		})
	}
}

// TestWailsAdapterFatalDoesNotExit Fatal 只记日志、不退出进程。
//
// 这是刻意的：Wails 在调用本方法后会自行 os.Exit(1)，若这里也退，
// defer（日志 flush、数据库连接关闭）将来不及执行。
// 本测试能跑完本身就是断言——真退了就是 fail。
func TestWailsAdapterFatalDoesNotExit(t *testing.T) {
	buf, restore := captureSlog(t)
	defer restore()

	WailsAdapter{}.Fatal("不应导致进程退出")

	if !strings.Contains(buf.String(), "不应导致进程退出") {
		t.Error("Fatal 的消息未写进日志")
	}
}

// TestWailsAdapterKeepsMessageVerbatim 消息原样透传（Wails 已做过格式化），
// 适配器不得再套 Format，否则含 % 的前端日志会被二次解释。
func TestWailsAdapterKeepsMessageVerbatim(t *testing.T) {
	buf, restore := captureSlog(t)
	defer restore()

	WailsAdapter{}.Info("进度 100% 完成")

	if !strings.Contains(buf.String(), "进度 100% 完成") {
		t.Errorf("消息被改动：%q", buf.String())
	}
}
