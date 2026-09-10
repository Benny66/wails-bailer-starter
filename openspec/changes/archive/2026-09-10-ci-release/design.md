# ci-release — 技术设计

## Context

地基有 `make test`（Go 测试）、`make build`/`make package`（wails 打包）。缺：冒烟（客观"能跑"证据）、CI（提交即编译）、README（交付文档）。

Wails 无 headless 模式，冒烟在 CI 无显示器环境受限。

## Goals / Non-Goals

**Goals:**
- `make smoke` 本地冒烟：构建 → 启动 → 绑定握手 → 断言 → 清理。
- CI：静态检查 + 三平台编译 + artifact 上传。
- README 交付文档。

**Non-Goals:**
- CI 里不启动 GUI（无显示器）。
- 不做代码签名/公证（v1.1）。
- 不做 release 自动化发布（release.sh 可选，不进 CI 主流程）。

## Decisions

### D1: 冒烟定位为"本地验证"，不进 CI
Wails 无 headless，CI 无显示器启动会失败。冒烟 = 本地 `make smoke`：`wails build` 后启动 app，经 bindings 发握手，`trap cleanup EXIT` 保证不残留。CI 只跑 `make test` + `make lint` + 三平台 `wails build`。

### D2: CI 用 GitHub Actions，三平台矩阵
`matrix.os: [ubuntu-latest, macos-latest, windows-latest]`，各自 `wails build` 编译原生产物，`actions/upload-artifact` 上传 exe/app/appimage。不启动应用。

### D3: 冒烟脚本用 shell + trap
`scripts/smoke.sh`：`wails build` → 后台启动二进制 → sleep 等就绪 → 经绑定调 Go 方法（或检查进程存活 + 日志）→ 断言 → `trap 'kill' EXIT`。防进程残留。

### D4: README 覆盖"上手 + 换肤 + 部署"
环境准备（Go/Node/wails 安装）、`make` 命令表、目录结构（指 docs/map.md）、换肤说明（改 name/logo/primary 三值）、Windows WebView2 部署说明。

### D5: 版本号管理可选，不进主流程
`scripts/release.sh`（版本号 + 打包 + release 说明）作为可选脚本，README 提及，但 CI 不依赖它。

## Risks / Trade-offs

- **[Risk] 冒烟启动 GUI 在 CI runner 上无法跑** → 已决策：冒烟本地、CI 只编译。
- **[Risk] 三平台编译在 CI 首次跑暴露交叉编译问题（CGO/图标）** → 已在 foundation 用纯 Go SQLite 驱动规避 CGO；图标三平台已在模板内置。
- **[Risk] 冒烟断言太弱（只看进程存活）** → 优先做"绑定握手"（调一个 Go 方法验证往返），进程存活只是兜底。
- **[Risk] CI 耗时过长** → 静态检查与编译分 job，编译矩阵可并行；缓存 go mod / npm。
