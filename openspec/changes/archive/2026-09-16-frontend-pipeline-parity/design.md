# frontend-pipeline-parity — 设计

## D1：镜像常量，而不是运行时从 Go 拉

**问题**：页大小上下界、错误码、事件动作这三类常量，前端要用。两种做法——
(A) 前端写一份镜像，靠护栏保证不漂移；(B) 前端启动时经绑定从 Go 拉一次。

**决策：A（镜像 + 护栏）**。

理由：
- 这三类常量都是**编译期就要用**的东西（`PageResult` 是类型、错误码用于 `switch` 分支），
  运行时拉取会把「静态事实」变成「异步依赖」——首屏渲染前拿不到就无从判断，还得处理
  「拉失败怎么办」。为省一次复制引入一个加载态是亏的。
- 镜像本身不丢人，**无人看管的镜像才丢人**。护栏把「二者必须一致」编译成会红的检查后，
  镜像的成本被压到「改 Go 常量时顺手改一行 TS，否则红」。
- 已有先例：`invoke.ts` 的 `ErrorCode` 就是镜像（注释自认），本次只是把它纳入强制。

**代价**：新增常量要走两处。护栏的报错信息里直接写明「请同步 `frontend/src/lib/*.ts`」，
把成本显性化并指路。

**边界：只镜像常量，不镜像逻辑**。`Request.Normalized()`（页码下界、页大小夹取）**不**在
前端照抄一份——一件规则有两个实现，比一个常量有两份更糟：两者会在边界条件上悄悄分叉，
而谁都不知道该信哪个。归一化仍唯一地活在 Go 侧，前端发原始值、以回填值为准。

## D2：护栏如何解析 TS

**问题**：`internal/guard` 现有护栏都用 `go/ast` 解析 Go。前端镜像是 TS，没有现成 AST。

**决策：受约束的正则提取 + 锚点纪律**。

- 不引第三方 TS 解析器（会新增依赖 → 触发依赖登记制，代价大于收益）。
- 约定前端镜像文件里，常量 MUST 写在**具名导出对象/常量**中，形如
  `export const PAGE_DEFAULT_SIZE = 20` 与 `ErrorCode: { NotFound: 'not_found', … }`。
  正则只吃这种形状。
- **护栏感知自己瞎了**（铁律 5）：任一侧解析到 0 个常量即 `Fatal`，并提示
  「写法可能已变更，请同步更新护栏解析规则」。因此正则脆弱性不会静默通过——
  它只会红，不会漏。
- 断言是**双向**的：Go 有而 TS 没有 → 报缺；TS 有多余项 → 报多。防止镜像变成超集后
  被误当真相使用。

## D3：`internal/appdir` 保持既有落盘路径不变

四处的路径算法完全一致（`os.UserConfigDir()/<appName>/`）。抽离**只收敛出处，不改路径**：
数据库仍是 `<dir>/<app>/<app>.db`、日志仍是 `<dir>/<app>/app.log`、崩溃日志仍在同目录。
因此**已装用户的存量数据无需迁移**，本次变更为纯重构 + 新增读取能力。

`appdir` 只暴露两个函数，足够覆盖四处用途：

| 函数 | 用途 |
|---|---|
| `Dir(appName)` | 数据目录（确保存在） |
| `File(appName, name)` | 目录内文件路径（config.json / app.log / <app>.db / crash-*.log） |

崩溃日志带时间戳的文件名由 `crash` 包自己拼（`File` 不掺和时间语义）。

> 刻意不做成全局单例/包级变量：`appName` 是组合根传入的参数，包级状态会让
> `verify-gen` 这类「同一进程跑两份配置」的场景变脆。

## D4：打开目录用 `os/exec`，不用 `BrowserOpenURL`

Wails v2 运行时**没有**「在文件管理器中显示」的 API（只有 `BrowserOpenURL`）。

- 若走 `BrowserOpenURL("file://" + dir)`：路径含空格（macOS 的 `Application Support`、
  Windows 的 `AppData\Roaming` 下带空格用户名）必须 URL 编码，漏编码即静默失败——
  是个「在别人机器上才炸」的坑。
- **决策**：`internal/reveal` 用 `os/exec` 直接调用各平台的打开命令，路径作为**独立参数**
  传入（不经 shell），空格天然安全：
  - macOS：`open <dir>`
  - Windows：`explorer <dir>`
  - Linux：`xdg-open <dir>`
- 未知平台返回明确错误（不静默）；命令不存在（如精简 Linux 无 `xdg-open`）时错误经
  `apperr.Wrap` 归一化，前端据 `internal` 分流。失败不阻断——打开目录本就是尽力而为。

> 命令用 `exec.Command(name, args...)` 而非 `sh -c`：避免路径中的特殊字符被 shell 解释。

## D5：`usePagedList` 注入 fetcher，不直接 import bindings

```ts
const { list, total, page, pageSize, loading, error, load } = usePagedList(
  (req) => ListExamples(req),   // 注入：签名 (PageRequest) => Promise<PageResult<T>>
)
```

**决策：注入而非内联绑定调用**。理由：
- 组合式函数保持零业务耦合，任何模块的列表页都能用，无需改基座；
- 换实现（如加缓存、加轮询）只改调用方一行；
- 未来若引入前端测试，可注入 fake fetcher 而不必 mock 整个 wails 运行时。

**契约细节**：`load()` 内部经 `invoke()` 归一化错误 → `error` 恒为 `AppError | null`；
成功后**以 Go 回显的 `page`/`page_size` 为准**回写本地状态（这正是分页协议里
`NewResult` 回填的用意，手写页面最容易漏掉这一条）。

## D6：事件订阅的自动解绑

`onEvent(name, handler)` 返回取消函数（与 Wails `EventsOn` 同形，便于逐层替换）；
`useEvent` 组合式函数在 `onUnmounted` 自动调用它。

**为何必须有 `useEvent`**：`onEvent` 只把「取消函数」递到手上，漏不漏是人的纪律；
组件里漏解绑的后果是路由切回时同一事件触发两次、三次……而这类 bug 极难在开发态察觉
（开发时通常只切一次页面）。挂到组件生命周期上，才让「正确用法」成为「最省事的用法」。

## D7：`make lint` 补类型检查的代价

`vue-tsc --noEmit` 在冷启动下约数秒、增量下亚秒级。相对「类型错误漂到打包才炸」，
这个代价可以接受，且与 `npm run build`（已含 `vue-tsc`）同源，不引入第二套工具。
