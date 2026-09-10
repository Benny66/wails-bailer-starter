# wails-bailer-starter

基于 **Wails v2 (Go + Vue3)** 的桌面客户端脚手架基座，开箱即用、可反复复用。

> **这是脚手架母版，不是可直接运行的项目。** 用它开发新项目，先执行实例化脚本生成干净项目。

- 后端：Go + gorm + SQLite（纯 Go 驱动，跨平台交叉编译无 CGO 依赖）
- 前端：Vue3 + Vite + Element Plus + Vue Router + Pinia
- 设计系统：暗色优先的专业工具风，单主色派生全色阶
- 运行时：单实例锁、系统托盘（Windows/Linux）、日志轮转、配置读写、原生对话框、崩溃落盘
- 工程化：架构护栏（AST/ESLint）、依赖登记制、`make gen` 模块生成器、OpenSpec 治理

## 实例化新项目

```bash
bash scripts/init.sh myapp    # 生成干净新项目（替换模块名 + 清空归档历史 + git init）
cd ../myapp
make dev                      # 启动开发态
```

`init.sh` 会把母版的占位符 `__APP_NAME__` 全局替换为你的项目名（go.mod / import / wails.json / index.html / 生成器脚本），并清空母版的归档历史、保留能力基线。

## 环境准备

| 依赖 | 版本 | 说明 |
|---|---|---|
| Go | ≥ 1.25 | `go version` 确认 |
| Node | ≥ 20 | `node --version` 确认 |
| Wails CLI | v2.15+ | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| 平台依赖 | — | Windows 需 WebView2；Linux 需 `libgtk-3-dev libwebkit2gtk-4.1-dev` |

## 快速开始

```bash
make dev       # 启动开发态（含前端热更新）
make build     # 编译当前平台产物
make package   # 打包安装包
```

## 命令表

| 命令 | 作用 |
|---|---|
| `make dev` | 启动开发态（wails dev，前端热更新） |
| `make build` | 编译当前平台产物（快速，.app 不封装 dmg） |
| `make package` | 打包真安装包（`make package os=windows\|macos\|linux`） |
| `make test` | 运行 Go 测试（含架构护栏） |
| `make lint` | 静态检查（护栏 + go vet + ESLint） |
| `make smoke` | 冒烟测试（构建 → 启动 → 断言 → 清理） |
| `make gen name=<module>` | 生成新模块（锚点注入 + 幂等） |

## 目录结构

见 [`docs/map.md`](docs/map.md) —— 代码导航地图，AI 干活前先读它。

核心分层：

```
app.go                     # 绑定方法（前端可调用的 Go 方法）
internal/model/            # 数据模型 + AllModels() 注册表（唯一真相）
internal/service/          # 业务服务层（CRUD 逻辑）
internal/database/         # gorm + SQLite 连接、迁移
internal/guard/            # 架构护栏（AST 断言，随 go test 跑）
internal/{config,logging,dialog,tray,crash}/  # 运行时服务
frontend/src/styles/       # 设计令牌（tokens.css / theme.css）
frontend/src/layouts/      # 应用壳（侧栏 + 标题栏）
frontend/src/views/        # 页面
_example/                  # 黄金范例模板（make gen 的母版）
scripts/gen.sh             # 模块生成器
```

## 生成新模块

```bash
make gen name=asset
```

从 `_example/` 模板生成四段：model + service + 绑定方法 + 前端页面，并自动注入到 `AllModels()`、`app.go`、路由、菜单。生成后填 `// TODO` 处业务逻辑，再 `make dev` 重新生成 bindings。

## 换肤（设计系统）

专业工具风，暗色优先。换肤只需改 `frontend/src/styles/tokens.css` 里**品牌区**的主色，其余颜色全部由主色自动派生，禁硬编码 hex 色值（ESLint 护栏强制）。

logo 在 `build/appicon.png`（前端资源与托盘图标共用）。

> 品牌名/模块名已占位符化为 `__APP_NAME__`，通过 `scripts/init.sh` 实例化时全局替换，无需手改。

## 架构护栏

本基座把"约定"编译成"会红的检查"（详见 `AGENTS.md`）：

- **分层越界**：`app.go` 不直接 import gorm/database（经 service）；model 是叶子。
- **模型注册双向校验**：带 `BaseModel` 的结构体必须登记进 `AllModels()`。
- **依赖登记制**：新增依赖必须登记 `deps.yaml`（双向校验）。
- **前端 import 安全**：渲染层禁 import node 能力；禁硬编码色值。

护栏"感知自己瞎了"：解析到 0 个结果会 Fatal，而非静默放行。

## Windows WebView2 部署

Windows 端依赖 WebView2 运行时。大多数 Win10/11 已预装；若未安装，Wails 会在启动时引导安装（在线）。离线环境见 [Microsoft WebView2 部署文档](https://learn.microsoft.com/microsoft-edge/webview2/concepts/distribution)。

## 打包与发布

```bash
make package              # 打包当前平台真安装包（mac 产 .dmg，win 产 .exe）
make package os=windows   # 交叉编译 Windows .exe（mac/linux 上可用）
make package os=macos     # 仅 mac 上可用（Wails 不支持交叉编译 mac）
make package os=linux     # 仅 linux 上可用（Wails 不支持交叉编译 linux）
```

- **产物位置**：`build/bin/`。
- **mac**：`.app` 会封装为 `.dmg`（hdiutil，系统自带）。Windows：有 `makensis`（NSIS）时产 `.exe` 安装器，否则降级为裸 exe。
- **母版防呆**：含 `__APP_NAME__` 占位符时拒绝打包，需先 `scripts/init.sh` 实例化。
- CI：`.github/workflows/build.yml` 在 push/PR 时自动跑静态检查 + 三平台编译，产物上传为 artifact。
- 版本号管理：`scripts/release.sh`（可选，版本号 + 打包 + release 说明）。

## 已知限制

- **macOS 无系统托盘**：Wails v2 的 NSApplication delegate 与所有 systray 库冲突（详见 `openspec/changes/runtime/design.md` D1），故 macOS 关闭即退出，托盘能力仅在 Windows/Linux 提供。
- **无 headless 模式**：冒烟测试定位为本地验证，CI 只编译不启动 GUI。
