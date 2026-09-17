# ci-pipeline Specification

## Purpose
TBD - created by archiving change ci-release. Update Purpose after archive.
## Requirements
### Requirement: 提交即检查
CI MUST 在提交后自动运行静态检查（Go 测试 + 护栏 + lint）。

#### Scenario: 静态检查失败即红
- **WHEN** 提交引入了护栏或 lint 违规
- **THEN** CI 失败，问题在合并前暴露

### Requirement: 生成器端到端验证纳入 CI
CI MUST 执行生成器端到端验证，使该检查不再只依赖开发者记得在本地运行。

#### Scenario: 生成器回归在 CI 变红
- **WHEN** 生成器产出的代码无法编译、不符格式或被护栏拦住
- **THEN** CI 失败，而不是等人工在本地跑验证脚本才发现

#### Scenario: 单腿执行即可
- **WHEN** 该检查与平台无关，且在部分平台上依赖不可用
- **THEN** 只在一条具备全部前置条件的腿（有构建工具链与所需系统命令）上执行一次，
  不重复三遍

### Requirement: 检查 job 自备编译前置条件
检查 job MUST 在执行编译类检查前满足其前置条件，不得假定检出目录中已存在构建产物。

#### Scenario: 嵌入产物的前置条件被满足
- **WHEN** 根包以嵌入方式依赖某个构建产物目录，而该目录不入库
- **THEN** 检查 job 在编译前先构建它，而不是直接失败

#### Scenario: 检查 job 真的能跑通
- **WHEN** 在干净检出上运行检查 job 的全部步骤
- **THEN** 每一步都通过，不出现「检查本身跑不起来」的长期红灯

### Requirement: 编译矩阵
CI MUST 编译并上传产物，且矩阵 MUST 只包含能稳定通过的平台。
编译结果取决于 runner 镜像装了什么系统包这类环境耦合，MUST NOT 作为矩阵成员长期维护。

#### Scenario: 矩阵内平台产物
- **WHEN** CI 运行编译 job
- **THEN** 产出 app（macOS）与 exe（Windows），并上传为 artifact

#### Scenario: 环境耦合的平台不进矩阵
- **WHEN** 某平台的编译结果取决于 runner 镜像装了什么系统包
- **THEN** 该平台不在编译矩阵内，其构建说明标注为「无 CI 验证」，
  而非以经常失败的形态留在矩阵中

#### Scenario: 不软化失败
- **WHEN** 一条腿无法稳定通过
- **THEN** 从矩阵移除，而不是用「允许失败」的方式让它红着不拦——
  后者会让失败长期无人处理

#### Scenario: Go 层覆盖不随之丢失
- **WHEN** 某平台因产物编译不稳定而退出矩阵
- **THEN** 它的 Go 层行为仍由静态检查腿覆盖（测试 / 护栏 / lint / 生成器端到端），
  只是不再承诺其产物可编译

