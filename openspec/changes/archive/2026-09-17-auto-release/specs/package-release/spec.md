# package-release

打包与发布：参数化打包、安装包产物、版本与标识。

## ADDED Requirements

### Requirement: 打 tag 自动发布
推送版本 tag MUST 自动产出 GitHub Release，含安装包与自动生成的变更说明。

#### Scenario: 发布自动完成
- **WHEN** 推送形如 `v0.1.0` 的 tag
- **THEN** 自动构建各平台安装包，并创建 Release（说明由平台按提交自动生成）

#### Scenario: 产物可识别
- **WHEN** 查看 Release 的资产
- **THEN** 资产名含应用名、版本与平台，用户无需打开即可分辨

#### Scenario: 单平台上无法产出的产物由流水线补齐
- **WHEN** 某平台的安装包只能在对应操作系统上构建
- **THEN** 由该平台的构建环境产出并汇入同一次发布，不要求操作者具备所有平台

### Requirement: 版本号写入必须单一实现
把版本号写进配置的代码 MUST 只有一处实现，本地发布与 CI 发布 MUST 共用它。

#### Scenario: 两处共用
- **WHEN** 检查本地发布脚本与 CI 发布工作流
- **THEN** 二者都调用同一个写入脚本，各自不重复实现校验与写入

#### Scenario: 非法版本早失败
- **WHEN** 传入非数字点分格式的版本（如带前缀的 tag）
- **THEN** 在写入前即失败并说明原因，而非等到某平台打包的最后一步才炸

## MODIFIED Requirements

### Requirement: 母版防呆
母版（模板）MUST 拒绝被打包成可分发产物，且其发布流程 MUST NOT 因此失败。

#### Scenario: 母版跳过打包但不阻断发布
- **WHEN** 在含占位符的母版仓库上推送 tag
- **THEN** 跳过安装包构建，Release 仍照常创建（只含平台自动附加的源码归档）

#### Scenario: 占位符检测不被自身替换破坏
- **WHEN** 工作流文件中的占位符检测串经过实例化处理
- **THEN** 检测结果仍然正确（母版命中、实例化项目不命中）
