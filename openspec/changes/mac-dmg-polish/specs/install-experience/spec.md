# install-experience (delta)

## MODIFIED Requirements

### Requirement: mac dmg 美化引导
dmg MUST 通过 Finder 窗口布局提供拖拽引导（应用在左、Applications 在右、背景图/箭头）。
图标落点与背景图箭头 MUST 由同一组常量推导（箭头锚点 == 两图标落点中点），
不得各写各的魔数，也不得依赖「窗口宽/2」（窗口 bounds 宽 ≠ content 视口宽）。

#### Scenario: 打开 dmg 有可视化引导
- **WHEN** 用户打开 dmg 窗口
- **THEN** 呈现背景图与图标定位（应用在左、Applications 在右），箭头落在两图标之间，视觉上引导拖拽

#### Scenario: 箭头与图标对齐（跨机稳定）
- **WHEN** 打包产物在真机 GUI 下打开（任意 Finder 侧栏宽度）
- **THEN** 箭头中心与两图标落点的中点一致，不随侧栏宽度漂移

#### Scenario: 无背景图时仍保留布局
- **WHEN** 背景图缺失
- **THEN** 图标定位与窗口尺寸仍生效（仅跳过背景图），而非整段布局被跳过

#### Scenario: 美化失败不阻断
- **WHEN** Finder 布局设置因环境原因失败
- **THEN** dmg 仍生成（含软链的最小可用形态），仅缺失视觉美化
