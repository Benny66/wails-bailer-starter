# app-lifecycle

单实例锁、关到托盘、优雅关闭与崩溃捕获。

## ADDED Requirements

### Requirement: 单实例唤起
应用 MUST 保证单实例，二次启动时唤起已有窗口而非报错。

#### Scenario: 二次启动唤起
- **WHEN** 应用已运行，用户再次启动该应用
- **THEN** 已有窗口被唤起（置顶/恢复），新进程退出，不弹"已在运行"错误

### Requirement: 优雅关闭
应用退出 MUST 按依赖顺序回收资源（停托盘 → 落盘配置 → 关数据库），不残留进程。

#### Scenario: 退出无残留
- **WHEN** 应用退出
- **THEN** 数据库连接关闭、goroutine 回收，无残留进程

### Requirement: 崩溃落盘
应用崩溃时 MUST 将崩溃信息写入独立日志文件，便于事后排查。

#### Scenario: 崩溃可查
- **WHEN** 应用发生 panic 崩溃
- **THEN** 崩溃信息（含堆栈）写入 crash 日志文件，且不覆盖常规日志
