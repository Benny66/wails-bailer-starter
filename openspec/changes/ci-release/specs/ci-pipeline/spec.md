# ci-pipeline

GitHub Actions 工作流——静态检查 + 三平台编译 + artifact 上传。

## ADDED Requirements

### Requirement: 提交即检查
CI MUST 在提交后自动运行静态检查（Go 测试 + 护栏 + lint）。

#### Scenario: 静态检查失败即红
- **WHEN** 提交引入了护栏或 lint 违规
- **THEN** CI 失败，问题在合并前暴露

### Requirement: 三平台编译
CI MUST 编译 Windows/macOS/Linux 三平台产物并上传 artifact。

#### Scenario: 三平台产物
- **WHEN** CI 运行编译 job
- **THEN** 产出 exe（Windows）/ app（macOS）/ 二进制或 AppImage（Linux），并上传为 artifact
