# install-experience — 技术设计

## Context

当前 `package.sh` 的 mac 分支只做 `hdiutil create -srcfolder <app>`，产出裸 dmg，无 Applications 软链、无美化。用户双击后无语义上的"安装"引导。

mac 与 win 的安装模型不同：
- macOS：`.app` 即完整应用，"安装" = 拖到 `/Applications`；dmg 负责引导。
- Windows：`.exe` 安装器（NSIS）执行向导，装到 `%LOCALAPPDATA%\Programs`（user）或 `Program Files`（machine）。

## Goals / Non-Goals

**Goals:**
- mac dmg：含 app + Applications 软链 + Finder 美化（背景图/箭头/图标定位/window 尺寸）。
- win NSIS：安装范围参数化（默认 user），makensis 缺失时明确提示。
- 文档讲清两端安装模型。

**Non-Goals:**
- 不做 .pkg（需开发者证书签名）。
- 不处理图标缓存问题（独立事项）。
- 不引入 create-dmg 等外部依赖（用 hdiutil + osascript 零依赖实现）。

## Decisions

### D1: mac dmg 美化用 hdiutil + osascript（零依赖）
标准流程（业界成熟做法）：
1. `mkdir staging`，放 `MyApp.app` + `ln -s /Applications staging/Applications` + `.background/bg.png`。
2. `hdiutil create -srcfolder staging -volname <name> -fs HFS+ -format UDRW tmp.dmg`（先做**可读写** dmg，才能改 Finder 布局）。
3. `hdiutil attach tmp.dmg`（挂载可读写卷）。
4. `osascript` 设置 Finder 视图：`.background` 背景图、窗口 bounds、图标位置（app 在左、Applications 在右）。
5. `hdiutil detach`。
6. `hdiutil convert tmp.dmg -format UDZO -o <name>.dmg`（压缩为只读最终产物）。

**关键点**：可读写中间态（UDRW）是必须的——Finder 布局信息（.DS_Store）只能在可写卷上写入。

**备选**：create-dmg——否决，引入外部依赖，违背脚手架零依赖原则。

### D2: 背景图资源与生成
背景图放 `build/darwin/dmg-background.png`。若项目无此图，用现有 `build/appicon.png` 或纯色占位。尺寸建议 600×400（Finder 窗口标准）。可后续让下游替换。

**决策**：母版提供一个**中性占位背景图**（纯色 + 暗示"拖到这里"），下游可替换。避免"没有背景图就崩"。

### D3: Applications 软链在 staging 阶段创建
`ln -s /Applications staging/Applications`——软链指向系统 /Applications，用户拖 app 进去即装。这是 mac dmg 的标准机制。

### D4: Windows 安装范围参数化，默认 user
`make package os=windows` 透传 `-installscope`，默认 `user`（`%LOCALAPPDATA%\Programs\<产品>`，免管理员）。可显式 `scope=machine` 切 `Program Files`。

**理由**：B 端工具分发给普通员工，免管理员最省事；machine 场景保留但非默认。

### D5: makensis 缺失时明确提示（不静默降级）
当前 package.sh 的 windows 分支无 makensis 时降级裸 exe。升级为：**明确打印安装指引**（`brew install makensis` 或 Windows 装 NSIS），让用户知道怎么获得真安装器。

## Risks / Trade-offs

- **[Risk] osascript 设置 Finder 布局在某些 macOS 版本上失败** → 加错误容忍：布局失败不阻断 dmg 生成（降级为"无美化的 dmg + 软链"仍可用）。
- **[Risk] hdiutil 中间态 dmg 挂载失败（残留挂载点）** → `trap` 清理 + detach 容错。
- **[Risk] 背景图缺失导致美化失败** → 母版内置中性占位图，缺图时跳过背景设置。
- **[Risk] 可读写 dmg 流程比直接 create 慢** → 可接受（打包不是高频操作）。
- **[Risk] Windows 安装范围改动影响 CI 打包** → CI 未跑 NSIS（makensis 未装），不受影响。
