# guardrails — 架构护栏与依赖登记制

## Why

地基立了"单一真相""护栏感知自己瞎了"等宪法铁律，但没有机器执行——这些还是软约束，AI 生成代码时一定会被违反。本 change 把约定编译成会红的检查，让违规在 `make test` / `make lint` 阶段暴露。

## What Changes

- **Go 侧架构护栏**：用 `go/parser + go/ast` 解析源码做结构化断言，放在 `internal/guard/` 包，随 `go test` 跑。护栏必须"感知自己瞎了"——解析到 0 个结果要 Fatal。
- **护栏规则**：分层越界（App 绑定方法不得直接 `gorm`；model 不写业务逻辑）、模型注册双向校验（带 `BaseModel` 的结构体必须登记进 `AllModels()`，且登记项真实存在）。
- **前端 ESLint 护栏**：自定义规则，`renderer` 禁 `import` `node:*`/`fs`（必须走 bindings），禁硬编码 hex 色值/品牌字符串。
- **依赖登记制 `deps.yaml`**：新增依赖必须登记 + 附理由，护栏做双向校验（正向：清单里直接依赖必须登记；反向：登记项必须真实存在）。Go 依赖解析 `go.mod`，前端解析 `package.json`。
- **Makefile 接线**：`make lint` 从占位改为真实执行（ESLint + go vet + 护栏测试）。

## Capabilities

### New Capabilities

- `architecture-guardrails`: Go AST 护栏 + ESLint 自定义规则，分层越界/模型注册/前端 import 安全的三类断言，且护栏必须"感知自己瞎了"。
- `dependency-registry`: `deps.yaml` 依赖登记制与双向校验。

### Modified Capabilities

- `project-scaffold`: `make lint` 从"未实现占位"改为"真实执行静态检查"（原 foundation 里 lint 是占位 exit 1）。

## Impact

- **新增代码**：`internal/guard/`（Go 护栏 + 测试）、`frontend/eslint.config.js` + 自定义规则、`deps.yaml`。
- **依赖**：eslint 及插件（前端 devDependencies）、golang.org/x/tools（go/ast，可能）。
- **宪法同步**：`AGENTS.md` 中标注"待 guardrails change"的 ⚙️ 规则，改为指向真实护栏。
- **破坏性**：`make lint` 行为变化（占位→真实），此前 `exit 1` 的占位被替换。
