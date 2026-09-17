# architecture-guardrails

Go AST 护栏 + ESLint 自定义规则，把分层与注册约束编译成会红的检查。

## ADDED Requirements

### Requirement: 可观测性接线不得静默缺失
「删掉后不会报错、只是失去可见性」的观测接线 MUST 由护栏核对。

#### Scenario: 数据库日志接线被核对
- **WHEN** 数据库连接未配置日志实现
- **THEN** 护栏失败并说明后果（慢查询回到无人可见的输出）

#### Scenario: 启动打点被核对
- **WHEN** 组合根的启动阶段打点被移除或数量异常减少
- **THEN** 护栏失败，说明后果（启动优化重新变成盲调）

#### Scenario: 解析不到即报错
- **WHEN** 护栏在目标文件中解析到 0 个打点
- **THEN** 报错并提示同步更新护栏解析规则，而非静默通过
