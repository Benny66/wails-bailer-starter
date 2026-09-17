# ci-embed-dist — 实施任务

## 1. 定位根因

- [x] 1.1 用 GitHub API 取失败运行（`gh` 未登录；仓库公开，API 可读但 job 日志需管理员权限）
- [x] 1.2 对照历史：`7bbeffa` / `56b24a2` / `c242a18` 三次运行的失败步骤一致
      （`check` job 的「Run tests」）→ 确认是既有问题，非本轮引入
- [x] 1.3 本地复现：移走 `frontend/dist` 后 `make test` 得到同一句
      `pattern all:frontend/dist: no matching files found` + `FAIL __APP_NAME__ [setup failed]`
      → 报错里唯一的显眼标识是模块名，这就是「看起来像占位符问题」的来源

## 2. 修复 CI

- [x] 2.1 `check` job 在 `npm install` 与测试之间补 `npm run build`（附原因注释）
- [x] 2.2 确认 `build` job 不受影响（`wails build` 本就会构建前端）

## 3. 修复本地体验

- [x] 3.1 新增 `require-dist` 前置目标：缺 dist 时输出「原因 + 修复命令」并非零退出
- [x] 3.2 `test` / `lint` 依赖它；`.PHONY` 补登
- [x] 3.3 `GO_PKGS` 由立即展开改递归展开（否则 `make help` 也会先喷编译错误）
- [x] 3.4 刻意不用 `2>/dev/null` 吞错（会退化成「跳过检查却显示通过」）

## 4. 验证

- [x] 4.1 缺 dist：`make help` 干净；`make test` / `make lint` 各只输出一条可读错误、退出 2
- [x] 4.2 按 CI 新序列本地复跑：`npm run build` → `make test`（0）→ `make lint`（0）
- [x] 4.3 Docker `golang:1.25`（与 CI 同版本）Linux 上 `go test` 全绿、
      `go vet` 与 `gofmt` 干净（排除平台差异这一变量）
- [ ] 4.4 未实跑：CI 上的真实执行——需 push 后由 GitHub 触发

## 5. 文档与归档

- [x] 5.1 `README.md` 快速开始补「先构建前端再跑测试」的前置说明
- [x] 5.2 `openspec validate ci-embed-dist --strict` 通过
- [ ] 5.3 归档
