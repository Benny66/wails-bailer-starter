# design-tokens — 主题切换行为修正

## MODIFIED Requirements

### Requirement: 主题切换
系统 MUST 提供暗色与亮色主题，默认跟随系统偏好，可通过切换应用。

#### Scenario: 默认跟随系统
- **WHEN** 应用首次启动且用户未手动选择过主题
- **THEN** 主题初始值为系统偏好（系统亮色则亮色，系统暗色则暗色），且 `<html>` 的 `data-theme` 与初始值一致

#### Scenario: 切换主题
- **WHEN** 用户触发主题切换
- **THEN** 全站配色随之切换、设置页按钮高亮状态同步更新，无需刷新页面

#### Scenario: 手动选择后不再跟随系统
- **WHEN** 用户已手动选择过主题
- **THEN** 后续不再因系统偏好变化而自动改变主题
