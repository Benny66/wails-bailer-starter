# ci-pipeline

CI：提交即检查、编译矩阵与发布。

## ADDED Requirements

### Requirement: 发版前先过静态检查
发布工作流 MUST 在构建安装包之前执行静态检查，MUST NOT 从检查不通过的提交发版。

#### Scenario: 红提交不发版
- **WHEN** tag 指向一个未通过测试或 lint 的提交
- **THEN** 发布流程在打包前失败，不产出也不创建 Release

#### Scenario: 检查通过才继续
- **WHEN** 静态检查通过
- **THEN** 进入各平台打包阶段

### Requirement: 发布产物在多平台间正确汇合
跨平台构建的产物 MUST 汇聚到同一次发布，且 MUST NOT 因并行而产生重复发布。

#### Scenario: 并行构建不重复创建
- **WHEN** 多个平台的构建并行完成
- **THEN** 发布只被创建一次，全部产物附加到同一个发布

#### Scenario: 打包失败不留空发布
- **WHEN** 某一平台的打包失败
- **THEN** 不创建（或不同步）缺少产物的发布
