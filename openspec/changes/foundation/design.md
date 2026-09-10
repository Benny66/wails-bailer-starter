# foundation — 技术设计

## Context

目标：从零搭出 `wails-bailer-starter` 的地基——一个 Wails v2 (Go + Vue3) 模板基座，后续要长出设计系统、护栏、运行时服务、示例模块、CI。本 change 只交付"能编译、能跑、结构正确、约定已立"的最小底座，不做任何业务表、不做护栏实现、不做 UI 设计系统（那是后序 change）。

关键约束来自 `docs/脚手架功能说明.md`（技术栈锁定 Wails v2 + Vue3 + Element Plus）与 `docs/脚手架搭建经验.md`（方法论：护栏先于业务、单一真相、宪法分层、Makefile 统一入口、干净性）。本 change 是这些方法论的第一块落地。

## Goals / Non-Goals

**Goals:**
- `wails init` 骨架落地，`make dev` 能起、`make package` 能跨平台打包。
- gorm + SQLite 连接 + AutoMigrate 机制立起来（只有空迁移注册表，无业务表）。
- 宪法四件套（AGENTS.md + frontend/CLAUDE.md + backend/CLAUDE.md + docs/map.md）立结构。
- Makefile 统一命令入口，.gitignore 挡住运行时产物。

**Non-Goals:**
- 不做任何业务模型/业务表（资产表属于 example-module change）。
- 不做护栏实现、不做 deps.yaml 双向校验（属于 guardrails change）。
- 不做设计系统 / tokens / 应用壳（属于 design-system change）。
- 不做单实例锁 / 托盘 / 日志 / config / 对话框（属于 runtime change）。
- 不做自动更新（已从主线移除）。
- 不写 README 全量部署文档（属于 ci-release change，本 change 只留最小占位）。

## Decisions

### D1: 目录结构采用 Wails v2 原生布局，非自定义 backend/frontend 拆分
Wails v2 `wails init -t vue-ts` 生成的布局是 `main.go` + `app.go`（Go 侧）+ `frontend/`（Vue3 侧），Go 侧默认在根目录而非 `backend/` 子目录。

**决策**：采用 Wails 原生布局（`main.go`、`app.go`、`frontend/`），在其上叠加领域子目录（`internal/` 放 service/model 等）作为后续业务分层。理由：
- 与 Wails CLI 的 dev/build 工具链零摩擦，不逆框架。
- 经验文档第 5 节的三层纪律，用 `internal/` 内的包边界表达，而不是强改 Wails 布局。

**备选**：强行拆 `backend/` + `frontend/` 根目录——否决，需重写 wails.json 路径配置，徒增成本无收益。

### D2: SQLite 用 gorm + 纯 Go 驱动（glebarez/sqlite 或 mattn/go-sqlite3）
**决策**：gorm + `github.com/glebarez/sqlite`（纯 Go，无 CGO，跨平台交叉编译零痛苦）。理由：Wails 要跨平台交叉编译，`mattn/go-sqlite3` 依赖 CGO，交叉编译到 Windows/macOS 很痛；glebarez 是纯 Go 实现，`wails build` 交叉编译顺滑。

**备选**：`mattn/go-sqlite3`（CGO）——否决，交叉编译成本高；`modernc.org/sqlite`——可用，但 glebarez 的 gorm driver 集成更常见、文档多。

### D3: AutoMigrate 采用"显式注册表"而非"自动扫描"
**决策**：立一个 `internal/model/` 包，里面一个 `AllModels()` 切片显式列出待迁移模型，`migrate.go` 遍历它调用 `db.AutoMigrate(...)`。本 change 该切片为空（或仅占位注释），后续 example-module 往里面加。

理由：这是经验文档"单一真相 + 双向校验"的第一处落地——模型注册只此一处，护栏 change 会双向断言"每个带 BaseModel 的结构体都注册了"。显式注册表是护栏可验证的前提，自动扫描则无法验证。

### D4: Makefile 是唯一命令真相，命令名对齐经验文档
`make dev / build / test / lint / smoke / package / gen`。其中 `test/lint/smoke/gen` 本 change 先给占位 target（打印"待后续 change 实现"），避免 AI/人误以为能力已存在。理由：经验文档第 7 节"统一命令入口"，且空 target 要显式报"未实现"而非静默通过（呼应"护栏要感知自己瞎了"的精神）。

### D5: 宪法文档分层 + ⚙️ 标记约定
根 `AGENTS.md` 管跨端铁律（目录命名、依赖登记、干净性、验证入口）；根 `CLAUDE.md` 管 Go 规范（复用 `docs/代码规范.md` 的 Go 部分）；`frontend/CLAUDE.md` 管 Vue 规范（复用 Vue 部分）；`docs/map.md` 是代码导航地图（"哪类代码在哪，先读它"）。

> 注意：Wails v2 的 Go 代码在根目录（`main.go`/`app.go`/`internal/`），
> 没有 `backend/` 子目录。因此 Go 域规范落在根 `CLAUDE.md`，而非 `backend/CLAUDE.md`。
> 经验文档里"backend/CLAUDE.md"是 base/ 的形态，迁移到 Wails 时映射为根 `CLAUDE.md`。

**决策**：能被机器强制的规则标 `⚙️` 并指向对应检查（如"禁硬编码品牌名 → 由 guardrails change 的护栏强制"）。本 change 里 ⚙️ 标记指向"未来护栏"，先立约定，护栏后补。

## Risks / Trade-offs

- **[Risk] Wails v2 版本/模板随 CLI 升级漂移** → 在 `AGENTS.md` 记录生成时锁定的 wails 版本号，`go.mod`/`package.json` 作为真相；升级 wails 需过 OpenSpec change，不随手升。
- **[Risk] 空 AutoMigrate 注册表让"迁移框架"形同虚设** → 本 change 的验收标准明确要求"加一个模型、迁移即生效"（用临时示例验证后删掉，或留一个无副作用的 `_example` 占位模型）。
- **[Risk] Makefile 空 target 静默成功，误导 AI 以为能力已存在** → 空 target 一律 `echo "未实现，待 X change" && exit 1`，显式失败。
- **[Risk] 宪法文档流于空壳、与代码脱节** → 宪法与骨架同 change 交付，每条铁律要么此刻可执行、要么标注"待 X change 强制执行"，不留无法验证的软约束。
