# example-module (delta)

## MODIFIED Requirements

### Requirement: 黄金范例
`_example/` MUST 作为可抄的黄金范例，演示符合管道契约的正确写法
（分页列表、错误处理、事件推进度），且 MUST 可作为生成器的可靠模板
（`make gen` 产出的链路可编译可跑）。

#### Scenario: 范例演示分页列表
- **WHEN** 阅读 `_example` 的列表实现
- **THEN** 其返回分页结果并归一化分页参数，而非返回裸切片

#### Scenario: 范例演示错误处理
- **WHEN** 阅读 `_example` 的服务/绑定实现
- **THEN** 业务错误经统一错误构造（如记录不存在返回 not_found），可直接照抄

#### Scenario: 范例演示事件推进度
- **WHEN** 阅读 `_example` 或其配套文档
- **THEN** 有一个「长任务 + 进度事件 + 前端监听」的最小范例，演示事件协议约定

#### Scenario: 生成器模板与范例同源
- **WHEN** 执行 `make gen`
- **THEN** 生成物源自 `_example`，二者契约一致（不出现范例正确但生成物过期的漂移）
