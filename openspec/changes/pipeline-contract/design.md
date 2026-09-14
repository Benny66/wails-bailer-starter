# pipeline-contract — 技术设计

## Context

管道定位下的契约空白（详见 proposal Why）。当前绑定面极小且**无一个返回列表**：

```
frontend/wailsjs/go/main/App.d.ts
  Confirm(a,b): Promise<boolean>          ← 无 error 通道
  GetTheme(): Promise<string>
  SelectFile/Directory/SaveFile: Promise<string>
  SetTheme(): Promise<void>
```

`frontend/wailsjs/go/models.ts` **尚不存在**（从未生成过模型绑定）。
`_example/service` 的 `ListExamples()` 返回裸 `[]model.Example`，注释已自认需要封装分页。

## Goals / Non-Goals

**Goals:**
- 定义错误 / 分页 / 事件三类契约，落地为**可编译、可断言**的形态。
- 让下游「照着契约盖楼」时**不需要重新发明**这三样。
- 把契约中可判定的部分编译成护栏（对齐 AGENTS.md 铁律 5）。
- 纠正 `docs/代码规范.md` 的遗骸。

**Non-Goals:**
- 不补产品面 UI（按定位）。
- 不做自动更新。

## 承重未知数（已实测 —— 2026-09-14 / Wails v2.15.0）

### U1: Wails v2 绑定是否支持 Go 泛型？ → **✅ 支持**

沙箱实测：`func (a *App) ProbePage() (probe.PageResult[probe.Item], error)` 绑定生成成功。

```ts
// App.d.ts
export function ProbePage(): Promise<probe.PageResult_<pkg>_<param>_>;
// models.ts —— 字段与嵌套类型正确
export class PageResult_pcsbx_internal_probe_Item_ {
    list: Item[];        // ← 类型安全，不是 any[]
    total: number; page: number; page_size: number;
}
```

**发现**：Wails 把泛型实例**压平成含模块路径 + 类型参数的类名**（`PageResult_<pkg>_Item_`）。
类型安全，但名字冗长难看。

**决策**：**采用泛型**（`PageResult[T]`）。理由：
- 下游拿到的是类型安全的 `list: Asset[]`，不是 `any[]`；
- 压平类名只在 TS 类型标注处可见，**不泄漏到下游业务代码**；
- 「每模块专属类型」（原降级方案 C）由泛型**自动达成**，无需 `gen.sh` 额外派生。

### U2: Wails 把 Go 的 error 以什么形态传给前端？ → **❌ 默认拍平成字符串，但有官方钩子**

**证据链（源码 + 运行时 JS + Node 模拟）**：
- `internal/frontend/dispatcher/calls.go:57`：`callbackMessage.Err = err.Error()`（默认拍平）
- 运行时 JS：`if (t.error) { let i = t.error instanceof Error ? t.error : new Error(t.error); o.reject(i) }`
- **官方公开 API**：`options.App.ErrorFormatter func(error) any`（`pkg/options/options.go:70`）
  —— 存在注入点，可覆盖错误的序列化形态。

**关键陷阱（Node 模拟实证）**：`ErrorFormatter` 返回 `any` **极具欺骗性**：

```
   返回【对象】     → e.message = "[object Object]"      ❌ 数据全丢
   返回【JSON 串】  → e.message = '{"code":"not_found"}' ✅ 前端 JSON.parse 可还原
   返回【纯字符串】 → e.message = "记录不存在"           （默认行为）
```

原因：前端用 `new Error(payload)` 包装，传对象会被强转成 `"[object Object]"`。

**决策**：错误协议采用 **`ErrorFormatter` + 返回 JSON 字符串 + 前端 `JSON.parse`**
（即 apply 中讨论的方案「甲」，探针修正后落地）。**不得返回对象**——签名允许但运行时会丢数据。

## Decisions

### D1: 错误协议 —— 结构化，经 ErrorFormatter 序列化为 JSON 字符串

**Go 侧**：统一错误类型，带机器可读 `code` 与面向用户的中文 `message`。

```go
// internal/apperr（包名待定）
type Error struct {
    Code    string `json:"code"`             // "not_found" / "validation" / "internal" ...
    Message string `json:"message"`          // 中文，可直接展示
    Detail  string `json:"detail,omitempty"` // 可选，仅排查用，不展示
}
func (e *Error) Error() string { return e.Message }

func New(code, message string) *Error
func NotFound(msg string) *Error      // code = not_found
func Validation(msg string) *Error    // code = validation
func Wrap(err error) *Error           // 未知 → code = internal，message 兜底中文
```

**序列化（关键）**：Go 侧经 `ErrorFormatter` 把错误**编码为 JSON 字符串**（**不是对象**，见 U2 陷阱）：

```go
// internal/apperr —— 供 main.go 的 options.App.ErrorFormatter 使用
func Format(err error) any {          // 返回 any，但【必须】落在 string 上
    var ae *Error
    if errors.As(err, &ae) {
        b, _ := json.Marshal(ae)
        return string(b)              // ✅ JSON 串；返回 ae 会变成 "[object Object]"
    }
    fallback := &Error{Code: CodeInternal, Message: "系统错误", Detail: err.Error()}
    b, _ := json.Marshal(fallback)
    return string(b)
}
```

```go
// main.go
wails.Run(&options.App{ ..., ErrorFormatter: apperr.Format })
```

**前端侧**：一个 `invoke` 包装（**不是 UI 组件**），把 `Error.message` 里的 JSON 解析回结构化。

```ts
// frontend/src/lib/invoke.ts
export interface AppError { code: string; message: string; detail?: string }

export async function invoke<T>(fn: () => Promise<T>): Promise<T> {
  try { return await fn() }
  catch (raw) {
    throw normalizeError(raw)   // → AppError { code, message, detail }
  }
}

// 容错：解析失败（如非本协议的错误）退化为 internal + 原文
function normalizeError(raw: unknown): AppError {
  const text = raw instanceof Error ? raw.message : String(raw)
  try {
    const p = JSON.parse(text)
    if (p && typeof p.code === 'string') return p as AppError
  } catch { /* 落到兜底 */ }
  return { code: 'internal', message: text || '未知错误' }
}
```

**为什么是「结构化 + 包装」**：结构化保证 `code` 可靠（下游能按 code 分流）；
包装保证**归一化的容错**——即便某处漏了 ErrorFormatter，前端也不会炸成
`"[object Object]"`，而是退化为 `internal` + 原文。包装是可选便利层，不强迫下游使用。

**必须注意**：绑定方法**不能**把 `*apperr.Error` 放进返回签名（如 `(T, *apperr.Error)`）。
那会生成 `Promise<T|apperr.Error>` 的**误导联合类型**，且与运行时不符（见 U2 陷阱）。
错误一律走 `error`，靠 `ErrorFormatter` 决定形态。

### D2: 分页协议 —— 结构化 PageRequest / PageResult

```go
type PageRequest struct {
    Page     int `json:"page"`      // 1-based
    PageSize int `json:"page_size"` // 0 → 默认 DefaultPageSize；>MaxPageSize → 截断
}

type PageResult[T any] struct {     // ← 泛型，U1 已确认支持
    List     []T   `json:"list"`
    Total    int64 `json:"total"`
    Page     int   `json:"page"`
    PageSize int   `json:"page_size"`
}
```

**已知代价（U1 实测）**：生成的 TS 类名会被压平成
`PageResult_<pkg路径>_<类型参数>_`（如 `PageResult_pcsbx_internal_probe_Item_`）。
类型安全，名字难看，但只在类型标注处可见，接受。

**约束**：`DefaultPageSize` / `MaxPageSize` 是常量**单一真相**，service 层与护栏共享。

**`PageRequest` 归一化方法**（防负数/超限）放结构体上，供 service 统一调用：

```go
func (r PageRequest) Normalized() PageRequest  // page<1 → 1；size 越界 → 夹取
```

### D3: 事件协议 —— 约定式（命名 + payload，不做结构强制）

```
命名：<domain>:<action>        如  "import:progress" / "import:done" / "import:error"
payload：纯 JSON 对象
   - 进度类 MUST 带  done / total（数字）
   - 结束类 MUST 带  ok（布尔），失败时带 message
   - 事件名中的 domain 与业务模块名一致（对齐 model/service 命名）
```

**为何不做结构强制**：事件本质灵活（不同业务 payload 差异大），硬套结构会变成教条，
违背「管道不该越界到产品面」。用**命名约定 + 范例**引导即可。

`_example` 里给一个「长任务推进度」的最小范例（Go 侧 `EventsEmit` + 前端 `EventsOn`），
作为可抄的写法——**这是范例，不是产品面组件**。

### D4: 契约护栏 —— 只断言可判定的部分

对齐 AGENTS.md 铁律 5（护栏须感知自己瞎了）。**可判定**的：

```
① 分页结构体存在且字段齐全（List/Total/Page/PageSize）
② DefaultPageSize / MaxPageSize 常量存在且夹取逻辑有测试
③ 生成器产出的列表方法返回分页结果（gen.sh 模板断言）
④ 挂载错误协议的绑定方法其 error 经 apperr 构造（AST 断言构造器调用）
```

**不可判定、不设护栏**的：事件 payload 语义、业务错误的合理性——靠 review。

### D5: 规范文档纠偏

`docs/代码规范.md` 删两句遗骸：
- 「由 controller 转成响应」→ 本仓无 controller，绑定层是 `app.go`。
- 「API 定义集中在 `api/index.js`」→ 无此文件，前端直接调 wailsjs 绑定。

替换为本仓真实的契约描述（指向本 change 落地的协议）。

## Risks / Trade-offs

- **[Risk] U1 泛型不被 Wails 支持** → 探针先行；降级方案 B/C 已备（见上）。
- **[Risk] U2 结论为「字符串」** → 错误协议需退化为「编码在字符串里」，前端包装 JSON 解析。
  这会略微牺牲可读性，但包装层把影响挡在下游之外。
- **[Risk] 破坏性变更**（绑定签名从裸 slice → 分页）→ 本仓无真实业务模块，现在改代价最小。
- **[Risk] 契约过度设计** → 严格守住「只定管道，不定产品面」；事件协议刻意保持约定式。
- **[Risk] 契约定了没人用** → 靠 `_example`（黄金范例）演示 + `gen.sh` 默认产出，
  让「正确用法」成为「最省事的用法」。

## 落地顺序

```
  0. 技术探针 U1 / U2（未定不得动手）   ← 必须先做
  1. 错误协议（apperr + invoke 包装）+ 护栏
  2. 分页协议（PageRequest/PageResult） + 护栏
  3. 事件协议（命名约定 + _example 范例）
  4. gen.sh / _example 据契约重写
  5. 规范文档纠偏
```
