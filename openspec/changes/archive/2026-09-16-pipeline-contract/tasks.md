# pipeline-contract — 实施任务

## 0. 技术探针（未定不得动契约形状）

- [x] 0.1 **U1 探针（2026-09-14，Wails v2.15.0，沙箱 `wails build`）—— 泛型✅支持**。
      写 `func (a *App) ProbePage() (probe.PageResult[probe.Item], error)`，绑定生成成功：
      - `App.d.ts`：`ProbePage(): Promise<probe.PageResult_pcsbx_internal_probe_Item_>`
      - `models.ts`：生成类 `PageResult_pcsbx_internal_probe_Item_`，字段完整（list/total/page/page_size）
      - **结论**：Wails 把泛型实例**压平成含模块路径+类型参数的类名**。
        类型安全（每模块专属类，`list: Item[]` 正确），但**名字难看且冗长**。
      - **决策**：**采用泛型**（方案 A），接受类名难看。理由：下游拿到的是类型安全的
        `list: Asset[]`，类名丑只在 TS 类型标注处可见、不泄漏给下游业务代码；
        方案 C（每模块派生类型）由泛型**自动达成**，无需 gen.sh 额外派生。
- [x] 0.2 **U2 探针（Wails v2.15.0 源码 + 运行时 JS + Node 模拟）—— 默认拍平，但有官方钩子**。
      证据链：
      - `internal/frontend/dispatcher/calls.go:57`：`callbackMessage.Err = err.Error()`（默认拍平）
      - 运行时 JS：`if (t.error) { let i = t.error instanceof Error ? t.error : new Error(t.error); o.reject(i) }`
      - **官方公开 API**：`options.App.ErrorFormatter func(error) any`（`pkg/options/options.go:70`）
      - **Node 模拟（决定形态）**：
        `返回对象 → e.message = "[object Object]"`（❌ 丢数据）
        `返回 JSON 串 → e.message = '{"code":...}'`（✅ 可 parse）
      - **决策（已定稿）**：错误协议走 **`ErrorFormatter` 返回 JSON 字符串 + 前端 `JSON.parse`**。
        **不得返回对象**（签名允许，运行时会退化成 `[object Object]`）。
        绑定方法签名**不得**把 `*apperr.Error` 放进返回值（会产生误导的联合类型 `Promise<T|AppErr>`）。

## 1. 错误协议

- [x] 1.1 新增 `internal/apperr`：`Error{Code,Message,Detail}` + `New/NotFound/Validation/Conflict/Wrap`
      + `Format`（返回 **JSON 字符串**，供 ErrorFormatter）+ 4 个单测（含「MUST 是 string 不是 object」）
- [x] 1.2 前端归一化包装 `frontend/src/lib/invoke.ts`：`normalizeError` + `invoke`，
      容错退化为 `internal` + 原文，保证下游永不撞见 `[object Object]`
- [x] 1.3 `main.go` 注入 `ErrorFormatter: apperr.Format`；`app.go` 的 `SetTheme` 据协议改造
      （`apperr.Validation` / `apperr.Wrap`）
- [x] 1.4 护栏 `TestBindingsUseAppErr`：绑定层禁用 `errors.New`/`fmt.Errorf`
      （已破坏性验证会红）

## 2. 分页协议

- [x] 2.1 新增 `internal/page`：`Request`（含 `Normalized()`/`Offset()`）+ 泛型 `Result[T]` + `NewResult`
- [x] 2.2 常量 `DefaultPageSize=20` / `MaxPageSize=200` 单一真相 + 夹取/offset/回填/nil 列表 4 组单测
- [x] 2.3 护栏 `TestPageContractShape` + `TestPageSizeConstantsExist`（均已破坏性验证会红）

## 3. 事件协议

- [x] 3.1 新增 `internal/event`：动作常量 + `Name(domain,action)` + `Progress{Done,Total}` /
      `Result{Ok,Message}` payload 约定；约定写入 `docs/代码规范.md`
- [x] 3.2 事件协议在分包注释与规范文档中给出可抄的最小范例
      （注：为保持 `_example` 与分页/错误契约聚焦，未在其中塞入长任务范例）

## 4. 生成器与范例据契约重写

- [x] 4.1 `_example/service/example_service.go`：列表返回 `page.Result[T]`、错误经 `apperr` 构造
- [x] 4.2 `_example/bindings/app_example.go.txt`：绑定方法遵循错误协议（返回分页结果）
- [x] 4.3 `_example/frontend/ExampleList.vue`：据分页结果 + `invoke` 包装改写
- [x] 4.4 `scripts/gen.sh`：**修复两个真实 bug**（见下）+ `page` import 幂等注入
      - Bug A：绑定片段用了 `page.Request` 但只注入 `model` import → 生成物编译失败
      - Bug B（**既有**）：import 注入非幂等 → 生成第 2 个模块即 `redeclared` 编译失败
      - Bug C（**既有**）：`deps_test` 的 go.mod 正则会把 `module <v开头的名字>` 误当依赖
        → 任何以 v 开头的项目名被凭空报出未登记的 "module" 依赖
- [x] 4.5 新增 `scripts/verify-gen.sh` + `make verify-gen`：临时沙箱连续生成 2 个模块 →
      生成 bindings → 编译 → 护栏。**已在母版上真跑通过**

## 5. 文档纠偏

- [x] 5.1 `docs/代码规范.md` 删遗骸（controller / `api/index.js`），替换为管道契约三节
- [x] 5.2 `docs/脚手架功能说明.md` 重写：删未实现承诺（自动更新/WebView2/ps1/表格示例），
      补「只给管道」定位与真实能力边界
- [x] 5.3 `docs/map.md` 补全全部已落地路径（service/config/logging/dialog/tray/crash/
      apperr/page/event/guard + lib/invoke.ts + scripts + build）

## 6. 验证

- [x] 6.1 契约护栏全部通过；三条新护栏 + deps 护栏修复均做了**破坏性验证（确认会红）**
- [x] 6.2 `make verify-gen` 端到端跑通（母版上实测，见 4.5）
- [x] 6.3 `make test` / `make lint` 通过（含 gofmt + 护栏 + go vet + 前端 ESLint）
- [x] 6.4 `openspec validate pipeline-contract --strict` 通过

