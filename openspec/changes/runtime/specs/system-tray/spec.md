# system-tray

系统托盘、右键菜单与双击恢复。

## ADDED Requirements

### Requirement: 托盘常驻
应用 MUST 在 Windows/Linux 提供系统托盘图标，最小化/关闭时驻留托盘；macOS 因 Wails v2 架构冲突不支持托盘，关闭即退出。

#### Scenario: 关闭到托盘（Windows/Linux）
- **WHEN** 用户在 Windows/Linux 点击窗口关闭按钮
- **THEN** 窗口隐藏，应用驻留托盘，进程不退出

#### Scenario: macOS 无托盘
- **WHEN** 应用运行在 macOS
- **THEN** 不启动托盘，关闭窗口即退出应用（不残留无窗口进程）

### Requirement: 托盘交互
托盘 MUST 提供右键菜单与双击/单击恢复窗口（Windows/Linux）。

#### Scenario: 双击恢复
- **WHEN** 用户双击托盘图标
- **THEN** 窗口恢复显示并置前

#### Scenario: 右键菜单退出
- **WHEN** 用户右键托盘图标并选择退出
- **THEN** 应用优雅退出，托盘图标消失
