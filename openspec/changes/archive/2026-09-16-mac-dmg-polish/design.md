# mac-dmg-polish — 技术设计

## Context

`scripts/package.sh` 的 mac 分支（`install-experience` 归档产物）已实现
「staging → UDRW → osascript 布局 → UDZO」链路。本 change 不动主链路，只解决：
**挂载点错误导致美化静默失效**（实测新发现）、背景图印不上应用名、美化结果不可验证。

当前相关代码事实：

```
scripts/package.sh:89    MOUNT_PT="$(mktemp -d)"    ← ⚠️ 实测：Finder 不认此挂载点
scripts/package.sh:121   set bounds of container window to {100, 100, 700, 500}   → 600×400
scripts/package.sh:126   set position of item "${APP_NAME}.app" to {150, 200}
scripts/package.sh:127   set position of item "Applications"      to {450, 200}
build/darwin/dmg-background.png                                                   → 600×400，箭头水平居中
```

### 真机实测结论（2026-09-14 · macOS 26.4 · Finder 运行中）

- **发现 A（阻断级）**：`tell disk "<vol>"` 要求卷挂载在 `/Volumes/<name>`。
  用 `mktemp -d`（`/var/folders/…`）当挂载点，Finder 报 `-1728 不能获得 disk`。
  ⇒ **当前美化 100% 静默失败**，被 `|| echo "提示…"` 吞掉。这是归档 tasks 4.2 假勾选的根因。
- **发现 B**：`position of item` = 图标**中心**（`.DS_Store` 的 `Iloc` 记录精确等于
  `(150,200)`/`(450,200)`）。故 `(150+450)/2 = 300 == 600/2`，**坐标本就对称**。
  原「偏右 50px」假设（把 bounds 当同一坐标系）**不成立**，已删。
- **发现 C**：窗口 `bounds` 宽 ≠ content 视口宽（后者要减侧栏 + splitter）。
  图标按 content 宽定位，背景图却按完整 600 宽绘制 → 箭头错位随侧栏浮动：
  侧栏 0→偏 +0.5，侧栏 175→偏 +88。Finder **不暴露** content 视口宽度属性
  （`properties of container window` dump 已确认），无法读取后校正。

## Goals / Non-Goals

**Goals:**
- **修复挂载点 bug**：dmg 挂载到 `/Volumes/<name>`，让 Finder 能解析卷、美化真正生效（发现 A）。
- 背景图箭头与图标**同一坐标系绝对像素对齐**（发现 C 的解法，见 D1）。
- 背景图可选地印上当前应用名（构建期合成，模板缺失/合成失败时降级）。
- dmg 产出后**回读** `.DS_Store`，产出携带「真美化 / 降级」标记，退出码可区分。
- 全流程零新增外部依赖；失败一律降级为「可用但无美化」，绝不产出坏 dmg。

**Non-Goals:**
- 不做 dmg 签名 / 公证（`install-experience` 已划入 Non-Goals，本 change 不拉回）。
- 不改 win/linux 分支。
- 不替用户改 Finder 偏好（如侧栏显示已挂载卷）——见 Q2。

## Decisions

### D0: 挂载点改为 /Volumes/<name>（实测驱动的阻断级修复）

```bash
# 旧（静默失败）
MOUNT_PT="$(mktemp -d)"
hdiutil attach "$TMP_DMG" -mountpoint "$MOUNT_PT" -nobrowse -quiet

# 新（Finder 可解析）
MOUNT_PT="/Volumes/$APP_NAME"          # 与 -volname 同名
hdiutil attach "$TMP_DMG" -mountpoint "$MOUNT_PT" -nobrowse -quiet
```

**边界**：`/Volumes/<APP_NAME>` 可能与既有卷重名（Application Name 撞车）。
策略：先探测，若已存在则用 `/Volumes/${APP_NAME}-$$` 后缀**并把 `tell disk` 同步改用该名**
（卷名由 `-volname` 决定；探测器守 `hdiutil info`）。
这个「同名后改名」逻辑必须在改脚本时同步改 `-volname`，否则 `tell disk` 又对不上卷名。

### D1: 坐标参数化 + 背景图箭头对齐图标锚点（修正原「对称中心」思路）

实测既证明坐标本就对称（发现 B），也证明「以背景图/窗口中心对齐」思路不成立（发现 C）。
**改用绝对像素对齐**：箭头与图标都在 **container 坐标系**，只要箭头画在
两图标中心点之间，就与侧栏宽度无关。

```
WINDOW_X=100  WINDOW_Y=100  WINDOW_W=600  WINDOW_H=400
ICON_LEFT_X=150   ICON_RIGHT_X=450   ICON_Y=200   ICON_SIZE=100
                                          ↑ container 坐标，pos=中心
背景图箭头: 画在 ((ICON_LEFT_X+ICON_RIGHT_X)/2, ICON_Y) = (300, 200)
断言: (ICON_LEFT_X + ICON_RIGHT_X) / 2 处即箭头锚点，与内容视口无关
```

**关键差异**：不再断言「中点 == 窗口宽/2」（那会随侧栏漂移而假失败），
而是**用同一组常量同时喂给 osascript 的图标落点与背景图合成脚本的箭头坐标**——
单一来源，二者永远一致。

### D2: 背景图合成用 macOS 原生工具，零外部依赖

三条候选：

| 方案 | 优点 | 缺点 | 结论 |
|---|---|---|---|
| `sips` 叠加文字 | 系统自带 | `sips` 无绘制文字能力（仅变换/格式） | ✗ 不可行 |
| `qlmanage` 渲染 HTML/SVG | 系统自带，支持富文本 | 输出质量/尺寸控制别扭，依赖 HTML 转义 | △ 备选 |
| `swift` 调 CoreGraphics | 系统自带（装了 Xcode CLT 即无），完全可控 | 首次编译慢（几秒），依赖 CLT | ✓ **首选** |

**决策**：首选 `swift` 内联脚本（`swift -` 读 stdin 执行，不改仓库结构）绘制
「底色 + 居中应用名 + 箭头@`(300,200)`」；`swift` 不可用 → 降级为 `qlmanage` 渲染模板 SVG；
再不可用 → 直接用模板/占位图，跳过文字。

箭头坐标由 D1 的 `ICON_*` 常量推导并**作为参数传给合成脚本**（单一来源）。
文字内容取 **`wails.json` 的 `outputfilename`**（已定 Q4）；为空时回退到无文字的中性占位图。
**理由**：与 `install-experience` 的「零依赖」原则一致（拒绝 create-dmg 是同一逻辑）。
`swift` 是系统随附编译器，不计入 `deps.yaml`（非 Go module / npm 包）。

### D3: 回读校验 —— 挂载成品 dmg，检查 .DS_Store

产出 dmg 后，**二次挂载**成品（只读）：

```
hdiutil attach <final>.dmg -readonly -nobrowse -mountpoint <verify_pt>
test -f <verify_pt>/.DS_Store          # 存在性
grep -a "bg.png" <verify_pt>/.DS_Store # 背景图引用（.DS_Store 是二进制，用 -a 文本搜）
hdiutil detach <verify_pt>
```

- 两者皆真 → `✓ 美化已生效`，退出码 0。
- 缺失 → `⚠ 降级产物（无美化）：本机无 GUI 会话或 Finder 不可用`，退出码 0（**不阻断**，
  与 `install-experience` 的「美化失败不阻断」一致），但**打印显著区分**。
- `STRICT_POLISH=1`（**已定**）→ 降级时退出码非 0，供 CI 在需要时把「无美化」当构建失败。

**权衡**：`.DS_Store` 是二进制格式、无稳定公开规范，`grep` 到 `bg.png` 字面量属于
「启发式」。失败方向安全（查不到 = 报降级，不会把降级误报成成功），可接受。

**实测佐证**：`.DS_Store` 内确实含 `/.background/bg.png` 字面量（strings 已验证），
且卷根出现 10244 字节 `.DS_Store` → 该判据成立。

### D4: 移除死代码 —— `[ -f "$BG_SRC" ]` 守卫包住整个 osascript

`package.sh:113-134` 的现状是：**整个** Finder 布局 osascript 被 `[ -f "$BG_SRC" ]` 守卫。
含义是「没背景图就完全不做布局（连图标定位都没有）」。但设计意图是「缺图只跳过背景」。
两者矛盾。本 change 把守卫收窄到**仅** `set background picture` 一行，
图标定位/窗口尺寸**始终执行**，使「无背景图」仍得到「有布局、无背景」的降级态。

## Risks / Trade-offs

- **[Risk] `/Volumes/<APP_NAME>` 与既有卷重名** → D0 的重名探测 + 后缀回退（同步改 `-volname`）。
- **[Risk] `swift` 首次编译拖慢打包** → 打包非高频操作，可接受；且失败自动降级。
- **[Risk] 二次挂载成品 dmg 在 CI 上失败（无 GUI）** → 只读挂载不依赖 Finder，应可用；
  若失败则报「无法校验」而非误报「无美化」。
- **[Risk] 回读校验通过 ≠ 视觉一定对齐** → 校验只证明「布局已写入」，不证明「好看」。
  最终视觉仍需 `make package` + 人工打开确认（tasks 5.1）。
- **[Risk] `.DS_Store` 格式随 macOS 版本变化致回读失效** → 失败方向安全（报降级），
  并附「写法可能已变更」提示（对齐 AGENTS.md 铁律 5「护栏必须感知自己瞎了」）。

## Resolved Decisions（原 Open Questions）

- **Q1（已定 · 实测推翻原假设）**：`position` = 图标**中心**；坐标本就对称；
  真正的错位来自内容视口 ≠ 窗口宽（发现 C）。解法见 D1（绝对像素对齐）。
  **原 design 的「偏右 50px」论断是错的，已删除。**
- **Q2（已定 · 超范围）**：不处理 Finder 侧栏找回挂载卷的问题。列为 **Known Limitation**：
  用户误关 dmg 窗口后，再次双击 `.dmg` 即可重新打开挂载卷（写入 README）。
- **Q3（已定）**：采用 `STRICT_POLISH` 开关。默认降级不阻断（exit 0），
  `STRICT_POLISH=1` 时降级 → exit 1，供 CI 按需把「无美化」当构建失败。
- **Q4（已定）**：背景图文字用 **`wails.json` 的 `outputfilename`**（脚本已有该变量，
  零额外解析）。若该名为空则回退到中性占位图。


