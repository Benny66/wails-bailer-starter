# runtime-pipeline — 补全桌面运行时管道（窗口 / 跨实例 / 可观测 / 数据安全）

## Why

重新按「下游开工第一天需要什么」盘点，发现基座的运行时管道有几处**半截**：

- **窗口几何不持久化**。`config.json` 只有 `theme`。用户把窗口拉大、最大化，重启回到 1024×768。
  这是桌面应用最容易被抱怨、也最容易修的一处——基座完全不碰。
- **单实例参数被丢弃**。「打开方式」二次启动时 `data.Args` **只写进日志就没了**
  （`main.go` 的 `OnSecondInstanceLaunch`），前端拿不到。命令行的信息进不了应用。
- **前端日志是黑洞**。Wails 的 `options.App.Logger` **没接线**，于是：
  - 前端 `LogError/LogInfo` 走 Wails 默认 logger → 只到 stdout；
  - Wails 自身的内部日志（资源加载失败、前端异常）同样只到 stdout。
  打包后 stdout 无人可见 —— 这些日志**一条都没进 `app.log`**。
  更糟的是前端连全局错误兜底都没有：`window.onerror` / `unhandledrejection` /
  `app.config.errorHandler` 一个都没挂，生产态（无 DevTools）异常彻底静默。
- **数据没有任何导出能力**。整个库文件删了就全没了，用户换机/备份无从下手。
- **二进制不知道自己是哪个版本**。`release.sh` 的版本号**只用在产物文件名上**，
  没有注入任何地方；`app.log` 里没有版本号。用户报障说「我用的是 0.3」，无从核对。

这五处与「只给管道」的定位一致——都是运行时能力，不是产品面。缺了它们，
下游每个项目都要各自补一遍，而且多半会补漏（尤其是 logger 接线与全局错误兜底，
漏了不会有任何症状，直到用户丢日志来报障）。

## What Changes

- **窗口几何持久化**：`config.json` 增加窗口宽高与最大化状态；启动时作为创建参数
  （无跳动），退出时回写。**不持久化位置**——理由见 design D1（Wails 无创建期位置选项，
  运行期 `WindowSetPosition` 与窗口显示存在竞态，且需屏幕边界夹取否则窗口可能落在已拔掉的显示器上）。
- **跨实例通信**：首次启动的 argv 经 `GetLaunchArgs()` 可查；二次启动的 argv 经
  `app:second-instance` 事件推给前端（复用事件契约）。
- **可观测性**：
  - 把 Wails 的 logger 接到 `slog`，前端与 Wails 内部的日志**全部进 `app.log`**；
  - 新增 `frontend/src/lib/log.ts`（前端日志出口，非 wails 环境安全退化）
    与全局错误兜底（`window.onerror` / `unhandledrejection` / Vue `errorHandler`），
    统一归一化后写后端日志；
  - 启动首行日志记录版本与数据目录（排查时一眼可见）。
- **数据安全**：新增 `ExportDatabase(targetPath)`（SQLite `VACUUM INTO` + 原子改名），
  与既有 `SelectSaveFile` 组合即可做「导出/备份」。
- **版本可追溯**：`GetAppInfo()` 返回版本 / 提交 / 平台 / 数据目录 / 日志路径；
  版本支持 `-ldflags -X main.version` 注入（`release.sh` 用上它），未注入时回退到
  Go build info 的 VCS 提交号，再回退 `dev`。
- **护栏**：断言 `main.go` 的 `options.App` 字面量**必须设置 `Logger`**
  ——这类接线漏了没有任何症状，只能靠检查兜住。

## Capabilities

### Modified Capabilities

- `app-lifecycle`: 窗口几何持久化；二次启动参数转发（前端可见）。
- `app-config`: 新增窗口几何配置字段与取值合法化。
- `logging`: Wails logger 接入 slog（前端/框架日志入文件）；启动记录版本。
- `data-persistence`: 数据库导出（`VACUUM INTO` + 原子替换）。
- `architecture-guardrails`: 新增「Wails 接线」护栏。

### New Capabilities

- `observability`: 前端日志出口与全局错误兜底；应用信息（版本/路径）可查。

## Non-Goals

- **不持久化窗口位置**（理由见上，含 off-screen 风险）。
- **不做自动更新、不做文件关联/双击打开文件**（macOS 走 `open-file` 事件，与 argv 模型差异大，
  属下游业务决策）。
- **不做应用内菜单**（`options.App.Menu` 保持默认，避免与无边框标题栏/托盘交互）。
- **不加 toast/关于页 UI**：`observability` 只提供管道，界面由下游长。
- **不做日志上报/远端采集**（涉及隐私与网络，超出基座边界）。

## Impact

- **修改代码**：`main.go`（窗口创建参数、Logger 接线、二次启动转发）、`app.go`（新绑定）、
  `internal/config`（窗口几何字段）、`internal/logging`（logger 适配器 + 启动记录）、
  `internal/service`（导出）、`frontend/src/main.ts`（安装全局兜底）、`Settings.vue`。
- **新增资源**：`frontend/src/lib/log.ts`、`internal/guard/wiring_test.go`。
- **破坏性**：**无**。`config.json` 只增字段（旧文件缺字段即取默认值）；绑定只增不改。
