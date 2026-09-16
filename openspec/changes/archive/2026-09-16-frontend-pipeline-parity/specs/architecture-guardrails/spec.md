# architecture-guardrails

Go AST 护栏 + ESLint 自定义规则，把分层与注册约束编译成会红的检查。

## ADDED Requirements

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
