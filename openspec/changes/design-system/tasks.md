# design-system — 实施任务

## 1. 设计令牌

- [x] 1.1 建 `frontend/src/styles/tokens.css`：colors/spacing/radius/shadow/typography/motion 六类令牌
- [x] 1.2 实现单主色派生色阶（color-mix，OKLAB），primary-50…900 + 语义色
- [x] 1.3 提供 color-mix 不兼容时的静态色阶 fallback

## 2. 主题层

- [x] 2.1 建暗色/亮色主题变量映射，默认暗色、跟随系统（`prefers-color-scheme`）
- [x] 2.2 实现 `data-theme` 切换，全站引用 CSS variables

## 3. Element Plus 重主题

- [x] 3.1 引入 element-plus、vue-router、pinia
- [x] 3.2 用 tokens 覆盖 EP 的 CSS 变量（`--el-color-primary` 等）与主题色

## 4. 应用壳

- [x] 4.1 建 `layouts/` 应用壳：侧栏（可折叠）+ 内容主区域
- [x] 4.2 建路由骨架（占位首页/设置，内容由 example-module 填）
- [x] 4.3 Windows 无边框：`main.go` 设 `Frameless` + `windows.Options` + 自绘标题栏（三按钮右）
- [x] 4.4 macOS 无边框：`mac.TitleBarHiddenInset` + 交通灯 + 内容顶到顶
- [x] 4.5 标题栏拖拽区：`--wails-draggable` 自定义属性 + 控件/交通灯 `no-drag`

## 5. 微交互

- [x] 5.1 侧栏折叠过渡动画
- [x] 5.2 路由切换 fade/slide
- [x] 5.3 骨架屏 + 空状态组件

## 6. 验证

- [x] 6.1 `make build` 通过（前端编译 + 无硬编码色值抽查）
- [x] 6.2 `openspec validate design-system` 通过

## 7. Bugfix：macOS 交通灯被 Frameless 吞掉

- [x] 7.1 `Frameless` 拆平台分叉：`frameless_other.go` 返回 `true`、`frameless_darwin.go` 返回 `false`
- [x] 7.2 `main.go` 用 `Frameless: frameless()` 替换全局 `true`，使 macOS 交通灯恢复
- [x] 7.3 验证 macOS 窗口有交通灯 + 透明贴边观感（详见 design D3 的 bug 说明）
