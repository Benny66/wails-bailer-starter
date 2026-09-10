# example-module — 实施任务

## 1. 黄金范例 _example

- [x] 1.1 建 `_example/` 模板四段：model + service + 绑定方法 + 前端页面，最小干净 + TODO 锚点
- [x] 1.2 确保 `_example/` 无任何业务残留（无具体字段名/业务逻辑）

## 2. 生成器 make gen

- [x] 2.1 建 `scripts/gen.sh`：fail-fast 前置校验（锚点/占用/幂等）
- [x] 2.2 锚点注入：`AllModels()`、app.go 绑定 + import、路由/菜单
- [x] 2.3 占位符替换（Example→Tool）+ `// TODO` 锚点
- [x] 2.4 `make gen` 从占位改为真实生成器

## 3. 护栏同步

- [x] 3.1 护栏断言生成锚点存在（`gen_anchors_test.go`）

## 4. 验证

- [x] 4.1 `make gen name=test` 生成成功，go build + make build 编译通过，git 恢复后零残留
- [x] 4.2 开箱零业务代码：`internal/model` 仅注册表、`internal/service` 仅聚合、app.go 无示例绑定
- [x] 4.3 `openspec validate example-module` 通过
