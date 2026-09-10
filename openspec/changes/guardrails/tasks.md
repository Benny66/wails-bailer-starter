# guardrails — 实施任务

## 1. Go 护栏框架

- [x] 1.1 建 `internal/guard/` 包与测试入口，用 `go/parser + go/ast` 解析源码
- [x] 1.2 实现"感知自己瞎了"：解析到 0 结果即 Fatal

## 2. Go 护栏规则

- [x] 2.1 分层越界：绑定方法文件不得 import gorm/database；model 不含业务逻辑
- [x] 2.2 模型注册双向校验：`AllModels()` 与带 BaseModel 结构体集合互相断言
- [x] 2.3 绑定方法名与 `frontend/wailsjs` 生成文件一致

## 3. ESLint 护栏

- [x] 3.1 建 `frontend/eslint.config.js`（flat config）
- [x] 3.2 自定义规则 `no-node-imports`：禁 import node:*/fs/child_process
- [x] 3.3 自定义规则 `no-hardcoded-brand`：禁硬编码 hex 色值与品牌串

## 4. 依赖登记制

- [x] 4.1 建 `deps.yaml`，登记现有 Go 直接依赖 + 前端 dependencies（附理由）
- [x] 4.2 写 `internal/guard/deps_test.go` 双向校验（go.mod + package.json）

## 5. 接线与同步

- [x] 5.1 `make lint` 改为真实执行（ESLint + go vet + 护栏测试）
- [x] 5.2 `AGENTS.md` 中 ⚙️ 规则指向真实护栏

## 6. 验证

- [x] 6.1 `make test` 护栏全绿
- [x] 6.2 `make lint` 通过
- [x] 6.3 `openspec validate guardrails` 通过
