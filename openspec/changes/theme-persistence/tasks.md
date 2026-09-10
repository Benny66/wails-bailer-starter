# theme-persistence — 实施任务

## 1. Go config 空串语义

- [ ] 1.1 `config.Default()` 的 `Theme` 改为 `""`（空串 = 未设 = 跟随系统）

## 2. 前端启动读 config

- [ ] 2.1 `store.initTheme()` 改为 async：`GetTheme()` → 空串 fallback 系统偏好 → 设 data-theme
- [ ] 2.2 initTheme 加 try/catch，非 wails 环境（纯 vite dev）GetTheme 失败时 fallback 系统偏好

## 3. 前端切换写 config

- [ ] 3.1 `store.setTheme()` 改为同步改 Pinia + data-theme，异步 `SetTheme()` 落盘

## 4. 联调范例

- [ ] 4.1 Settings 页 import wailsjs 绑定，作为"前端调 Go"的活范例

## 5. 验证

- [ ] 5.1 `make build` 通过（前端 vue-tsc + Go 编译）
- [ ] 5.2 手动验证：首次启动跟随系统 → 切主题 → 重启保持
- [ ] 5.3 `openspec validate theme-persistence` 通过
