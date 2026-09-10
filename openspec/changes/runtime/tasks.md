# runtime — 实施任务

## 1. 生命周期

- [x] 1.1 单实例锁：`options.SingleInstanceLock` + 二次启动唤起窗口
- [x] 1.2 关到托盘：`HideWindowOnClose`（Windows/Linux true，macOS false）
- [x] 1.3 优雅关闭：`OnShutdown` 按顺序回收（停托盘 → 落盘 config → 关 db）

## 2. 日志

- [x] 2.1 引入 lumberjack，实现按大小轮转
- [x] 2.2 文件 + 控制台双写，分级（dev Debug / 生产 Info）

## 3. 配置

- [x] 3.1 建 `internal/config/`：结构体 + json 读写，落用户配置目录
- [x] 3.2 启动加载，不存在写默认值，暴露绑定方法保存

## 4. 对话框

- [x] 4.1 建 `internal/dialog/`：封装 runtime 的四种对话框为绑定方法

## 5. 托盘

- [x] 5.1 引入 energye/systray（RunWithExternalLoop 共存 Wails 主循环），goroutine 内跑
- [x] 5.2 右键菜单 + 双击恢复窗口 + 退出项（Windows/Linux）
- [x] 5.3 三平台图标适配；macOS 降级为 no-op（Wails 架构冲突，见 design D1）

## 6. 崩溃捕获

- [x] 6.1 panic recovery + 崩溃信息落盘 crash 日志

## 7. 验证

- [x] 7.1 `make build` 通过
- [x] 7.2 手动验证：二进制启动稳定、config.json/app.log 落盘、单实例/对话框绑定生成
- [x] 7.3 `openspec validate runtime` 通过
