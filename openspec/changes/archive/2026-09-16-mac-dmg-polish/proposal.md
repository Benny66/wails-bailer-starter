# mac-dmg-polish — 让 dmg 美化「可验证 + 对齐 + 带应用名」

## Why

`install-experience` 已把 dmg 美化链路写进 `scripts/package.sh`，但真机实测
（2026-09-14 · macOS 26.4）暴露了三个问题，其中第一个是阻断级：

1. **美化根本没生效（阻断级）**。脚本用 `mktemp -d`（`/var/folders/…`）当挂载点，
   Finder 无法解析该卷——`tell disk "<vol>"` 报 `-1728 不能获得 disk`，
   整个 osascript 静默失败，被 `|| echo "提示…"` 吞掉。
   **当前 dmg 始终是「无美化」的**，归档 tasks 4.2 的 `[x]` 是假勾选。
   实测改用 `/Volumes/<name>` 后，布局立刻写入成功（卷根出现 `.DS_Store`）。
2. **美化结果不可验证**。产物不携带任何标记，用户/CI 无法区分「真美化了」还是「降级了」。
   这正是问题 1 能潜伏至今的原因——没有任何机制能发现它。
3. **背景图印不上应用名**。`init.sh` 只替换文本占位符 `__APP_NAME__`，改不了 PNG 像素。
   下游实例化后拿到的是与自己的 app 名无关的通用背景图。

> 原 proposal 曾称「图标偏右 50px」。**实测推翻该论断**：`position of item` 是图标**中心**
> （`.DS_Store` 的 `Iloc` 记录精确等于写入值），`(150+450)/2 = 300 == 600/2`，坐标本就对称。
> 真正会影响视觉的是**另一个**问题：窗口 `bounds` 宽 ≠ content 视口宽（后者要减侧栏），
> 背景图按完整宽绘制而图标按视口定位，导致箭头错位且随侧栏浮动（找不到可读的视口宽属性）。
> 解法见 design D1：箭头与图标都用同一组 container 绝对坐标，与视口无关。

## What Changes

- **修复挂载点 bug（阻断级）**：dmg 挂载到 `/Volumes/<name>`（带重名回退），
  让 Finder 能解析卷、美化真正生效。
- **箭头与图标同源对齐**：`WINDOW_*` / `ICON_*` 提为单一来源常量，背景图箭头的绘制坐标
  由 `ICON_*` 推导并**传给合成脚本**，二者永远一致，不受侧栏/视口影响。
- **新增回读校验**（核心）：dmg 产出后挂载成品，读 `.DS_Store` 是否存在且含背景图引用，
  打印 `✓ 美化已生效` / `⚠ 降级产物（无美化）`，并给出明确的退出码语义。
- **背景图构建期合成**：新增 `build/darwin/dmg-background.tpl.png`（模板）→ 打包时用
  macOS 原生工具合成 `outputfilename` 应用名 → `dmg-background.png`。
  **零新增依赖**（只用 `sips` / `qlmanage` / `swift` 之一，见 design D2）。
  模板缺失或合成失败 → 降级为中性占位图，不阻断。
- **移除死代码**：`package.sh` 中 `[ -f "$BG_SRC" ]` 守卫包住**整个** osascript —— 含义是
  「没背景图就连图标定位都不做」，与「缺图只跳过背景」的设计意图矛盾，收窄守卫范围。

## Capabilities

### New Capabilities

- `dmg-polish`: dmg 背景图构建期合成（模板 + `outputfilename` → 成品 png，带中性降级）
  **+ 窗口/图标坐标归口单一来源，箭头锚点与图标同源**。

### Modified Capabilities

- `package-release`: dmg 产出后**回读校验美化结果**（真美化 vs 降级产物可区分），而非仅「产出 dmg」。
- `install-experience`: mac dmg 美化从「写了 Finder 布局（实际因挂载点错误从未生效）」升级为
  「挂载点修正 + 箭头与图标同源对齐 + 可回读验证」。

## Impact

- **修改代码**：`scripts/package.sh`（挂载点修正 + 坐标参数化 + 回读校验 + 调用背景合成）、
  `Makefile`（如需暴露 `BG_TEMPLATE` 覆盖点）。
- **新增资源/脚本**：`build/darwin/dmg-background.tpl.png`（模板）、背景合成脚本
  （`scripts/` 下，纯 macOS 原生工具，零外部依赖）。
- **依赖**：无新增（`sips`/`qlmanage`/`swift` 均 macOS 自带）。若走 `swift` 路线，
  仅是系统编译器调用，不入 `deps.yaml`（非 Go/npm 依赖）。
- **破坏性**：无。失败一律降级为「可用但无美化」的 dmg，行为向后兼容。
  唯一行为变化是：**过去静默失败的美化现在会真正生效**（这是修复，非破坏）。

