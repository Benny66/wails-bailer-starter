# theme-persistence — 技术设计

## Context

Go 层有 `GetTheme()/SetTheme()` 绑定方法 + `config.json` 读写，前端主题只存 Pinia 内存态，两者脱节。主题切换重启即丢。

关键现状：
- `config.Default()` 返回 `Theme: "dark"`，`Load()` 无文件时写默认值 → config.json 里永远有值，但语义错位（"dark" 被当成了"未设"）。
- 前端 `store.initTheme()` 只读 `matchMedia` 系统偏好，从不调 `GetTheme()`。
- `setTheme()` 只改 Pinia + DOM，从不调 `SetTheme()`。

## Goals / Non-Goals

**Goals:**
- 主题状态单一真相 = config.json，Pinia 只是前端镜像缓存。
- 首次启动跟随系统，用户切换后持久化、重启恢复。
- Settings 页成为"前端调 Go"的活范例。

**Non-Goals:**
- 不做窗口大小/位置记忆（下一步，本 change 只跑通主题这一个模式）。
- 不改 wailsjs 绑定生成机制（已有）。

## Decisions

### D1: 空串 = 未设，不加标志位
`config.Default()` 的 `Theme` 改为 `""`。`GetTheme()` 返回空串 → 前端 fallback 到系统偏好；返回非空 → 用 config 值。避免为"是否设置过"额外加 `ThemeSet bool` 字段。

**备选**：加 `ThemeSet bool`——否决，空串更简洁，且让 config 每个字段有统一的"未设"表达。

### D2: 前端启动时序 = GetTheme → fallback → data-theme
`initTheme()` 改为 async：
1. `await GetTheme()`（Go 读 config）
2. 空串 → `matchMedia` 系统偏好；非空 → config 值
3. 设 `document.documentElement` 的 `data-theme`

### D3: 前端切换 = 改 Pinia + data-theme（即时）→ SetTheme 落盘（异步）
`setTheme(v)`：同步改 `theme.value` + `data-theme`（即时反馈），异步 `SetTheme(v)` 落盘。落盘失败打日志、不阻断切换（主题切换不该因写盘失败而卡住）。

### D4: 空串不在前端 store 出现
store 的 `theme` ref 类型保持 `'dark' | 'light'`（非空）。空串只存在于 Go config 和 `GetTheme()` 返回值，前端在 initTheme 里完成"空串 → 具体值"的映射后，store 里永远是具体值。

## Risks / Trade-offs

- **[Risk] GetTheme 异步导致首屏短暂默认色闪烁** → `initTheme` 在 `app.mount` 前 await 完成（main.ts 已 `useAppStore(pinia).initTheme()`），但 initTheme 变 async 后需确保 mount 前完成，否则回退到 CSS 的 `@media` 兜底（已有）。
- **[Risk] SetTheme 落盘失败静默** → 打 `console.error`/slog，不阻断。config 目录不可写是极端情况，不值得为此让主题切换失败。
- **[Risk] 老用户 config.json 里 `theme:"dark"` 与新语义兼容性** → `"dark"` 非空串，会被正确读取，行为不变（老用户默认暗色，新用户跟随系统）。可接受。
- **[Risk] wailsjs 绑定在非 wails 环境（纯 vite dev）下 GetTheme 报错** → initTheme 里 try/catch，调用失败 fallback 系统偏好，保证纯前端开发也能跑。
