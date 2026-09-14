# code-generator (delta)

## MODIFIED Requirements

### Requirement: 生成模块骨架
`make gen name=<module>` MUST 生成一条符合管道契约的完整链路
（model + service + 绑定方法 + 前端页面 + 路由 + 菜单），且产出的列表方法 MUST 返回分页结果。

#### Scenario: 生成的列表方法返回分页
- **WHEN** 执行 `make gen name=asset`
- **THEN** 生成的 service 列表方法与绑定方法返回分页结果（含 list/total），
  而非裸切片

#### Scenario: 生成的绑定方法遵循错误协议
- **WHEN** 查看生成器产出的绑定方法片段
- **THEN** 其错误返回经统一错误构造，前端可据 code 分流

#### Scenario: 生成物可编译可跑
- **WHEN** 执行 `make gen name=asset` 后编译
- **THEN** 生成物通过编译与护栏检查（生成器端到端正确性被验证，不留未跑通的模板）
