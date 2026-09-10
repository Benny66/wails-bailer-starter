# design-tokens — 主题初始值来源修正

## MODIFIED Requirements

### Requirement: 主题切换
系统 MUST 提供暗色与亮色主题，默认跟随系统偏好，可通过切换应用，且用户选择持久化到 config。

#### Scenario: 首次启动跟随系统
- **WHEN** 应用首次启动且 config 中主题为空（未设置）
- **THEN** 主题初始值为系统偏好（系统亮色则亮色，系统暗色则暗色），且 `<html>` 的 `data-theme` 与初始值一致

#### Scenario: 已设置则用 config 值
- **WHEN** 应用启动且 config 中主题非空（用户之前选过）
- **THEN** 主题初始值为 config 值，不再跟随系统

#### Scenario: 切换主题并持久化
- **WHEN** 用户触发主题切换
- **THEN** 全站配色随之切换、设置页按钮高亮同步更新、主题值写入 config.json（重启后保持）

#### Scenario: 手动选择后不再跟随系统
- **WHEN** 用户已手动选择过主题
- **THEN** 后续不再因系统偏好变化而自动改变主题
