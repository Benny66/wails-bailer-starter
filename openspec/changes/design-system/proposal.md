# design-system — 设计系统与平台感知的应用壳

## Why

地基（foundation）只有一个中性骨架，前端是 `App.vue` 里的一句占位文案。基座要"别人 30 分钟就能做出好看的界面"，就必须先有**一套带默认审美的设计系统 + 一个平台感知的无边框应用壳**——否则后面 example-module 的示例页会各写各的、越写越脏。

## What Changes

- **设计令牌层（Design Tokens）**：`frontend/src/styles/tokens.css`，定义 colors / spacing（4px 基准）/ radius / shadow / typography / motion，作为唯一真相。
- **单主色派生色阶**：只填一个 `--color-primary`，用色阶函数派生 primary-50…900 及 hover/active/边框/浅底/文字强调，全站自动换肤。
- **主题层**：暗色优先（专业工具风），默认跟随系统，`data-theme` 切换，全部引用 CSS variables，禁硬编码 hex 色值。
- **Element Plus 深度重主题**：用 tokens 覆盖 EP 的 CSS 变量，视觉翻新为自定义观感；EP 当功能底座（表格/表单/分页）。
- **平台感知应用壳**：侧栏（可折叠）+ 内容主区域；无边框标题栏按平台拆分——Windows `Frameless` + 自绘三按钮 + 拖拽区 + Aero Snap；macOS `TitleBarHiddenInset` + 原生交通灯 + 内容顶到顶。
- **微交互**：侧栏折叠过渡、路由切换 fade、骨架屏、空状态、按钮 loading 防重复提交。

## Capabilities

### New Capabilities

- `design-tokens`: 设计令牌层、单主色派生色阶、暗/亮主题切换、禁硬编码色值的约束。
- `app-shell`: 平台感知应用壳——侧栏 + 无边框标题栏（Windows 自绘 / macOS 原生交通灯）+ 拖拽区 + 窗口控制按钮。

### Modified Capabilities

<!-- 无 -->

## Impact

- **新增代码**：`frontend/src/styles/`（tokens/theme）、`frontend/src/layouts/`（应用壳）、`frontend/src/components/`（标题栏/侧栏）、Element Plus 引入与主题覆盖。
- **依赖**：element-plus、vue-router、pinia（前端）。
- **Go 侧**：`main.go` 的 `options.App` 增加 `Frameless`、`Mac.TitleBar`、`Windows.Options` 等配置。
- **破坏性**：`App.vue` 从占位文案改为挂载应用壳与路由出口。
