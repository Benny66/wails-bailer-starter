# package-release — 参数化打包 + mac dmg 安装包 + 母版防呆

## Why

用户执行 `make package` 后"拿不到安装包"，根因有三：① Wails v2 在 macOS 只产 `.app` 目录、不产 `.dmg` 安装包；② `make package` 与 `make build` 语义几乎相同（只差 `-clean`），命名误导；③ 母版占位符化后直接 `make package` 会产出 `__APP_NAME__.app` 这种怪名字。需要把打包做成"真安装包 + 参数化平台 + 母版防呆"。

## What Changes

- **`make package [OS]` 参数化**：无参=当前平台；`make package windows`=交叉编译 Windows .exe（mac/linux 上可用）；`make package macos`/`make package linux` 仅对应平台可用，否则明确报错。
- **macOS 真安装包 `.dmg`**：`.app` 后用 `hdiutil create` 封装成 `.dmg`（macOS 自带，零依赖）。
- **Windows 真安装包**：`-nsis` 生成 .exe 安装器。
- **母版防呆**：打包前检测 `__APP_NAME__` 占位符残留，若存在则报错"请先 `scripts/init.sh` 实例化"，阻止产出怪名字产物。
- **`make build` 保留快速编译**（.app 不封装 dmg），`make package` 升级为"真安装包"。

## Capabilities

### New Capabilities

- `package-release`: 参数化打包（当前平台 + windows 交叉编译）、mac `.dmg` 安装包、windows NSIS 安装器、母版占位符防呆。

### Modified Capabilities

- `project-scaffold`: `make package` 语义从"清理重编译"变为"真安装包 + 参数化 + 防呆"。

## Impact

- **修改代码**：`Makefile`（package target 参数化 + 防呆 + dmg/nsis）、`scripts/`（新增 `package.sh` 或内联 Makefile 逻辑）。
- **新增依赖**：无（`hdiutil` 是 macOS 系统自带，NSIS 由 wails build 内置）。
- **破坏性**：`make package` 行为变化（从"只产 .app"到"产 .dmg/.exe 安装包"）。README 打包章节需同步。
