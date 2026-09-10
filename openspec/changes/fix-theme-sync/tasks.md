# fix-theme-sync — 实施任务

## 1. store 初始主题跟随系统

- [x] 1.1 `stores/app.ts` 初始 theme 用 `matchMedia('(prefers-color-scheme: dark)')` 决定，含空值兜底

## 2. 启动时同步 data-theme

- [x] 2.1 `main.ts` 挂载后根据初始 theme 设置 `document.documentElement` 的 `data-theme`

## 3. 修复设置页响应式

- [x] 3.1 `Settings.vue` 用 `storeToRefs(store)` 取 `theme`，替换断链的 `ref(store.theme)`

## 4. 验证

- [x] 4.1 `make build` 通过（前端 vue-tsc + ESLint）
- [x] 4.2 逻辑走查确认：初始跟随系统、切换后按钮高亮同步（storeToRefs 真响应式）
- [x] 4.3 `openspec validate fix-theme-sync` 通过
