package guard

import (
	"go/ast"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// 护栏 8：框架接线不得静默缺失
//
// 有些「约定」漏了不会有任何症状——不报错、不崩溃，只是某条链路默默断掉，
// 直到出问题要查日志时才发现日志里什么都没有。这类缺失只能靠检查兜住。
//
// 当前覆盖：
//   - options.App.Logger：不接 Wails logger，前端的 LogError/LogInfo 与 Wails 自身的
//     内部错误【只写 stdout】，打包后无人可见（详见 internal/logging/wails.go）。
//
// 铁律：解析到 0 个目标必须报错（护栏感知自己瞎了）。
// ---------------------------------------------------------------------------

// requiredWailsOptions 是必须出现在 options.App 字面量里的字段及其缺失后果。
var requiredWailsOptions = []struct {
	field  string
	reason string
}{
	{
		field:  "Logger",
		reason: "不设置它，前端日志与 Wails 内部错误只写 stdout（打包后无人可见），日志会静默丢失",
	},
}

// TestWailsOptionsAreWired 断言 main.go 里创建窗口的 options.App 字面量设置了必需字段。
func TestWailsOptionsAreWired(t *testing.T) {
	files, err := parseDir(projectRoot())
	if err != nil {
		t.Fatalf("解析项目根目录失败: %v", err)
	}
	f, ok := files["main.go"]
	if !ok {
		t.Fatalf("护栏找不到 main.go——写法可能已变更，请同步更新护栏解析规则")
	}

	lit := findWailsAppLiteral(f)
	if lit == nil {
		t.Fatal("护栏在 main.go 中未找到 &options.App{...} 字面量——写法可能已变更，请同步更新护栏解析规则")
	}

	present := map[string]bool{}
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if ident, ok := kv.Key.(*ast.Ident); ok {
			present[ident.Name] = true
		}
	}
	if len(present) == 0 {
		t.Fatal("护栏解析到 options.App 字面量但没有任何具名字段——写法可能已变更，请同步更新护栏解析规则")
	}

	for _, req := range requiredWailsOptions {
		if !present[req.field] {
			t.Errorf("main.go 的 options.App 未设置 %s：%s", req.field, req.reason)
		}
	}
}

// ---------------------------------------------------------------------------
// 护栏 9：构建期注入目标必须真实存在
//
// 链接器的 `-X pkg.Symbol=value` 对【不存在的符号是静默忽略】的（已实测：
// 写错符号名时构建成功、无任何警告，变量保持零值，版本悄悄退回 "dev"）。
// 故打包脚本里的注入目标必须由护栏核对，否则改了变量名不会有任何症状，
// 只是发布的版本号从此永远是 "dev"。
// ---------------------------------------------------------------------------

// injectionTargetRe 匹配脚本里的 -X 注入：`-X <module>/internal/appinfo.<Symbol>=`
var injectionTargetRe = regexp.MustCompile(`-X\s+\S*internal/appinfo\.([A-Za-z0-9_]+)=`)

// TestVersionInjectionTargetsExist 断言打包脚本注入的符号在 appinfo 包真实存在。
func TestVersionInjectionTargetsExist(t *testing.T) {
	scripts := []string{"scripts/package.sh", "scripts/release.sh"}

	// 收集 internal/appinfo 里的包级变量名（注入目标必须是可寻址的字符串变量）
	dir := filepath.Join(projectRoot(), "internal", "appinfo")
	files, err := parseDir(dir)
	if err != nil {
		t.Fatalf("解析 internal/appinfo 失败: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("护栏在 %s 未解析到任何 .go 文件——写法可能已变更，请同步更新护栏解析规则", dir)
	}
	vars := map[string]bool{}
	for _, f := range files {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok.String() != "var" {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, n := range vs.Names {
					vars[n.Name] = true
				}
			}
		}
	}
	if len(vars) == 0 {
		t.Fatal("护栏在 internal/appinfo 未解析到任何包级变量——写法可能已变更，请同步更新护栏解析规则")
	}

	checked := 0
	for _, rel := range scripts {
		data, err := os.ReadFile(filepath.Join(projectRoot(), rel))
		if err != nil {
			t.Fatalf("读取 %s 失败: %v", rel, err)
		}
		for _, m := range injectionTargetRe.FindAllStringSubmatch(string(data), -1) {
			checked++
			if !vars[m[1]] {
				t.Errorf("%s 注入 %s，但 internal/appinfo 中无此包级变量——"+
					"链接器会静默忽略，版本将永远退回 dev", rel, m[1])
			}
		}
	}
	if checked == 0 {
		t.Fatalf("护栏未在任何打包脚本中解析到 -X internal/appinfo.* 注入——"+
			"注入被移除或写法已变更，请同步更新护栏解析规则（%v）", scripts)
	}
}

// findWailsAppLiteral 在 main.go 中定位 `&options.App{...}`（wails.Run 的参数）。
func findWailsAppLiteral(f *ast.File) *ast.CompositeLit {
	var found *ast.CompositeLit
	ast.Inspect(f, func(n ast.Node) bool {
		if found != nil {
			return false
		}
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		// 形如 options.App{...}
		sel, ok := cl.Type.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if pkg.Name == "options" && strings.EqualFold(sel.Sel.Name, "App") {
			found = cl
			return false
		}
		return true
	})
	return found
}
