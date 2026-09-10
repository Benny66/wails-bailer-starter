# project-scaffold Specification

## Purpose
TBD - created by archiving change foundation. Update Purpose after archive.
## Requirements
### Requirement: 最小可跑骨架
脚手架 MUST 提供一个 Wails v2 (Go + Vue3) 骨架，`make dev` 能启动开发态，`make package` 能产出跨平台安装包。

#### Scenario: 开发态启动
- **WHEN** 用户在项目根目录执行 `make dev`
- **THEN** Wails 开发服务启动，Vue3 前端热更新生效，主窗口正常显示

#### Scenario: 打包
- **WHEN** 用户执行 `make package`
- **THEN** 系统产出目标平台安装包（当前平台），且构建过程无致命错误

### Requirement: 统一命令入口
所有高频操作 MUST 通过 `Makefile` 的 target 暴露，AI 与人不记零散脚本路径。

#### Scenario: 命令入口存在
- **WHEN** 用户执行 `make help` 或查看 `Makefile`
- **THEN** 系统列出 `dev / build / test / lint / smoke / package / gen` 等 target 及其用途

#### Scenario: 未实现能力显式失败
- **WHEN** 用户执行某个尚未实现的 target（如 `make smoke`）
- **THEN** 系统打印"未实现，待 X change"并返回非零退出码，而非静默成功

### Requirement: 干净性
`.gitignore` MUST 挡住所有运行时产物，禁止其进入版本控制。

#### Scenario: 产物不入库
- **WHEN** 项目产生 `*.db`、`node_modules/`、`dist/`、`build/`、`*.exe`、日志文件等
- **THEN** 这些文件被 `.gitignore` 规则排除，`git status` 不显示它们

