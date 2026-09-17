# ci-pipeline

CI：提交即检查与三平台编译。

## ADDED Requirements

### Requirement: 检查 job 自备编译前置条件
检查 job MUST 在执行编译类检查前满足其前置条件，不得假定检出目录中已存在构建产物。

#### Scenario: 嵌入产物的前置条件被满足
- **WHEN** 根包以嵌入方式依赖某个构建产物目录，而该目录不入库
- **THEN** 检查 job 在编译前先构建它，而不是直接失败

#### Scenario: 检查 job 真的能跑通
- **WHEN** 在干净检出上运行检查 job 的全部步骤
- **THEN** 每一步都通过，不出现「检查本身跑不起来」的长期红灯
