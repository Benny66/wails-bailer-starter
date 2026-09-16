# package-release Specification

## Purpose
TBD - created by archiving change package-release. Update Purpose after archive.
## Requirements
### Requirement: 参数化打包
`make package os=<OS>` MUST 支持参数化：无参打包当前平台，`windows` 参数在 mac/linux 上交叉编译出 .exe。

#### Scenario: 无参打包当前平台
- **WHEN** 执行 `make package`（无参数）
- **THEN** 打包当前平台产物（mac 产 .dmg，windows 产 .exe 安装器）

#### Scenario: 交叉编译 windows
- **WHEN** 在 mac/linux 上执行 `make package os=windows`
- **THEN** 交叉编译出 Windows .exe（无 makensis 时降级为裸 exe 并提示）

#### Scenario: 非本机平台明确报错
- **WHEN** 在 mac 上执行 `make package os=linux`（或反过来）
- **THEN** 报错提示"请在对应平台执行"，不产出无效产物

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

### Requirement: 母版防呆
打包 MUST 在检测到 `__APP_NAME__` 占位符残留时拒绝执行。

#### Scenario: 占位符残留拒绝打包
- **WHEN** 在含 `__APP_NAME__` 占位符的母版上执行 `make package`
- **THEN** 报错提示"请先 scripts/init.sh 实例化"，不产出 `__APP_NAME__.app` 等怪名字产物

