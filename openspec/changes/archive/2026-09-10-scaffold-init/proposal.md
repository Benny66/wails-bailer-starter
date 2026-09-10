# scaffold-init — 脚手架实例化入口（母版占位符 + init 脚本）

## Why

脚手架已交付全部能力，但"从脚手架生成新项目"这一步仍是空白——模块名 `wails-bailer-starter` 硬编码散落在 7 处文件（go.mod / wails.json / main.go / app.go / internal/database / _example 模板 / index.html），手改必然漏改。复用型基座需要像 `wails init` 一样的实例化入口：母版用占位符，`scripts/init.sh <name>` 一键生成干净新项目。

## What Changes

- **母版占位符化**：把模块名/品牌名 `wails-bailer-starter` 全部替换为占位符 `__APP_NAME__`（Go 模块名、import 路径、wails.json、appName 常量、options.Title、UniqueId、index.html title、_example 模板 import）。
- **`scripts/init.sh <name>` 实例化脚本**：
  1. `git clone` 母版（天然排除 `.git`、`.claude/settings.local.json`、node_modules、build 产物）。
  2. 全局替换 `__APP_NAME__` → `<name>`（Go import、go.mod、wails.json、index.html、_example 模板）。
  3. 清空 `openspec/changes/archive`（新项目从零开始治理）。
  4. 重新 `git init` + 首次提交。
  5. 打印"下一步"指引（`make dev` / `make gen name=xxx`）。
- **验证**：`init.sh` 跑完后 `make build` 必须通过，证明替换无遗漏。

## Capabilities

### New Capabilities

- `scaffold-init`: 母版占位符化 + `scripts/init.sh` 实例化脚本（一键生成干净新项目）。

### Modified Capabilities

<!-- 无 spec 级变化，纯新增实例化能力 -->

## Impact

- **修改代码**：7 处文件的模块名 → `__APP_NAME__` 占位符。
- **新增代码**：`scripts/init.sh`。
- **风险**：占位符化后，母版自身 `make dev`/`make build` 会因模块名 `__APP_NAME__` 无法编译——需验证 init 生成的项目可编译，母版本身不再是"可直接跑"状态，而是"母版 + 实例化后才可跑"。
- **破坏性**：母版不能再直接 `make dev`（模块名是占位符）。这是脚手架化的代价，需在 README 明确。
