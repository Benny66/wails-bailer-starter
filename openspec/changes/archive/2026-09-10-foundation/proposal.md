# foundation — 脚手架地基

## Why

`wails-bailer-starter` 是一个**要被反复复用的 Wails v2 (Go + Vue3) 模板基座**，但目前仓库只有三份设计文档、没有一行代码。后序所有能力（设计系统、护栏、运行时服务、示例模块、CI）都要长在一个能编译、能跑、结构正确的地基上——地基不先立稳，后面每个 change 都会踩在流沙上。

## What Changes

- **初始化 Wails v2 最小可跑骨架**：`wails init` 生成的 Go 后端 + Vue3 前端 + 标准目录结构，`make dev` 能起、`make package` 能打包（跨平台）。
- **接入 SQLite 持久化基座**：gorm 连接 + AutoMigrate 机制 + 空迁移注册表（本 change 不建业务表，只立"迁移框架"这一真相）。
- **立宪法文档体系**：根 `AGENTS.md`（跨端通用铁律）+ `frontend/CLAUDE.md`（Vue 规范）+ `backend/CLAUDE.md`（Go 规范）+ `docs/map.md`（代码导航地图），机器可强制规则标 `⚙️` 并指向对应检查。
- **统一命令入口 `Makefile`**：`make dev / build / test / lint / smoke / package / gen`，AI 与人只记 target，不记零散脚本路径。
- **`.gitignore` 预配置**：挡住运行时产物（`*.db`、`node_modules/`、`dist/`、`build/`、`*.exe`、日志等）。

## Capabilities

### New Capabilities

- `project-scaffold`: Wails v2 (Go + Vue3) 标准目录结构、最小可跑骨架、统一 `Makefile` 命令入口、`.gitignore` 干净性约定。
- `constitution`: 宪法文档体系——根 `AGENTS.md` + 域 `CLAUDE.md`（frontend/backend）+ `docs/map.md`，含 ⚙️ 机器强制规则标记约定。
- `data-persistence`: SQLite + gorm 连接、AutoMigrate 迁移框架、模型注册真相（本 change 仅立机制，无业务表）。

### Modified Capabilities

<!-- 无 —— 全新项目，无既有 capability 需要改需求 -->

## Impact

- **新增代码**：Go 后端骨架（`backend/` 或 Wails 约定目录）、Vue3 前端骨架（`frontend/`）、`Makefile`、`.gitignore`。
- **新增文档**：`AGENTS.md`、`frontend/CLAUDE.md`、`backend/CLAUDE.md`、`docs/map.md`。
- **依赖**：Go 模块（wails v2、gorm、SQLite 驱动）、Node 依赖（vue3、vite、element-plus、vue-router、pinia）。
- **无破坏性变更**：全新项目，无既有代码受影响。
