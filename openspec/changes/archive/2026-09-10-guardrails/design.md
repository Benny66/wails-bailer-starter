# guardrails — 技术设计

## Context

宪法（AGENTS.md）立了"单一真相""护栏感知自己瞎了""依赖登记制"等铁律，标了 ⚙️ 但未实现。本 change 把它们变成会红的检查。

Wails v2 分层：`renderer (Vue) ──bindings──▶ App 方法 (app.go) ──▶ service ──▶ model`。没有 preload 中间层，所以"IPC 通道名登记"不适用，改为"绑定方法名登记"。

## Goals / Non-Goals

**Goals:**
- Go AST 护栏随 `go test` 跑，违反即红。
- ESLint 自定义规则管前端 import 安全与禁硬编码。
- deps.yaml 双向校验。
- `make lint` 真实执行。

**Non-Goals:**
- 不做完整静态分析（那是 golangci-lint 的活，`make lint` 里调它即可）。
- 不做运行时权限系统（纯工具无鉴权）。

## Decisions

### D1: Go 护栏用 `go/parser + go/ast`，放 `internal/guard/`，用普通 `go test` 承载
护栏就是测试文件（`*_test.go`），解析源码目录做断言。零额外依赖（go/ast 标准库）。护栏第一条判断：`if len(parsed) == 0 { t.Fatal("护栏解析到 0 个结果，写法可能已变更") }`。

### D2: 三条 Go 护栏规则
1. **分层越界**：`app.go` 的绑定方法文件不得 `import gorm`/`internal/database`（DB 走 service）；`model` 包文件只含结构体定义，无业务逻辑（启发式：不 import 除 gorm 外的内部包）。
2. **模型注册双向校验**：解析 `internal/model/` 里所有带 `BaseModel` 内嵌的结构体集合 A；解析 `AllModels()` 返回切片里登记的集合 B；断言 A ⊆ B 且 B ⊆ A（防漏登记 + 防僵尸条目）。
3. **绑定方法命名**：`app.go` 导出的方法名，与 `frontend/wailsjs/go/` 生成的绑定文件一致（防手改绑定）。

### D3: ESLint 自定义规则（flat config）
- `no-node-imports`：`frontend/src/` 禁 `import` `node:*`、`fs`、`child_process`、`path` 等（必须走 bindings）。
- `no-hardcoded-brand`：禁硬编码 hex 色值（应引 tokens）与品牌字符串。

### D4: deps.yaml 双向校验，Go 与前端分开解析
- 正向：`go.mod` 的 `require`（非 indirect）+ `package.json` 的 `dependencies`，每条都必须在 `deps.yaml` 有登记（附理由）。
- 反向：`deps.yaml` 每条都必须真实存在于二者之一（防僵尸/拼写）。
- 校验器用 Go 写（放 `internal/guard/deps_test.go`），复用同一套 go test 入口。

### D5: `make lint` = ESLint + go vet + 护栏测试
`make lint` 聚合三者；护栏测试其实在 `go test` 里，`lint` 只是再显式跑一遍 guard 包，保证"lint 一条命令全查"。

## Risks / Trade-offs

- **[Risk] AST 护栏对写法变化敏感（注释、import 别名）** → 护栏"感知自己瞎了"已兜底：解析不到就 Fatal，不会静默绿。
- **[Risk] deps.yaml 校验把 wails 等间接依赖也误判为需登记** → 只校验 direct `require`（非 indirect）与 `dependencies`（非 devDependencies）。
- **[Risk] ESLint 规则误伤合法代码** → 提供 `// eslint-disable-next-line` 逃生口，但宪法要求逃生需注释理由。
- **[Risk] 护栏过早引入、规则太少显得形式主义** → 先立 3+2 条真实规则，后续模块增多时增量加规则，护栏框架已就位。
