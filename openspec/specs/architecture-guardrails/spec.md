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

### Requirement: 前端镜像一致性护栏
前端复制的 Go 常量镜像（错误码、事件动作、页大小上下界）MUST 与 Go 单一真相保持一致，
由护栏做双向校验。

#### Scenario: 取值漂移即红
- **WHEN** Go 侧某常量取值变更而前端镜像未同步
- **THEN** `make test` 失败，报错指出镜像文件、键名与两侧取值

#### Scenario: 镜像缺项即红
- **WHEN** Go 侧新增了需镜像的常量而前端未补
- **THEN** `make test` 失败，报错指出应在哪个文件补哪一项

#### Scenario: 镜像多出项即红
- **WHEN** 前端镜像里存在 Go 侧已无对应常量的条目
- **THEN** `make test` 失败，指出该镜像已成第二处真相

#### Scenario: 解析不到即报错
- **WHEN** 护栏在 Go 包或前端镜像文件中解析到 0 个常量（写法变更导致匹配失效）
- **THEN** 护栏报错并提示"写法可能已变更，请同步更新护栏解析规则"，
  而非静默通过

### Requirement: TS 类型检查纳入静态验证入口
`make lint` MUST 包含 TypeScript 类型检查。

#### Scenario: 类型错误在 lint 阶段暴露
- **WHEN** 前端代码存在类型错误
- **THEN** `make lint` 失败，无需等到打包阶段

