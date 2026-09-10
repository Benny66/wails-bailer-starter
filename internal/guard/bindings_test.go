package guard

import (
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// 护栏 3：绑定方法名一致
//   - app.go 里 App 结构体的每个导出方法，都必须有对应的 wails 绑定生成文件
//     （frontend/wailsjs/go/main/App.d.ts），防手改绑定导致 Go↔TS 漂移。
//   - 合法空态：App 尚无导出方法时，护栏通过（不是"瞎了"，因为已确认解析到
//     app.go 与 App 结构体）。
// ---------------------------------------------------------------------------

func TestBindingMethodsHaveGeneratedBindings(t *testing.T) {
	files, err := parseDir(projectRoot())
	if err != nil {
		t.Fatalf("解析项目根目录失败: %v", err)
	}
	f, ok := files[bindingsFile]
	if !ok {
		t.Fatalf("护栏找不到 %s——写法可能已变更，请同步更新护栏解析规则", bindingsFile)
	}

	exported := exportedAppMethods(f)
	if len(exported) == 0 {
		// 合法空态：尚无绑定方法
		return
	}

	bindingPath := filepath.Join(projectRoot(), "frontend", "wailsjs", "go", "main", "App.d.ts")
	data, err := os.ReadFile(bindingPath)
	if err != nil {
		t.Fatalf("app.go 有 %d 个导出绑定方法，但找不到生成的绑定文件 %s（请运行 make dev/build 重新生成 bindings）", len(exported), bindingPath)
	}

	content := string(data)
	for _, m := range exported {
		// App.d.ts 中每个方法声明形如 `methodName(...): ...`，检查方法名出现。
		// 用更稳的锚点：`methodName(` 后跟参数。
		if !strings.Contains(content, m+"(") {
			t.Errorf("绑定方法 %s 未在生成的 App.d.ts 中找到——bindings 可能过期，请重新生成", m)
		}
	}
}

// exportedAppMethods 收集 App 结构体的导出方法名（大写开头，且排除 startup 等）。
func exportedAppMethods(f *ast.File) []string {
	var out []string
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
			continue
		}
		// 接收者类型是否为 *App
		recv := fd.Recv.List[0].Type
		var recvName string
		switch t := recv.(type) {
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				recvName = ident.Name
			}
		case *ast.Ident:
			recvName = t.Name
		}
		if recvName != "App" {
			continue
		}
		// 导出方法 = 首字母大写
		name := fd.Name.Name
		if name[0] >= 'A' && name[0] <= 'Z' {
			out = append(out, name)
		}
	}
	return out
}
