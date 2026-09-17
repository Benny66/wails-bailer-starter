# ci-pipeline

CI：提交即检查与编译矩阵。

## RENAMED Requirements

- FROM: `### Requirement: 三平台编译`
- TO: `### Requirement: 编译矩阵`

## MODIFIED Requirements

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
