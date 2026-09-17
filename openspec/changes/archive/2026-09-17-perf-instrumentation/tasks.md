# perf-instrumentation — 实施任务

## 1. P3 慢查询接 slog

- [x] 1.1 `internal/database/gormlog.go`：gorm logger 接口适配到 slog
      （阈值 50ms，非 gorm 默认的 200ms；排除 ErrRecordNotFound；SQL 压平并按字符截断）
- [x] 1.2 `database.Init` 的 `gorm.Config` 显式设置 `Logger`
- [x] 1.3 单测 6 项：正常查询静默 / 慢查询告警 / 记录不存在不记 / 真实错误记 error /
      超长 SQL 压平与截断（含多字节不被切断）/ LogMode 不改变行为

## 2. P1 启动时间线

- [x] 2.1 `main.go`：`bootTimer` + 三个阶段打点（日志初始化 / 数据库初始化 / 配置加载与组装）
- [x] 2.2 `main.ts`：前端分段（to_js 取自导航计时、主题 IPC、挂载、总计）
- [x] 2.3 启动行带出 IPC 聚合（次数 / 最大耗时 / 最慢者）

## 3. P2 IPC 往返采样

- [x] 3.1 `invoke(fn, label?)`：每次调用计时；超 50ms 记 warn（含标签）；`finally` 中统计
      （失败的调用同样可能慢，而那正是最需要看到的一次）
- [x] 3.2 聚合统计 `ipcStats()` 供启动行一行带出，不逐条记录
- [x] 3.3 标签显式传入，**不自动猜名字**（`usePagedList` 路径下会猜成 `fetcher`）
- [x] 3.4 `usePagedList` 透传 `label`；`_example` 范例页演示该用法

## 4. 前端模块拆分（解开循环依赖）

- [x] 4.1 新增 `lib/error-report.ts`：全局兜底 + Vue 错误钩子（从 log.ts 迁出）
- [x] 4.2 `log.ts` 收窄为纯日志出口（只依赖 Wails 运行时）
- [x] 4.3 依赖方向变为单向：`invoke → log`、`error-report → {invoke, log}`，
      归一化逻辑仍只有一份（未复制）
- [x] 4.4 `main.ts` 改从 error-report 导入

## 5. 护栏

- [x] 5.1 `TestGormLoggerIsWired`：断言 `gorm.Config` 设置了 `Logger`
- [x] 5.2 `TestStartupTimelineStagesExist`：断言启动打点覆盖三个关键阶段
- [x] 5.3 两处均「解析到 0 个即 Fatal」+ 破坏性验证会红

## 6. 验证

- [x] 6.1 `make test` / `make lint` 通过；前端 typecheck + ESLint 通过
- [x] 6.2 **实跑三次**：启动时间线正确输出，且得出 D7 的结论
      （主导成本是窗口/webview 创建 1.4–2.3s，不是前端包体）
- [x] 6.3 实跑发现并修复 `ipc_calls=0`（stores/app.ts 绕过 invoke）+ SetTheme 错误只写 console
- [ ] 6.4 未做：多变体采样测 EP 的真实占比（to_js 运行间波动 1.7×，单次对比不可靠）

## 7. 文档与归档

- [x] 7.1 `docs/map.md`：补 `lib/error-report.ts`
- [x] 7.2 `docs/脚手架功能说明.md`：补启动时间线与慢查询
- [ ] 7.3 归档 + 提交
