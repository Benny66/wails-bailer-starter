# docs/map.md — 代码导航地图

> 目的：告诉 AI「哪类代码在哪个文件」，先读这里再动手，避免 `find` 全库盲扫。

## 顶层

| 路径 | 内容 |
|---|---|
| `main.go` | 应用入口，wails 启动配置（窗口/绑定/assets） |
| `app.go` | App 结构体与绑定方法（前端可直接调用的 Go 方法都在这） |
| `AGENTS.md` | 跨端通用铁律 |
| `CLAUDE.md` | Go 域规范 |
| `frontend/CLAUDE.md` | Vue 域规范 |
| `Makefile` | 统一命令入口 |
| `wails.json` | wails 打包/构建配置 |

## Go 侧（`internal/`）

| 路径 | 内容 |
|---|---|
| `internal/model/` | 数据模型 + `AllModels()` 注册表（模型唯一真相） |
| `internal/database/` | gorm + SQLite 连接、迁移执行 |

> 后续扩展（runtime / service / config 等）落地后在此补充。

## 前端（`frontend/src/`）

| 路径 | 内容 |
|---|---|
| `main.ts` | 前端入口 |
| `App.vue` | 根组件 |
| `components/` | 通用组件 |
| `stores/` | Pinia 状态 |
| `views/` | 页面（后续 example-module 落地） |

## 自动生成（勿手改）

| 路径 | 内容 |
|---|---|
| `frontend/wailsjs/` | wails 自动生成的 Go↔TS 绑定，由 `make dev` 重新生成 |
| `frontend/dist/` | 前端构建产物 |
