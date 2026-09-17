package guard

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

// TestVersionInjectionTargetsExist 断言构建脚本注入的符号在 appinfo 包真实存在。
//
// 注入点唯一：scripts/build.sh（make build / package.sh / release.sh 都经它构建）。
// 这本身就是「版本号单一真相」的一部分——多处各算各的版本是双源漂移的温床。
func TestVersionInjectionTargetsExist(t *testing.T) {
	scripts := []string{"scripts/build.sh"}

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

// ---------------------------------------------------------------------------
// 护栏 10：版本号与 bundle id 的单一真相不得失守
//
// 这两个字段都会「悄悄失效」：
//   - wails.json 缺 info.productVersion → wails 用默认值 "1.0.0"，
//     产物里写死 1.0.0，且不报任何错（本仓曾长期如此：系统显示 1.0.0、
//     日志显示 dev，两者都对不上）；
//   - build/darwin/Info.plist 沿用模板的 com.wails.<name> → 等于宣称 Wails 是厂商，
//     且与其他 Wails 项目共享 bundle id 命名空间。
// ---------------------------------------------------------------------------

// TestWailsJSONHasVersionSourceOfTruth 断言 wails.json 声明了可供产物渲染的版本与产品名。
func TestWailsJSONHasVersionSourceOfTruth(t *testing.T) {
	path := filepath.Join(projectRoot(), "wails.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 wails.json 失败: %v", err)
	}
	var cfg struct {
		Info struct {
			ProductName    string `json:"productName"`
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("解析 wails.json 失败: %v", err)
	}

	if cfg.Info.ProductVersion == "" {
		t.Error("wails.json 未声明 info.productVersion：wails 会静默用默认值 1.0.0，" +
			"产物里的版本从此与代码无关（macOS/Windows 版本资源都由它渲染）")
	} else if !numericVersionRe.MatchString(cfg.Info.ProductVersion) {
		t.Errorf("wails.json 的 info.productVersion = %q 不是数字点分格式："+
			"NSIS 的 VIProductVersion 要求数字版本，Windows 打包会在最后一步失败",
			cfg.Info.ProductVersion)
	}
	if cfg.Info.ProductName == "" {
		t.Error("wails.json 未声明 info.productName：macOS 的 CFBundleName 会为空")
	}
}

// numericVersionRe 匹配数字点分版本（如 0.1.0 / 1.2.3.4）。
var numericVersionRe = regexp.MustCompile(`^[0-9]+(\.[0-9]+){2,3}$`)

// TestBundleIDIsProjectControlled 断言 bundle id 不沿用 wails 模板的 com.wails.* 前缀。
func TestBundleIDIsProjectControlled(t *testing.T) {
	path := filepath.Join(projectRoot(), "build", "darwin", "Info.plist")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 build/darwin/Info.plist 失败: %v", err)
	}
	content := string(data)

	idx := strings.Index(content, "CFBundleIdentifier")
	if idx < 0 {
		t.Fatal("Info.plist 中找不到 CFBundleIdentifier——写法可能已变更，请同步更新护栏解析规则")
	}
	// 取该键之后的一小段文本，只看它对应的取值行
	rest := content[idx:]
	if end := strings.Index(rest, "</string>"); end > 0 {
		rest = rest[:end]
	}
	if strings.Contains(rest, "com.wails.") {
		t.Error("Info.plist 的 CFBundleIdentifier 仍是模板默认的 com.wails.<name>——" +
			"等于把 Wails 当厂商，且与其他 Wails 项目共享 bundle id 命名空间。" +
			"应使用 scripts/init.sh 生成的 __BUNDLE_PREFIX__.<name>（默认 com.example）")
	}
	if !strings.Contains(rest, "{{safeBundleID .Name}}") {
		t.Error("Info.plist 的 CFBundleIdentifier 未使用 {{safeBundleID .Name}} 派生应用名——" +
			"多个产品会撞同一个 bundle id")
	}
}

// ---------------------------------------------------------------------------
// 护栏 11：可观测性接线不得静默缺失
//
// 这类失效的共同点：删掉不会报错、不会崩溃，只是**失去可见性**——
// 慢查询回到无人可见的 stdout、启动优化重新变成盲调。症状要等到真出问题时才显形，
// 而那时正是最需要日志的时候。故只能靠检查兜住。
// ---------------------------------------------------------------------------

// TestGormLoggerIsWired 断言数据库连接配置了日志实现。
func TestGormLoggerIsWired(t *testing.T) {
	files, err := parseDir(filepath.Join(projectRoot(), "internal", "database"))
	if err != nil {
		t.Fatalf("解析 internal/database 失败: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("护栏在 internal/database 未解析到任何 .go 文件——写法可能已变更，请同步更新护栏解析规则")
	}

	lit := findCompositeLiteral(files, "gorm", "Config")
	if lit == nil {
		t.Fatal("护栏未找到 gorm.Config{...} 字面量——写法可能已变更，请同步更新护栏解析规则")
	}

	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "Logger" {
				return // 已接线
			}
		}
	}
	t.Error("gorm.Config 未设置 Logger：gorm 默认 logger 写 stdout，" +
		"打包后的 GUI 应用无人可见，慢查询与 SQL 错误会静默消失")
}

// TestStartupTimelineStagesExist 断言组合根的启动阶段打点覆盖关键阶段。
//
// 删掉打点不会有任何症状，只是启动优化重新变成盲调（此前只有首尾两点，
// 实测 359–733ms 的中间部分是黑盒）。
func TestStartupTimelineStagesExist(t *testing.T) {
	files, err := parseDir(projectRoot())
	if err != nil {
		t.Fatalf("解析项目根目录失败: %v", err)
	}
	f, ok := files["main.go"]
	if !ok {
		t.Fatal("护栏找不到 main.go——写法可能已变更，请同步更新护栏解析规则")
	}

	// 期望覆盖的阶段（改名即失败，提示同步护栏）
	want := []string{"日志初始化", "数据库初始化", "配置加载与组装"}
	found := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "stage" {
			return true
		}
		if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if name, uerr := strconv.Unquote(lit.Value); uerr == nil {
				found[name] = true
			}
		}
		return true
	})

	if len(found) == 0 {
		t.Fatal("护栏在 main.go 未解析到任何启动阶段打点——写法可能已变更，请同步更新护栏解析规则")
	}
	for _, w := range want {
		if !found[w] {
			t.Errorf("启动时间线缺少阶段 %q：删掉打点不会有任何症状，"+
				"只会让启动优化重新变成盲调（现有阶段: %v）", w, found)
		}
	}
}

// findCompositeLiteral 在给定文件集合中定位 `pkg.Type{...}` 形式的复合字面量。
func findCompositeLiteral(files map[string]*ast.File, pkg, typeName string) *ast.CompositeLit {
	var found *ast.CompositeLit
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			if found != nil {
				return false
			}
			cl, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			sel, ok := cl.Type.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != pkg || sel.Sel.Name != typeName {
				return true
			}
			found = cl
			return false
		})
	}
	return found
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
