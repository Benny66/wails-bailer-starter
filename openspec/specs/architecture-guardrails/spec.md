# architecture-guardrails Specification

## Purpose
TBD - created by archiving change guardrails. Update Purpose after archive.
## Requirements
### Requirement: 护栏感知自己瞎了
任何用 AST/解析源码做断言的护栏，MUST 在解析结果为 0 时报错，而非当作"通过"静默放行。

#### Scenario: 解析失败即报错
- **WHEN** 护栏解析目标源码得到 0 个结果（如写法变更导致匹配失效）
- **THEN** 护栏报错并提示"写法可能已变更，请同步更新护栏解析规则"

### Requirement: 分层越界拦截
Go 侧分层 MUST 由护栏强制：绑定方法层不直接操作 gorm，model 层不含业务逻辑。

#### Scenario: 绑定层不碰 gorm
- **WHEN** 绑定方法所在文件 import 了 gorm 或 database 包
- **THEN** 护栏测试失败

### Requirement: 模型注册双向校验
带 BaseModel 的结构体 MUST 全部登记进 `AllModels()`，且登记项 MUST 真实存在。

#### Scenario: 漏登记
- **WHEN** 存在带 BaseModel 的结构体未出现在 `AllModels()` 中
- **THEN** 护栏测试失败

#### Scenario: 僵尸条目
- **WHEN** `AllModels()` 中登记的条目在 model 包中不存在对应结构体
- **THEN** 护栏测试失败

### Requirement: 前端 import 安全
前端渲染层 MUST 禁止直接 import Node 能力（`node:*`、`fs` 等），必须走 wails bindings。

#### Scenario: 禁用 node 模块
- **WHEN** 前端源码 import 了 `node:*` / `fs` / `child_process` 等
- **THEN** ESLint 检查失败

