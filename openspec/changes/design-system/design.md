# design-system — 技术设计

## Context

地基已有 Vue3 + Vite 最小骨架，`frontend/src/` 下只有 `App.vue` 占位、`main.ts`、`style.css`（模板遗留的 Nunito 字体与深蓝底色）。`main.go` 用默认窗口（非无边框、默认标题栏）。

目标：立一套设计令牌 + 平台感知应用壳，为后续所有示例页提供统一视觉底座。

## Goals / Non-Goals

**Goals:**
- 设计令牌层成为颜色/间距/圆角/动效的唯一真相。
- 单主色 → 派生全色阶，换肤 = 改 3 个值（name/logo/primary）。
- 暗色优先，默认跟随系统。
- Element Plus 功能底座 + tokens 深度重主题。
- 平台感知无边框壳：Windows 自绘三按钮，macOS 原生交通灯。

**Non-Goals:**
- 不做业务页面（那是 example-module）。
- 不做完整组件库（只用 EP + 少量自定义壳组件）。
- 不做 ⌘K 命令面板（v1.1）。

## Decisions

### D1: tokens 用 CSS Custom Properties，不用 JS 运行时
原生 CSS variables 满足"换肤 = 改变量"，无需额外运行时；`data-theme` 切换只翻一组变量映射。色阶派生用 CSS `color-mix()` 或预置函数，避免引入 chroma-js 这类运行时依赖。

**备选**：JS 运行时主题（theme-provider + JS 计算色阶）——否决，Wails 前端是纯静态，运行时换肤收益低、复杂度高。

### D2: 单主色派生用 CSS `color-mix()` 生成 50–900 色阶
`--color-primary-500` 为输入，其余用 `color-mix(in oklab, primary-500 X%, white/black)` 派生。OKLAB 空间混色观感更均匀。fallback：不支持 color-mix 的老 WebView2 用静态预置色阶。

### D3: 无边框按平台拆分，macOS 不"自绘三按钮"
- **Windows**：`options.App{ Frameless: true }` + `windows.Options{ DisableFramelessWindowDecorations: false }`（保留圆角/阴影）+ 自绘标题栏（三按钮在右）+ `CSSDragProperty` 拖拽区。Aero Snap 由系统在 frameless 下仍支持（需 `WM_NCHITTEST` 处理，Wails 已处理）。
- **macOS**：`mac.TitleBarHiddenInset()` 保留原生交通灯（左上），内容 `FullSizeContent` 顶到顶，不自绘按钮——这是 macOS 桌面应用正确姿势。

**备选**：macOS 也自绘三按钮——否决，违反 macOS 用户习惯，且失去交通灯的原生手感。

> **⚠️ 已发现的 bug（待修）**：`Frameless` 必须按平台分叉，不能设全局 `true`。
> 当前 `main.go` 设了全局 `Frameless: true`，导致 macOS 上交通灯消失——
> 读 `WailsContext.m` 的 `CreateWindow` 确认：`if(!frameless)` 块里才加
> `NSWindowStyleMaskTitled`（交通灯按钮的 mask），`Frameless: true` 时整个块被跳过，
> 于是 `TitleBarHiddenInset` 的 `HideTitleBar:false`/`UseToolbar` 全部失效，交通灯与
> 标题栏都没了。前端 `TitleBar.vue` 的 `isMac` 分支又隐藏自绘三按钮，双失守。
>
> **修复**：`Frameless` 拆平台分叉——`frameless_other.go` 返回 `true`、
> `frameless_darwin.go` 返回 `false`（与 `hideWindowOnClose` 同模式）。macOS 下
> `Frameless:false` + `TitleBarHiddenInset()` 才能既透明贴边、又保留交通灯。
> 这是 design-system 自身引入的 bug，并入本 change 修复，不单独立项。

### D4: Element Plus 当功能底座，不换组件库
EP 提供表格/表单/分页/弹窗等脏活累活，用 tokens 覆盖其 CSS 变量（`--el-color-primary` 等）与少量组件样式，视觉翻新为自定义观感。下游永远只看到自定义外观。

**备选**：换 Naive UI/Arco——否决，推翻功能说明锁定，且失去 EP 生态。

### D5: 微交互用 CSS transition + Vue `<Transition>`，不引动画库
侧栏折叠、路由切换、骨架屏用原生 transition（motion tokens 定义 duration/easing），避免 framer-motion 等重依赖。

## Risks / Trade-offs

- **[Risk] color-mix 在老 WebView2 不支持** → 提供静态预置色阶 fallback，检测到不支持则回退。
- **[Risk] Windows frameless 的 Aero Snap 在个别版本失效** → 记录已知限制，必要时用 `WM_NCHITTEST` 手动处理（v1.1）。
- **[Risk] macOS 交通灯与自绘拖拽区重叠** → 拖拽区 CSS 避开左上角交通灯区域（`-webkit-app-region: drag` 只加在安全区）。
- **[Risk] EP 重主题覆盖不全（某些组件残留默认蓝）** → 用 EP 的 CSS 变量 + `theme.css` 覆盖，抽查所有用到组件。
