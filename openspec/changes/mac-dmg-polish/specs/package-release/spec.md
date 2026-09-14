# package-release (delta)

## MODIFIED Requirements

### Requirement: mac dmg 安装包
macOS 打包 MUST 在 `.app` 之后封装为 `.dmg` 安装包，并在产出后**回读校验**美化结果，
使「已美化」与「降级产物」可区分。

#### Scenario: 产出 dmg
- **WHEN** 在 macOS 执行 `make package`
- **THEN** 产出 `build/bin/<name>.dmg`（含 .app 的压缩映像）

#### Scenario: 回读校验美化结果
- **WHEN** dmg 产出完成
- **THEN** 二次挂载成品只读卷，检查 `.DS_Store` 与背景图引用：
  命中则打印 `✓ 美化已生效`，缺失则打印 `⚠ 降级产物（无美化）`（默认不阻断，退出码 0）

#### Scenario: 校验本身失败不误报
- **WHEN** 回读校验因无法挂载等原因失败
- **THEN** 报「无法校验」而非误报「无美化」，并提示校准方法可能已随 macOS 版本变化
