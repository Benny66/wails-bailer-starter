# install-experience — 安装体验（mac dmg 拖拽 + win NSIS 安装器）

## Why

打包产出了 dmg，但双击后只是挂载出一个临时卷、里面的 `.app` 原地运行，用户"没有装到应用程序里"。根因是 mac 与 win 的"安装"模型不同：macOS 的 dmg 分发靠"用户把 .app 拖到 /Applications"，而当前 dmg 里既没有 `/Applications` 软链、也没有拖拽引导（背景图/箭头），用户不知道怎么装。Windows 侧虽然有 NSIS 安装器能力，但从未明确安装范围与体验。本 change 补齐两端"双击即装"的完整体验。

## What Changes

- **macOS dmg 拖拽安装**：dmg 卷内含 `MyApp.app` + 指向 `/Applications` 的软链，用户拖拽即装。
- **macOS dmg 美化**：Finder 窗口布局（app 在左、Applications 在右）、箭头引导 + 背景图 + 窗口尺寸定位。用 `hdiutil + osascript`（零依赖）。
- **Windows NSIS 安装范围参数化**：`make package os=windows` 支持 `-installscope user|machine`（默认 `user`，装 `%LOCALAPPDATA%\Programs`，免管理员）。
- **makensis 缺失提示强化**：无 NSIS 编译器时明确提示安装方式，而非静默降级。
- **文档**：README 讲清两端"安装"模型差异与各自的操作。

## Capabilities

### New Capabilities

- `install-experience`: mac dmg 拖拽安装（Applications 软链 + 美化引导）、win NSIS 安装范围参数化。

### Modified Capabilities

- `package-release`: mac dmg 从"裸 .app 封装"升级为"拖拽安装体验（软链 + 美化）"；win 分支增加安装范围参数。

## Impact

- **修改代码**：`scripts/package.sh`（mac dmg 美化 + win scope 参数）、`Makefile`（透传 scope）。
- **新增资源**：dmg 背景图（`build/darwin/dmg-background.png`，可用 sips 从现有图缩放生成）。
- **依赖**：无新增（hdiutil/osascript/sips 均 macOS 自带）。
- **破坏性**：无（package.sh 行为增强，默认值合理）。
