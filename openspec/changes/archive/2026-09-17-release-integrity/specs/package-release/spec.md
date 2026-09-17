# package-release

打包与发布：参数化打包、安装包产物、版本与标识。

## ADDED Requirements

### Requirement: 版本号单一真相
版本号 MUST 只有一个出处（`wails.json` 的 `info.productVersion`），
所有产物与应用内取值 MUST 由它派生，且四处一致。

#### Scenario: 系统可见版本与配置一致
- **WHEN** 构建 macOS 产物
- **THEN** 产物 `Info.plist` 的 `CFBundleShortVersionString` 等于配置中的版本，
  而非平台或框架的默认值

#### Scenario: Windows 版本资源同源
- **WHEN** 交叉编译 Windows 产物
- **THEN** exe 版本资源中的版本等于同一出处

#### Scenario: 应用内版本同源
- **WHEN** 应用启动
- **THEN** 日志首行与 `GetAppInfo()` 返回的版本等于同一出处

#### Scenario: 版本注入点唯一
- **WHEN** 检查各构建脚本
- **THEN** 只有一处进行版本注入，其余构建入口都经它调用，
  不存在「各脚本各算一份版本」

#### Scenario: 构建后回读校验
- **WHEN** macOS 构建完成
- **THEN** 构建脚本回读产物 `Info.plist` 并断言版本一致，
  不一致时构建失败（而非等用户看「显示简介」才发现）

#### Scenario: 交叉编译不误读残留产物
- **WHEN** 在 macOS 上交叉编译其他平台的产物
- **THEN** 不做 macOS 产物回读校验（避免读到上一次构建残留的 `.app`）

### Requirement: 版本格式前置校验
版本号 MUST 在构建开始前校验为数字点分格式，非法值 MUST 立即失败。

#### Scenario: 非法版本早失败
- **WHEN** 版本号含非数字成分（如 `v1.2.3` 的前缀、git 短哈希）
- **THEN** 构建与发布脚本在动手前报错退出，并说明 Windows 打包会失败的原因

#### Scenario: 版本号缺失即失败
- **WHEN** 配置中缺少版本号
- **THEN** 构建脚本报错退出，不使用框架默认值静默继续

### Requirement: 提交号与版本号分开注入
提交标识 MUST 与版本号分开注入，使版本号保持数字格式的同时保留追溯能力。

#### Scenario: 构建期注入提交号
- **WHEN** 在 git 仓库内构建
- **THEN** 提交短哈希被注入并被 `GetAppInfo()` 返回，供排查定位

### Requirement: bundle id 由项目决定
应用的 bundle id MUST 由项目自定义，MUST NOT 沿用框架模板的默认厂商前缀。

#### Scenario: bundle id 使用项目前缀
- **WHEN** 构建 macOS 产物
- **THEN** `CFBundleIdentifier` 形如 `<项目前缀>.<应用名>`，而非框架默认前缀

#### Scenario: 前缀可配置且默认明显为占位
- **WHEN** 实例化新项目且未指定前缀
- **THEN** 使用 `com.example` 这类明显是占位符的前缀，并提示发布前替换
