# perf-instrumentation — 种下三件「让优化有据可依」的基础能力

## Why

客户端优化的瓶颈不是技巧，是证据。当前状态（均已实测/查证）：

- **启动只有一个总数**：`应用启动` → `前端已挂载` 两点之差。两次构建产物的实测是
  **359ms 与 733ms**（同一天连续两次运行，2× 差异）。中间包含 webview 启动、
  1.4MB 前端产物的解析执行、以及一次 IPC 往返——**比例完全未知**。
- **`main.ts` 的首帧被一次 IPC 挡着**：第 28 行 `await initTheme()` 在第 30 行 `mount()`
  之前（为避免主题闪烁，是有意的取舍）。但那次往返是 5ms 还是 200ms 无人知道。
  Electron 官方性能文档明确要求「避免阻塞 IPC」，而这里连测都没测。
- **慢查询形同不存在**：`gorm.Open(..., &gorm.Config{})` 用默认 logger → 写 **stdout**
  （打包后无人可见），且 `SlowThreshold` 默认 **200ms**（对本地 SQLite 等于永不触发）。

这与上一轮刚修的「Wails logger 没接线 → 前端日志全进黑洞」是同一类问题：**日志写到了
无人可见的地方**。只是这次漏在数据库这层。

## What Changes

- **P1 启动时间线**：Go 组合根各阶段（日志初始化 / 数据库 / 配置）打点；前端用
  `performance` 记录「到 JS 就绪 / 主题 IPC / 挂载」分段。两段都写进同一个 `app.log`，
  靠绝对时间戳拼成一条时间线——不做跨语言时间戳传递。
- **P2 IPC 往返采样**：在 `invoke()`（绑定的唯一漏斗）里测每次调用耗时；
  超阈值写 warn（带可选调用点标签），并把启动期的聚合（次数/最大/最慢者）放进前端启动行。
- **P3 慢查询接进 app.log**：gorm logger 适配到 slog，阈值按本地 SQLite 重定（非 200ms），
  并排除「记录不存在」（业务有意为之，不该当错误报）。
- **护栏**：gorm logger 是否接线、启动时间线的打点是否存在——这两类删掉都没有任何症状。

## Capabilities

### Modified Capabilities

- `logging`: 启动时间线（Go 阶段打点）；慢查询接入同一条日志。
- `observability`: 前端启动分段 + IPC 往返采样。
- `architecture-guardrails`: 新增「可观测性接线」护栏。

## Non-Goals

- **不做性能面板 UI**（产品面，且现在还没有数据可展示）。
- **不做自动上报 / RUM**（涉及隐私与网络，超出基座边界；Electron 官方文档也未建议）。
- **不做体积预算护栏（原 P4）**：现在基线是 1.4MB，护栏等于没护栏；应与第一次真正的
  优化同批做，否则只是把现状固化。
- **不引入自动取调用点的魔法**：`invoke(fn)` 拿不到绑定名，靠解析函数源码猜名字
  在 `usePagedList` 这类路径下会取到 `fetcher` 这种误导性结果。改为可选显式标签。
- **不做常驻采样线程**（成本高；先有快照级数据即可定位）。

## Impact

- **修改代码**：`main.go`（阶段打点）、`internal/database`（gorm logger 适配）、
  `frontend/src/lib/invoke.ts`（耗时采样）、`frontend/src/main.ts`（启动标记）、
  `frontend/src/composables/usePagedList.ts`（标签透传）、`_example` 与生成物范例、
  `internal/guard/wiring_test.go`。
- **破坏性**：**无**。`invoke` 新增**可选**第二参数；日志新增行不影响既有解析。
