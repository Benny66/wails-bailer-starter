# app-config Specification

## Purpose
TBD - created by archiving change runtime. Update Purpose after archive.
## Requirements
### Requirement: 配置读写
应用 MUST 读写本地 `config.json`，落在用户配置目录，不存在时写默认值。

#### Scenario: 首次启动写默认
- **WHEN** 应用首次启动且无配置文件
- **THEN** 生成默认 `config.json` 到用户配置目录

#### Scenario: 读取已有配置
- **WHEN** 应用再次启动
- **THEN** 加载已有配置值，覆盖默认值

#### Scenario: 保存配置
- **WHEN** 前端调用保存配置绑定方法
- **THEN** 新值写回 `config.json` 并持久化

