# runtime-pipeline — 设计

## D1：窗口几何只持久化尺寸与最大化，**不**持久化位置

**技术事实（已查 Wails v2.15.0 源码确认）**：

- `options.App` **没有**初始位置字段（只有 `Width`/`Height`/`WindowStartState`）。
  尺寸与最大化可以在**创建期**给定 → 窗口首帧就是对的，零跳动。
- 位置只能运行期 `runtime.WindowSetPosition`。但 `OnStartup` 是在 **goroutine** 里跑的
  （`internal/frontend/desktop/darwin/frontend.go:248`），紧接着 `mainWindow.Run()` 就显示窗口
  —— 在 `OnStartup` 里设位置与窗口显示**存在竞态**，用户可能看到窗口跳一下。
  要消除跳动只能 `StartHidden: true` + 设位置 + 手动 `WindowShow`，这会改变应用启动语义
  （并牵动托盘/`HideWindowOnClose` 的交互）。

**决策**：本期只做尺寸 + 最大化。位置留待单独评估，且必须一并做**屏幕边界夹取**
（`ScreenGetAll`）——否则用户拔掉外接显示器后，窗口会还原到不存在的坐标上，
表现为「应用启动后看不见」，而且他不知道要去改 `config.json`。这类 bug 的代价
远高于「位置没记住」。

**取值合法性**：`config.json` 是用户可编辑的。窗口尺寸必须夹取
（宽高小于下限或无意义时退回默认 1024×768），否则一个手改坏的配置会让窗口不可用。

## D2：跨实例用「绑定 + 事件」两条路，覆盖两种时序

二次启动有两个时刻需要区分：

| 时刻 | 通路 |
|---|---|
| 首次启动的 argv（应用尚未运行） | 绑定 `GetLaunchArgs()`，前端挂载后主动查 |
| 二次启动的 argv（应用已在运行） | `app:second-instance` 事件推送 |

**为何不统一用事件**：首次启动时事件会在前端订阅之前触发，**必然丢**（启动竞态）。
用绑定查则没有时序问题。二次启动时窗口已在前台、前端已挂载，事件是安全的。

事件名用 `event.Name("app", "second-instance")`，载荷为
`{args: []string, working_dir: string}` —— 复用既有事件契约，不新造一套。

`app.ctx` 在 `OnSecondInstanceLaunch` 里可能尚未赋值（`startup` 在 goroutine 中），
故发送前判空，未就绪只记日志不 panic。

## D3：Wails logger 接 slog —— 补的是一个**静默的黑洞**

Wails 的日志链路（已读源码确认）：

```
JS LogError(msg)
  → window.WailsInvoke('L' + 'error' + msg)
  → Go 侧 ctx 里的 *internal/logger.Logger
  → 按 LogLevel 过滤后 Sprintf 成字符串
  → options.App.Logger（pkg/logger.Logger 接口，7 个方法）
```

`options.App.Logger` 为空时 Wails 用 `NewDefaultLogger()`——**只写 stdout**。
打包后的 GUI 应用 stdout 无人可见，于是：

- 前端的 `LogError/LogInfo` 全部石沉大海；
- Wails 自身的内部日志（资源加载失败、前端 panic、绑定错误）同样石沉大海。

**决策**：实现一个 7 方法的适配器转发到 `slog`，注入 `options.App.Logger`；
同时设 `LogLevel: DEBUG`（dev）与 `LogLevelProduction: INFO`（`wails build` 默认带
`production` tag），与既有 slog 分级对齐。

- `Print` → `slog.Debug`（Wails 用它输出无级别信息，多为启动细节）
- `Trace/Debug` → `slog.Debug`；`Info` → `slog.Info`；`Warning` → `slog.Warn`
- `Error` → `slog.Error`
- `Fatal` → `slog.Error` + **不** `os.Exit`：Wails 在调用完 `Fatal` 后会自己 `os.Exit(1)`，
  我们再退一次会让 defer（日志 flush、数据库关闭）来不及跑。

每条日志带 `source=wails` 属性区分来源（结构化字段而非拼进消息，便于 grep 与过滤），
排查时能一眼看出是前端/框架发的还是 Go 业务发的。

**实跑验证**：构建产物启动后 `app.log` 出现
`level=INFO msg=前端已挂载 source=wails` —— 前端日志确实经这条链路落到了文件里。

## D4：前端全局错误兜底 —— 生产态唯一的出口

三处都要挂，缺一不可：

| 钩子 | 覆盖 |
|---|---|
| `window.onerror` | 同步运行时错误（含资源加载失败） |
| `window.onunhandledrejection` | 未 catch 的 Promise（绑定调用漏 catch 是重灾区） |
| `app.config.errorHandler` | Vue 组件树内的渲染/生命周期异常（前两个覆盖不到） |

处理逻辑：用既有 `normalizeError` 归一化为 `{code, message, detail}` → 经 `lib/log.ts`
写后端日志。**不做 UI**（toast 属产品面）——但保留了 `code` 与 `detail`，
下游要接管时只需替换 `installGlobalErrorHandler` 的一个回调。

同时**保留原行为**：仍然 `console.error` 一份，开发态在 DevTools 里照常可见。

## D5：数据库导出用 `VACUUM INTO` + 原子改名

- `VACUUM INTO '<file>'` 是 SQLite 官方的一致性快照方式：不必停写、不必自己处理
  页锁与 WAL，产出的是一个完整可用的库文件。
- **限制**：目标文件**必须不存在**，否则报 `output file already exists`。
  而下游通常先用系统「保存」对话框拿路径——**用户在那里确认过覆盖**，
  我们再报「文件已存在」是打自己的脸。
- **决策**：先 `VACUUM INTO '<target>.tmp'`，成功后 `os.Rename` 覆盖到目标路径
  （同目录，同文件系统，rename 是原子的）。失败路径清理临时文件。
- 校验：导出目标不能落在数据目录内（否则等于是把库往自己身上复制，
  且可能被下次导出的通配清理误伤）。用 `filepath` 判断前缀，拒绝并返回 `apperr.Validation`。

## D6：版本来源分层，不硬造

`release.sh` 现在拿的版本号只用于**产物文件名**，二进制里没有任何版本。
本期取分层回退：

1. `-ldflags "-X <module>/internal/appinfo.InjectedVersion=..."` 注入 → 有则用
2. 否则读 `debug.ReadBuildInfo()` 的 `vcs.revision` 取短哈希
3. 否则 `dev`

**实测修正（原设计有误）**：第 2 层在 wails 构建里**永远不可用**——
`wails build` 会给 `go build` 传 `-buildvcs=false`，VCS 信息根本不写进二进制。
首次实跑构建产物时日志显示 `version=dev`，才暴露这一点。

**因此第 1 层不是「锦上添花」而是必需品**：`package.sh` 与 `release.sh` 都改为
注入 `git describe --tags --always --dirty` 的结果（打过 tag 得 `v1.2.3`，
未打 tag 得提交短哈希，始终可得）。`make build`（裸 `wails build`）仍是 `dev`，
可接受——它不是发布路径。

**连带发现：`-X` 对不存在的符号是静默忽略的**。写错符号名时构建成功、无警告，
变量保持零值，版本悄悄退回 `dev`——又一个「没有任何症状」的失效模式。
故注入目标由 `internal/guard/wiring_test.go` 核对（符号必须真实存在），
模块路径也改为从 `go.mod` 现读，不假定它等于应用名。

`GetAppInfo()` 一并返回数据目录与日志路径——排障时用户只需念一次这个返回值。
启动首行日志写入版本，使**每个** `app.log` 自带版本上下文。
