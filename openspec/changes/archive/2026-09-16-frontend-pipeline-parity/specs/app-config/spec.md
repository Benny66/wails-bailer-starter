# app-config

应用本地配置与数据目录。

## ADDED Requirements

### Requirement: 数据目录唯一真相
应用数据目录（数据库 / 配置 / 日志 / 崩溃日志的共同落点）MUST 只有一个实现出处，
各消费方共同引用，不得各自演算路径。

#### Scenario: 消费方共用同一出处
- **WHEN** 检查 `internal/config`、`internal/database`、`internal/logging`、`internal/crash`
- **THEN** 四者均引用同一个数据目录实现，各自不出现 `os.UserConfigDir()` 的重复演算

#### Scenario: 落盘路径不变
- **WHEN** 数据目录实现被抽离后启动已装应用
- **THEN** 数据库、配置、日志仍落在原路径，存量数据无需迁移

#### Scenario: 恶意应用名被拒绝
- **WHEN** 应用名含路径分隔符或为上跳路径（如 `..`）
- **THEN** 取数据目录失败并报错，不把数据写到应用目录之外

### Requirement: 数据目录用户可达
应用 MUST 让用户能定位并打开数据目录，以便自行取用日志。

#### Scenario: 查询数据目录路径
- **WHEN** 前端调用 `GetDataDir()`
- **THEN** 返回数据目录绝对路径，供界面展示或复制

#### Scenario: 一键打开数据目录
- **WHEN** 前端调用 `OpenDataDir()`
- **THEN** 系统文件管理器打开该目录
- **AND** 失败（如精简 Linux 无 `xdg-open`）时返回结构化错误，不静默失败

#### Scenario: 路径含空格仍可打开
- **WHEN** 数据目录路径含空格（如 macOS 的 `Application Support`）
- **THEN** 打开行为不受影响（路径作为独立参数传给系统命令，不经 shell 解释）
