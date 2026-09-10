# theme-persistence — 主题持久化与前端↔Go 联调范例

## Why

脚手架交付了 Go 层的 `GetTheme`/`SetTheme` 绑定方法和 `config.json` 读写能力，但前端从没调用过——主题切换只改 Pinia 内存态，重启即丢；config.json 成了死代码。这暴露两个断层：①前端与 Go 的通信路断了（Go 有 6 个绑定方法，前端零调用）；②主题状态有两个互不相认的真相源（Pinia 内存 vs config 磁盘）。本 change 用"主题持久化"这个最小案例打通两者，并为后续所有"前端调 Go + 状态持久化"立下范式。

## What Changes

- **config 空串语义**：`config.Default()` 的 `Theme` 从 `"dark"` 改为 `""`（空串 = 未设置 = 跟随系统），确立"空值即未设"的统一约定。
- **前端启动读 config**：`store.initTheme()` 改为先 `GetTheme()`（Go 读 config），空串则 fallback 到系统偏好，非空则用 config 值。
- **前端切换写 config**：`store.setTheme()` 在改 Pinia + data-theme 的同时调 `SetTheme()`（Go 落盘 config.json）。
- **联调范例固化**：Settings 页成为"前端调 Go"的活范例（import wailsjs 绑定 + async 调用 + 错误处理）。

## Capabilities

### New Capabilities

<!-- 无 —— 不新增能力，修两个既有能力的断层 -->

### Modified Capabilities

- `app-config`: `Config.Default()` 的 Theme 空串语义（"空值即未设"）——spec 级行为变化。
- `design-tokens`: 主题初始值来源从"纯系统偏好"变为"config 优先、空串 fallback 系统偏好"——spec 级行为变化。

## Impact

- **修改代码**：`internal/config/config.go`（Default 空串）、`frontend/src/stores/app.ts`（initTheme 调 GetTheme、setTheme 调 SetTheme）、`frontend/src/views/Settings.vue`（示范联调）。
- **无新增依赖**（wailsjs 绑定已存在）。
- **破坏性**：已有用户的 config.json 里 `theme: "dark"` 仍有效（非空串，会被读取）；新用户首次启动 config 写 `theme: ""`，语义从"默认暗色"变为"默认跟随系统"。这是预期的行为变化。
