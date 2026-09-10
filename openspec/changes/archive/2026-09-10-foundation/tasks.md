# foundation — 实施任务

## 1. 初始化 Wails 骨架

- [x] 1.1 用 `wails init -n <name> -t vue-ts` 生成最小骨架（Go + Vue3），并在 `AGENTS.md` 记录锁定的 wails 版本号
- [x] 1.2 清理生成器的示例代码/占位文案，得到无业务残留的中性骨架
- [x] 1.3 确认 `make dev` 能启动、主窗口正常显示、Vue3 热更新生效

## 2. SQLite + gorm 迁移框架

- [x] 2.1 引入 gorm 与纯 Go SQLite 驱动（glebarez/sqlite），写入 `go.mod`
- [x] 2.2 建 `internal/model/` 包，提供 `AllModels()` 显式注册切片（本 change 为空或仅占位）
- [x] 2.3 建 `migrate.go`，启动时建立 gorm 连接并遍历 `AllModels()` 调用 AutoMigrate
- [x] 2.4 用临时示例模型验证"加模型→建表"生效，验证后移除（或保留无副作用占位）

## 3. 宪法文档体系

- [x] 3.1 写根 `AGENTS.md`：跨端铁律（目录命名、依赖登记、干净性、验证入口），硬约束标 ⚙️
- [x] 3.2 写 `frontend/CLAUDE.md`：Vue 规范（复用 `docs/代码规范.md` 的 Vue 部分）
- [x] 3.3 写根 `CLAUDE.md`：Go 规范（Wails 无 backend/ 目录，Go 域规范落在根目录）
- [x] 3.4 写 `docs/map.md`：代码导航地图，说明"哪类代码在哪，先读它"

## 4. 统一命令入口与干净性

- [x] 4.1 写 `Makefile`：`dev / build / test / lint / smoke / package / gen` target，`help` 列出用途
- [x] 4.2 未实现 target（lint/smoke/gen）打印"未实现，待 X change"并 `exit 1`（test 已实现，映射 go test）
- [x] 4.3 配 `.gitignore`：挡 `*.db`、`node_modules/`、`dist/`、`build/`、`*.exe`、日志等产物
- [x] 4.4 验证 `make package` 能跨平台打包（当前平台），构建无致命错误

## 5. 验证与收尾

- [x] 5.1 跑通 `make dev` 与 `make package`，确认真实可跑（非"嘴上说"）
- [x] 5.2 `openspec validate foundation` 通过
- [x] 5.3 检查运行时产物被 `.gitignore` 覆盖（项目尚未 git init，无 git status 可查；已核对 build/bin、frontend/dist、node_modules、*.db 均被忽略）
