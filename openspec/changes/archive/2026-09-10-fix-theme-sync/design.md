# fix-theme-sync — 技术设计

## Context

主题状态存在四处脱节（详见 proposal），导致设置页两个 bug。本 change 修复状态同步，并把"默认跟随系统"落成唯一口径。

## Goals / Non-Goals

**Goals:**
- 初始主题 = 系统偏好，且 store / data-theme / CSS 三者同步。
- 设置页按钮高亮随切换更新（真响应式）。
- 手动选择后不再跟随系统（可选增强，见 D4）。

**Non-Goals:**
- 不做主题持久化到后端 config（当前 `app.go` 已有 `GetTheme`/`SetTheme` 绑定，但前端 store 未接，本 change 不接后端持久化，仅修前端状态同步）。
- 不改设计令牌结构。

## Decisions

### D1: 初始主题读系统偏好
`stores/app.ts` 初始化时用 `matchMedia('(prefers-color-scheme: dark)').matches` 决定初始 theme，替代硬编码 `'dark'`。

**备选**：硬编码 `'dark'`——否决，与"默认跟随系统"矛盾（正是 bug 3 的根源）。

### D2: 应用启动时同步 data-theme
在 `main.ts` 挂载后（或 store 首次实例化时）调用一次 `document.documentElement.setAttribute('data-theme', initialTheme)`，消除"store 有值但 DOM 无属性"的脱节。

### D3: Settings.vue 用 storeToRefs
`const { theme } = storeToRefs(store)` 替代 `const theme = ref(store.theme)`。后者是值快照，断链；前者是真响应式引用。

### D4: 手动选择后不跟随系统（收编 @media）
`theme.css` 的 `@media (prefers-color-scheme: light) :root:not([data-theme])` 保留——它只在"无 data-theme"时兜底。由于 D2 保证启动后必有 data-theme，这条 @media 实际只影响"JS 未加载"的极早期瞬间。手动选择后，data-theme 存在，@media 不会覆盖，天然满足"手动后不跟随系统"。

## Risks / Trade-offs

- **[Risk] matchMedia 在 Wails webview 里返回 null/异常** → 加空值兜底：`window.matchMedia?.(...)?.matches ?? true`，异常时默认暗色。
- **[Risk] store 初始化时机早于 DOM 可写** → `document.documentElement` 在 store 首次实例化（组件 setup 时）已可用，无风险；但为稳妥，D2 放 main.ts 的 mount 后执行。
