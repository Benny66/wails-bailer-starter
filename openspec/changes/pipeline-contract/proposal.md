# pipeline-contract — 给管道立契约（错误/分页/事件）

## Why

本脚手架的定位是「**只给管道，产品面下游自己长**」。据此重新审视，管道的唯一产品就是
**契约**——下游要在契约上盖楼。但当前仓库：

- **零契约定义**。全仓没有一处定义 Go↔TS 的错误协议、分页协议、事件协议。
- **绑定的错误直接悬空**。绑定方法（`App.ListXxx`）返回 `(T, error)`，Wails 把它变成
  rejected promise；前端 `await` 时没有任何约定形态，**用户看不到任何东西**。
- **分页无结构**。`_example` 的 `ListExamples()` 返回裸 `[]model.Example`，
  注释自己承认「Wails 绑定只暴露多返回值的第一个，分页总数需封装成结构体」——但**没封**。
- **事件通道零使用**。Wails 的 `EventsEmit/EventsOn`（Go 跑重活、推进度给前端）是它相对
  Electron 的关键能力，本仓一次都没用，也无命名/payload 约定。
- **规范文档在漂移**。`docs/代码规范.md` 写着「由 controller 转成响应」「API 定义集中在
  `api/index.js`」——本仓没有 controller 层，也没有 `api/index.js`。是从别的 Web 项目
  泄漏进来的遗骸。

**后果**：下游 A/B/C 各自发明错误码、分页结构、事件命名 → 脚手架「管道价值」从第一天起
就在漏。且 AGENTS.md 明说读者是「AI 参与生成代码」——没有契约，AI 会**幻觉一套**。

> 与 `mac-dmg-polish` 同源：**没有可验证的契约，就一定会漂移**（那次是 dmg 美化静默失效
> 一个月无人察觉）。

## What Changes

- **错误协议（结构化）**：定义统一的业务错误形状（`code` + 中文 `message` + 可选 `detail`），
  提供 Go 侧错误类型与前端侧 `invoke` 包装，使「业务错误 vs 系统错误」在前端可区分。
- **分页协议（结构化）**：定义 `PageRequest`（page/page_size）与 `PageResult[T]`
  （list/total/page/page_size），并让生成器产出的列表方法默认返回分页结果。
- **事件协议（约定式）**：定义事件命名规范（如 `<domain>:<action>`）与 payload 约定，
  并在范例里给出一个「长任务推进度」的可抄写法。
- **契约护栏**：把上述协议中「可判定」的部分编译成会红的检查（对齐 AGENTS.md 铁律 5
  「护栏必须感知自己瞎了」）。
- **规范文档纠偏**：删除 `docs/代码规范.md` 里 controller / `api/index.js` 的遗骸，
  替换为本仓真实的绑定层契约。

## 契约硬度（已定 —— 分层，因三类性质不同）

| 契约 | 硬度 | 理由 |
|---|---|---|
| 错误 | **结构化** | 下游必须能区分业务/系统错误；形状错了要改所有调用点 |
| 分页 | **结构化** | 盖楼的数据形状，晚改代价最大 |
| 事件 | **约定式** | 事件本质灵活，硬套结构会变成教条 |

## Capabilities

### New Capabilities

- `pipeline-contract`: Go↔TS 的错误协议、分页协议、事件协议，及其护栏。

### Modified Capabilities

- `architecture-guardrails`: 新增「契约护栏」——把错误/分页协议中可判定的部分编译成检查。
- `code-generator`: `make gen` 产出的列表方法改为返回分页结果；绑定方法遵循错误协议。
- `example-module`: 黄金范例据新契约重写（分页列表 + 错误处理 + 事件范例）。
- `delivery-docs`: `docs/代码规范.md` 纠偏（删 controller / `api/index.js` 遗骸）。

## Non-Goals

- **不补产品面**（表格/表单/进度条 UI/toast 组件）——按定位「产品面下游自己长」。
- **不做自动更新**——超出管道边界（另见：应把 `docs/脚手架功能说明.md` 里的相关承诺删掉）。
- 不改 `mac-dmg-polish` 涉及的内容。

## Impact

- **修改代码**：`internal/`（新增负责错误与分页的管道包）、`app.go` 绑定形态、
  `_example/`（范例重写）、`scripts/gen.sh`（生成分页方法）、`internal/guard/`（新增护栏）。
- **新增资源**：前端 `invoke` 包装（错误呈现的第一跳，但**不是** UI 组件）。
- **破坏性**：**有** —— 绑定方法签名从 `([]T, error)` 变为分页返回，属 **breaking**。
  但本仓尚无任何真实业务模块（`_example` 未实例化），代价最小，**现在定最省**。
