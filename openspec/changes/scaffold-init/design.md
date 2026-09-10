# scaffold-init — 技术设计

## Context

脚手架能力已齐全，但实例化靠手改 7 处模块名，易漏。目标：母版占位符化 + `scripts/init.sh` 一键实例化。

关键事实（已实测）：
- `git clone` 天然排除 `.git`、`.claude/settings.local.json`（全局 gitignore `**/.claude/settings.local.json`）、node_modules、build 产物。
- 应保留的治理资产会被 clone 带走：`.claude/commands` + `.claude/skills`、`openspec/`、`AGENTS.md`、`CLAUDE.md`、`deps.yaml`、`Makefile`、`_example/`。
- 需替换模块名的文件：go.mod、wails.json、main.go、app.go、internal/database/database.go、_example/service/example_service.go、frontend/index.html。

## Goals / Non-Goals

**Goals:**
- 母版所有 `wails-bailer-starter` 占位符化为 `__APP_NAME__`。
- `init.sh <name>` 从母版生成可编译的新项目，替换零遗漏。
- 新项目保留 OpenSpec 治理结构（空 archive + specs 作为"已实现的能力基线"）。

**Non-Goals:**
- 不做独立 CLI/npm 包（方案 B，阶段过重）。
- 不处理品牌名以外的深度定制（logo/主色仍手改，init 只处理模块名）。

## Decisions

### D1: 占位符用 `__APP_NAME__`（双下划线包裹，避免误替换）
`__APP_NAME__` 是 Go 标识符里的合法占位符形态，且全局唯一、不易与业务代码冲突。init 时 `sed 's/__APP_NAME__/<name>/g'`。

**备选**：`{APP_NAME}` / `{{APP_NAME}}`——否决，Go import 路径里 `{` 不合法，`__APP_NAME__` 更安全。

### D2: init 用 `git clone` 而非 `cp -R`
`git clone` 天然排除私货（settings.local.json）和 `.git` 历史，比手动维护排除清单更可靠。init 后 `rm -rf .git` 再 `git init`。

### D3: 母版不再"可直接跑"，是"母版 + 实例化后可跑"
占位符化后母版自身 `make dev` 会因 `__APP_NAME__` 模块名无法编译。这是脚手架化的固有代价（create-vite 的模板本身也不直接跑）。README 明确"用 init 实例化后开发"。

### D4: openspec 归档历史在 init 时清空，specs 保留
- 清空 `openspec/changes/archive`（这是母版的开发历史，新项目不要）。
- 保留 `openspec/specs`（16 个 capability 是"基座已实现的能力基线"，新项目的护栏/治理依赖它们，且是"哪些能力已就绪"的权威清单）。

### D5: 替换清单显式列出，不用 `grep -r` 盲扫
替换文件清单在 init.sh 里硬编码（与 gen.sh 的锚点清单同思路），避免误替换 node_modules 等。文件名清单是单一真相，与护栏可断言一致。

## Risks / Trade-offs

- **[Risk] 占位符化后母版护栏跑不过（go vet / go test 因 __APP_NAME__ 失败）** → 母版护栏本就该在"实例化后的项目"里跑；母版仓库的 CI 只跑"init 自测"（init 一个临时名 → build 通过），不跑母版自身的 go test。
- **[Risk] 某处模块名没被占位符化，init 后残留 `__APP_NAME__`** → init.sh 末尾加 `grep -rn '__APP_NAME__'` 断言残留为空，非空则报错。
- **[Risk] init 目标名不合法（含空格/大写/Go 关键字）** → init.sh 校验 name 匹配 `^[a-z][a-z0-9_-]*$`，不合法则退出。
- **[Risk] 母版 README 里"改名需同步 4 处"的说法过期** → 本 change 同步更新 README 的换肤章节（改为"用 init.sh 实例化"）。
