# frontend-pipeline-parity — 前端侧管道补齐 + 让「单一真相」会红

## Why

`pipeline-contract` 给管道立了三份契约（错误 / 分页 / 事件），但**只落了 Go 半边**。
下游真正开工时撞到的是另一侧的空缺：

- **事件契约前端零对应物**。`internal/event` 有动作常量与 `Name(domain, action)`，
  前端 `src/lib/` 只有 `invoke.ts`。下游要收进度事件只能手拼 `EventsOn('asset:progress', …)`
  ——**第二处真相**（AGENTS.md 铁律 1）；Wails 的 `EventsOn` 返回取消函数、必须显式解绑，
  不知道的人会漏，路由来回切即叠加订阅（重复处理 + 内存泄漏）。
  `pipeline-contract/tasks.md` 3.2 自认「未在其中塞入长任务范例」——这是已知缺口。
- **分页契约前端无镜像**。`DefaultPageSize` / `MaxPageSize` 只写在 `internal/page`，
  前端没有对应常量与类型；`_example` 里 39 行的列表页有 **20 行是纯分页样板**，
  且 `make gen` 每生成一个模块就复制一份。
- **「单一真相」只写在注释里**。`invoke.ts` 的 `ErrorCode` 自称「单一真相在 Go，此处为镜像」，
  但**没有任何检查**阻止两边漂移。AGENTS.md 铁律 1 标了 ⚙️，实际只覆盖模型清单与依赖清单。
- **数据目录逻辑重复 4 处**。`config` / `database` / `logging` / `crash` 各写了一遍
  `os.UserConfigDir()` + `MkdirAll(appName)`。改一处就分叉——铁律 1 的正面违反。
- **出问题够不到日志**。日志落在用户配置目录，但没有任何绑定方法能打开它。
  用户报障时开发者只能口头教三套平台的路径。
- **`make lint` 不查 TS 类型**。只跑 ESLint；`vue-tsc` 只在 `npm run build` 里，
  而开发态 `vite` 不做类型检查——类型错误一路漂到打包才炸（铁律 4：验证入口统一）。

**后果**：契约在 Go 侧很硬，到前端就软了。下游绕过契约的代价比遵守低，
契约就会在第一个真实模块上开始漏。

## What Changes

- **前端事件契约镜像**：`frontend/src/lib/event.ts`（动作常量镜像 + `eventName()` +
  payload 类型 + `onEvent()` 返回取消函数），`frontend/src/composables/useEvent.ts`
  （组件卸载自动解绑）。文件头给出可抄的「长任务推进度」范例（Go 4 行 + 前端 6 行）。
- **前端分页契约镜像**：`frontend/src/lib/page.ts`（`PageRequest`/`PageResult<T>` 类型 +
  页大小常量镜像），`frontend/src/composables/usePagedList.ts`（收敛列表页的加载/分页/错误状态）。
  **只镜像常量与类型，不镜像 `Normalized()` 的归一化逻辑**——一个规则两个实现比复制常量更糟，
  归一化仍唯一地活在 Go 侧。
- **一致性护栏**：新增 `internal/guard/parity_test.go`，解析 Go 侧常量与前端 TS 镜像，
  逐项断言一致；解析到 0 个结果 Fatal（铁律 5）。覆盖错误码、页大小上下界、事件动作三类。
  **这让铁律 1 第一次由「注释承诺」变成「会红的检查」。**
- **数据目录单一真相**：新增 `internal/appdir`（唯一出处），`config`/`database`/`logging`/`crash`
  四处改引用它。
- **数据目录可达**：新增 `internal/reveal`（跨平台打开目录）+ 绑定 `GetDataDir()`、
  `OpenDataDir()`，设置页加一行演示。
- **验证入口补齐**：`make lint` 增加 `vue-tsc --noEmit` 类型检查。
- **范例与生成物据新契约改写**：`_example/frontend/ExampleList.vue` 改用 `usePagedList`。

## Capabilities

### Modified Capabilities

- `pipeline-contract`: 补齐分页/事件契约的**前端侧镜像**与消费辅助；错误码镜像纳入护栏。
- `architecture-guardrails`: 新增「镜像一致性护栏」（Go 单一真相 ↔ 前端镜像）。
- `code-generator`: 生成的前端页面改用 `usePagedList`，不再复制分页样板。
- `app-config`: 数据目录路径收敛为单一真相，并新增打开数据目录的绑定能力。

## Non-Goals

- **不做产品面**：不加 toast / 表格 / 分页控件 / 进度条组件。`usePagedList` 只管状态与
  加载，不管渲染——渲染仍是下游的事。
- **不引入前端测试框架**（vitest 等）。本次的价值在「镜像不漂移由 Go 护栏保证」，
  而不是给前端补单测；引入测试栈应单独立项。
- **不做自动更新**、不碰打包/发布链路。
- **不改事件协议的形态**（仍是约定式）。只补前端消费侧，不动 Go 侧约定。

## Impact

- **修改代码**：`internal/{config,database,logging,crash}`（改引用 `appdir`）、
  `app.go`（新增两个绑定方法）、`frontend/src/views/Settings.vue`、
  `_example/frontend/ExampleList.vue`、`Makefile` + `frontend/package.json`。
- **新增资源**：`internal/appdir`、`internal/reveal`、`frontend/src/lib/{event,page}.ts`、
  `frontend/src/composables/{useEvent,usePagedList}.ts`、`internal/guard/parity_test.go`。
- **破坏性**：**无**。绑定方法只增不改；前端镜像为新增文件；`appdir` 抽离保持路径不变
  （同一 `os.UserConfigDir()/<appName>`），历史数据无需迁移。
