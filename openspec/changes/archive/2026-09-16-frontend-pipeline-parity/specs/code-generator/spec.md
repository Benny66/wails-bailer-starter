# code-generator

`make gen` 模块生成器：从黄金范例产出符合管道契约的 model + service + 绑定 + 页面。

## ADDED Requirements

### Requirement: 生成物格式立即合规
生成器 MUST 保证产出的代码直接通过项目静态检查，不得让用户生成完就去手工格式化。

#### Scenario: 生成后 lint 不红
- **WHEN** 执行 `make gen name=<module>`
- **THEN** 生成物立即通过 `make lint` 的 gofmt 检查（导入顺序等由生成器负责排好）
- **AND** 端到端验证脚本对此有断言，防止回归

### Requirement: 生成的页面不复制分页样板
生成的前端列表页 MUST 复用基座的分页加载辅助，而非在每份页面里重抄分页状态与加载流程。

#### Scenario: 生成页使用基座辅助
- **WHEN** 查看 `make gen` 产出的列表页
- **THEN** 它用 `usePagedList` 承载 loading/error/页码/总数，页面代码只保留业务表达
