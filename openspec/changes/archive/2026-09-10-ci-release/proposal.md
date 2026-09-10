# ci-release — 冒烟测试、CI 打包与交付文档

## Why

基座要能"交付给下游用"，就必须有：客观的"能跑"证据（冒烟）、提交即编译的 CI、以及一份让人能上手的 README。否则"完成了"只是嘴上说，换台机器就崩。

## What Changes

- **冒烟测试**：本地 `make smoke`——`wails build` 后启动 app，经 bindings 发一次握手（调一个 Go 方法），断言返回，`trap cleanup` 保证不残留进程。
- **CI（GitHub Actions）**：提交自动跑 `make test` + `make lint`（不启动 GUI），再 `wails build` 编译三平台产物（windows/mac/linux），上传 artifact。
- **README**：环境准备、开发、打包、目录结构、`make` 命令、设计系统换肤说明、WebView2 部署说明（Windows）。
- **版本号管理脚本**：`scripts/release.sh`（版本号 + 打包 + release 说明，可选）。

## Capabilities

### New Capabilities

- `smoke-test`: `make smoke` 冒烟——构建后启动 + 绑定握手 + 断言 + 清理。
- `ci-pipeline`: GitHub Actions 工作流——静态检查 + 三平台编译 + artifact 上传。
- `delivery-docs`: README 交付文档（环境/开发/打包/换肤/部署）。

### Modified Capabilities

- `project-scaffold`: `make smoke` 从"未实现占位"改为"真实冒烟"。

## Impact

- **新增代码**：`scripts/smoke.sh`、`.github/workflows/build.yml`、`README.md`、`scripts/release.sh`。
- **依赖**：无新运行时依赖（CI 用 GitHub Actions 自带环境）。
- **风险**：Wails 无 headless 模式，冒烟在 CI 无显示器环境受限——故 CI 只跑构建/单测/护栏，冒烟定位为本地验证。
- **破坏性**：`make smoke` 行为变化。
