# ci-embed-dist — 修好长期必红的 CI（根包嵌入前端产物，干净检出里没有它）

## Why

CI 的 `check` job 在「Run tests」步骤**每次都红**，且早于本轮任何改动
（`7bbeffa`、`56b24a2`、`c242a18` 三次运行的失败步骤完全一致）。

**根因**：根包 `main.go` 用 `//go:embed all:frontend/dist` 嵌入前端产物，
而 `frontend/dist` 被 `.gitignore` 排除——**干净检出里根本不存在**。
`check` job 只跑 `npm install`、不构建前端，于是编译根包即失败：

```
main.go:24:12: pattern all:frontend/dist: no matching files found
# __APP_NAME__
FAIL	__APP_NAME__ [setup failed]
```

报错里唯一显眼的是模块名（`__APP_NAME__`），看起来像「项目名占位符的问题」，
实际与占位符无关——这也是排查时最容易走偏的地方。同一个坑在本地也存在：
新 clone 仓库后直接 `make test` 会得到同一句难懂的编译错误。

`build` job 不受影响（`make build` 会先构建前端），所以问题一直只显现在 `check`。

## What Changes

- **CI 的 `check` job 补上前端构建**：`go test` / `go vet` 需要根包可编译，
  而根包嵌入了前端产物，故必须先 `npm run build`。
- **`make test` / `make lint` 增加前置检查**（`require-dist`）：缺失时给出
  「根包用 go:embed 嵌入前端产物 + 修复命令」的可操作提示，而不是抛 Go 编译错误。
- **`GO_PKGS` 改为惰性求值**：它由 `go list ./...` 求值，而 `go list` 会编译根包。
  立即求值会让**任何** make 目标（含 `make help`）在新 clone 上先喷一句编译错误。

## Capabilities

### Modified Capabilities

- `ci-pipeline`: 检查 job 必须自备编译前置条件，否则检查本身跑不起来。
- `project-scaffold`: 统一命令入口在缺少前置条件时显式失败，并给出可操作的修复命令。

## Non-Goals

- **不把 `frontend/dist` 入库**（那是构建产物，入库违反干净性铁律）。
- **不让根包对 dist 变成可选**（`//go:embed` 不支持可选目录；绕过它需要改构建结构，
  代价远大于把前置条件讲清楚）。
- **不让 `make test` 自动构建前端**（隐式副作用 + 每次多花十几秒；
  显式失败并给出命令更符合本仓「验证入口统一」的做法）。

## Impact

- **修改文件**：`.github/workflows/build.yml`（check job 加一步）、`Makefile`
  （`require-dist` 目标 + `GO_PKGS` 惰性求值）、`README.md`（快速开始补前置说明）。
- **破坏性**：**无**。仅让原本就失败的两条路径（CI check、新 clone 的 make test）
  给出正确结果或可读错误。
