# fix-theme-sync — 修复主题状态脱节

## Why

设置页主题切换存在两个 bug：初始颜色与预期不一致、切换后按钮高亮不更新。根因是主题状态有四处脱节——Pinia 的 `theme` ref、`<html>` 的 `data-theme`、CSS 渲染、设置页的快照 `ref(store.theme)` 之间没有同步机制，且"默认暗色"与"默认跟随系统"两个设计意图互相矛盾。

## What Changes

- **统一默认口径为"跟随系统"**：store 初始 theme 读系统偏好（`matchMedia('(prefers-color-scheme: dark)')`），而非硬编码 `'dark'`。
- **初始化 data-theme**：应用启动时根据初始 theme 设置 `<html>` 的 `data-theme`，让 store 与 DOM 同步。
- **修复设置页响应式**：`Settings.vue` 用 `storeToRefs(store)` 取 `theme`，替代断链的 `ref(store.theme)` 快照。
- **收编 `theme.css` 的 `@media` 兜底逻辑**：保留 `@media` 作为"未设置 data-theme 时"的兜底，但明确 store 初始化后必有 data-theme。

## Capabilities

### New Capabilities

<!-- 无 —— 纯 bug 修复，不新增能力 -->

### Modified Capabilities

- `design-tokens`: 主题切换的"默认跟随系统"行为与 store 初始化逻辑发生 spec 级变化。

## Impact

- **修改代码**：`frontend/src/stores/app.ts`、`frontend/src/views/Settings.vue`、`frontend/src/main.ts`（或 store 内初始化）。
- **无新增依赖**。
- **破坏性**：无（内部状态同步修复，对外 API 不变）。
