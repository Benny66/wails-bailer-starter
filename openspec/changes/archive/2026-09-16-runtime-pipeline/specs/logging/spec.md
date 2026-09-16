# logging

日志：双写、轮转、分级，以及来源接入。

## ADDED Requirements

### Requirement: 框架与前端日志接入同一出口
Wails 的日志接口 MUST 接入应用日志，使前端与框架内部产生的日志落到同一个文件。

#### Scenario: 前端日志落盘
- **WHEN** 前端调用 Wails 的日志方法（如 `LogError`）
- **THEN** 该条日志写进应用的 `app.log`，并带有可区分来源的标记

#### Scenario: 框架内部日志落盘
- **WHEN** Wails 自身产生内部日志（资源加载失败、绑定错误等）
- **THEN** 同样写进 `app.log`，而不是只出现在无人可见的 stdout

#### Scenario: 接线缺失会红
- **WHEN** 窗口选项未设置日志接口
- **THEN** 护栏失败并说明后果（前端日志会静默丢失）

#### Scenario: 致命级别不重复退出
- **WHEN** 框架调用日志适配器的致命级别方法
- **THEN** 只记录日志、不调用进程退出（框架随后会自行退出，
  提前退出会让日志 flush 与资源释放来不及执行）

### Requirement: 构建模式决定日志级别
日志级别 MUST 由构建模式决定，且生产构建的判定必须真实有效。

#### Scenario: 开发构建输出调试日志
- **WHEN** 以开发模式运行（`wails dev` / 裸 `go build`）
- **THEN** 输出 Debug 及以上级别

#### Scenario: 生产构建不输出调试日志
- **WHEN** 以生产模式构建（`wails build`）并运行
- **THEN** 只输出 Info 及以上级别，避免调试噪音快速吃掉轮转窗口、
  把真正的错误冲走

### Requirement: 启动首行记录版本与路径
每次启动 MUST 在日志首行记录版本与日志文件自身路径。

#### Scenario: 日志自带版本上下文
- **WHEN** 查看任一 `app.log`
- **THEN** 首行含应用版本与平台，无需再问用户「你用的是哪个版本」

#### Scenario: 从日志定位日志
- **WHEN** 用户只提供了日志内容的一段
- **THEN** 可从中读到日志文件的绝对路径，便于取得完整日志
