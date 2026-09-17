# ci-embed-dist — 设计

## D1：为什么必须在 check job 里构建前端

`go test` / `go vet` 都要编译根包（`package main`），而根包顶部是：

```go
//go:embed all:frontend/dist
var assets embed.FS
```

`//go:embed` 的目录**必须存在**，否则是编译错误（Go 不支持可选 embed）。
而 `frontend/dist` 在 `.gitignore` 里（构建产物不入库），干净检出必然没有它。

所以「跑 Go 测试」这件事对前端产物有**硬依赖**。CI 里让这个依赖成立的最直接做法，
就是在测试前把前端构建一次。

**为何不把 dist 入库**：它每次构建都变，入库等于把产物当源码，直接违反干净性铁律，
而且会让 diff 全是噪音。

**为何不让根包对 dist 可选**：`//go:embed` 没有可选语法。绕开它要么改成运行时读文件
（丢掉单二进制分发的好处），要么引入 build tag 分叉（多一套要维护的编译形态）。
代价都远大于「在 CI 里多跑一条 `npm run build`」。

**为何放在 check job 而不是抽成独立 job**：check 与 build 都要 dist，
`build` job 的 `make build` 本来就会构建它；check job 补一条即可，不新增 job 编排。

## D2：本地也要拦住，而且要拦住得可读

同一个坑在本地是：新 clone → `make test` → 一句
`main.go:24:12: pattern all:frontend/dist: no matching files found`。
它没提「前端没构建」，也没说该怎么办。

故 `make test` / `make lint` 增加前置目标 `require-dist`：

```
错误：frontend/dist 不存在。
      根包用 //go:embed 嵌入前端产物，编译它需要先构建前端。
      修复：cd frontend && npm install && npm run build
```

**为何不让 `make test` 自动构建**：隐式副作用（跑测试顺手改了工作区）+ 每次多花十几秒。
本仓的既有做法是「未满足的条件显式失败并指路」（见 constitution 的「未实现能力显式失败」）。

## D3：`GO_PKGS` 必须惰性求值（否则提示会被自己的噪音淹没）

```makefile
GO_PKGS := $(shell go list ./... | grep -v node_modules)   # 立即展开
GO_PKGS =  $(shell go list ./... | grep -v node_modules)   # 递归展开（采用）
```

`$(shell ...)` 在**立即展开**（`:=`）下于 make 解析期执行——`go list` 会编译根包，
于是在缺少 dist 的仓库里，**任何**目标（包括 `make help`）都会先喷一句
`pattern all:frontend/dist: no matching files found`。

改成递归展开（`=`）后，它只在引用它的 recipe 里求值，而那时 `require-dist` 已通过。
实测确认：`make help` 干净，`make test` 只输出那一条可读错误。

**注意不要用「吞掉 stderr」来解决**（如 `go list ./... 2>/dev/null`）：
那会让 `go list` 因其它原因失败时静默产出空包列表，`go test` 退化成只测当前目录
而不报错——正是本仓铁律所禁的「静默通过」。

## D4：验证方式

| 场景 | 验证手段 |
|---|---|
| 新 clone 的 `make test` | 移走 `frontend/dist` 后跑，确认只输出可读错误 |
| `make help` 无噪音 | 同上条件下跑 |
| CI 的 check 序列 | 本地按新步骤复跑：`npm run build` → `make test` → `make lint` |
| Linux 专有差异 | Docker `golang:1.25`（与 CI 同版本）跑 `go test` / `go vet` / `gofmt` |
| 历史归因 | 用 GitHub API 对比 `7bbeffa`（改动前）的失败步骤 |

CI 上的真实执行需 push 后由 GitHub 触发，本地无法覆盖。
