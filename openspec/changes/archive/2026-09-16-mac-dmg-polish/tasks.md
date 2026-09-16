# mac-dmg-polish — 实施任务

## 0. 前置实测（未定稿不得动常量）

- [x] 0.1 真机 GUI 实测（2026-09-14，macOS 26.4 / Finder 运行中 / 真实 console 会话）。
      用自建 probe.app 走真实 UDRW→挂载→osascript→读回链路，结论：
      - **发现 A（阻断级）**：`tell disk` 要求卷挂载在 `/Volumes/<name>`。现脚本用
        `mktemp -d`（`/var/folders/…`）当挂载点，Finder 报 `-1728 不能获得 disk`——
        **当前美化 100% 静默失败**，被 `|| echo "提示…"` 吞掉。这是 tasks 4.2 假勾选的根因。
      - **发现 B**：`position of item` = 图标**中心**（`.DS_Store` 的 `Iloc` 记录精确等于
        `(150,200)`/`(450,200)`，排除左上角假设）。故 `(150+450)/2 = 300 == 600/2`——
        **坐标本就对称**。design 里「偏右 50px」的怀疑**不成立**，予以推翻。
      - **发现 C**：但窗口 `bounds` 宽 ≠ content 视口宽（后者要减侧栏 + splitter）。
        图标按 content 宽定位，背景图却按完整 600 宽绘制 → **跨机漂移**：侧栏越宽，
        箭头（固定 300）越偏离 icon 区中心（侧栏 0→偏差 +0.5，侧栏 175→偏差 +88）。
        Finder **不暴露** content 视口宽度属性（properties dump 已确认），无法读取后校正。
- [x] 0.2 据 0.1 结论重定方案 —— **已完成**（design 已更新：新增 D0 挂载点修复、
      D1 改为箭头同源锚点、删除「偏右 50px」错误论断；proposal/specs 同步）

## 1. 挂载点修复（阻断级，必须先做）

- [x] 1.1 `package.sh` mac 分支：挂载点由 `mktemp -d` 改为 `/Volumes/<VOLNAME>`
- [x] 1.2 重名探测：`/Volumes/<name>` 已存在时改用 `${APP_NAME}-$$` 后缀，并**同步**改
      `-volname` 与 `tell disk` 引用名（二者始终一致）
- [x] 1.3 traps 清理逻辑同步：detach `/Volumes/<VOLNAME>` + 校验挂载点

## 2. 箭头同源锚点对齐

- [x] 2.1 `package.sh` mac 分支：`WIN_*` / `ICON_*` 提为单一来源常量
- [x] 2.2 osascript 引用上述变量（不再硬编码 {150,200}/{450,200}）
- [x] 2.3 `ARROW_CX=(ICON_APP_X+ICON_APPS_X)/2` 传给背景合成脚本（实测居中 x=299±1）
- [x] 2.4 守卫收窄：仅「背景图」一行受 `$BG_SRC` 控制，图标定位/窗口尺寸始终执行（D4）

## 3. 背景图构建期合成

- [x] 3.1 新增纯底模板 `build/darwin/dmg-background.tpl.png`（600×400 #F6F7F9，箭头改由脚本绘制）
- [x] 3.2 新增合成脚本 `scripts/compose-dmg-bg.sh`：swift 内联绘制应用名 + 箭头（坐标参数传入）
- [x] 3.3 降级链：swift 不可用 / 无模板 → 复制模板；无模板无占位 → 跳过（均返回 0）
- [x] 3.4 `package.sh` 合成到临时文件并喂给 staging `.background/bg.png`（不污染入库占位图）
- [x] 3.5 产物 `build/bin/` 已由 `.gitignore` 覆盖；模板 `.tpl.png` 随源码入库

## 4. 回读校验

- [x] 4.1 dmg 产出后二次只读挂载成品，检查 `.DS_Store` 存在 + `grep -a bg.png`
- [x] 4.2 命中 → `✓ 美化已生效`；缺失 → `⚠ 降级产物（无美化）`；均退出码 0
- [x] 4.3 挂载失败 → 报「无法校验：…校准写法可能已随 macOS 版本变化」
- [x] 4.4 `STRICT_POLISH=1` 时降级 → 退出码 1（实测 0→0 / 1→1）

## 5. 文档

- [x] 5.1 README 补：背景模板替换方式（箭头别画进模板）、美化校验输出、`STRICT_POLISH`
- [x] 5.2 README 已知限制补：dmg 挂载卷不在 Finder 侧栏，误关窗口后再双击 `.dmg` 重开

## 6. 验证

- [x] 6.1 真机 GUI 沙箱全链路（真 `wails build` → 新 `package.sh` → 成品 dmg）：
      成品含 `.app`+软链+`.background/bg.png`+`.DS_Store`；背景应用名居中 x=299、
      箭头居中 x=299（期望 300，±1 抗锯齿）
- [x] 6.2 回读校验输出 `✓ 已生成: …（含 Applications 软链 + 拖拽引导，美化已生效）`，
      退出码 0 —— **关键回归通过**（旧脚本因 `mktemp` 挂载点必为「无美化」）
- [x] 6.3 降级路径实测：无模板无占位 → 跳过合成返回 0；`.DS_Store` 无 bg.png → 判「无美化」
- [x] 6.4 模板缺失降级实测：无模板但有占位图 → 复制占位图，dmg 仍可用
- [x] 6.5 `openspec validate mac-dmg-polish --strict` 通过
