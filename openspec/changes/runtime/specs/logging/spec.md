# logging

文件 + 控制台双写日志、分级与按大小轮转。

## ADDED Requirements

### Requirement: 双写日志
应用 MUST 同时输出控制台与文件日志，并支持分级。

#### Scenario: 分级输出
- **WHEN** 应用在开发态运行
- **THEN** Debug 级日志输出到控制台；生产态输出 Info 级

#### Scenario: 文件落盘
- **WHEN** 应用运行
- **THEN** 日志写入用户配置目录下的日志文件

### Requirement: 日志轮转
日志文件 MUST 按大小自动轮转，避免无限增长。

#### Scenario: 达到大小即轮转
- **WHEN** 日志文件超过设定大小
- **THEN** 自动轮转为新文件，保留设定份数，旧文件归档
