# app-config

应用本地配置与数据目录。

## ADDED Requirements

### Requirement: 窗口几何配置项
配置 MUST 持久化窗口尺寸与最大化状态，且取值必须经合法化后才可用于创建窗口。

#### Scenario: 字段落盘
- **WHEN** 应用正常退出
- **THEN** `config.json` 含窗口宽高与最大化状态（snake_case 字段名）

#### Scenario: 取值合法化集中一处
- **WHEN** 任何调用方需要窗口尺寸（如需创建窗口）
- **THEN** 通过统一的取值方法获得，非法值在该方法内回退默认，
  调用方不各自判断合法性
