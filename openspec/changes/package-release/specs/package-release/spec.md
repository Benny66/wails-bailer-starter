# package-release

参数化打包、mac dmg 安装包、windows NSIS 安装器、母版防呆。

## ADDED Requirements

### Requirement: 参数化打包
`make package [OS]` MUST 支持参数化：无参打包当前平台，`windows` 参数在 mac/linux 上交叉编译出 .exe。

#### Scenario: 无参打包当前平台
- **WHEN** 执行 `make package`（无参数）
- **THEN** 打包当前平台产物（mac 产 .dmg，windows 产 .exe 安装器）

#### Scenario: 交叉编译 windows
- **WHEN** 在 mac/linux 上执行 `make package windows`
- **THEN** 交叉编译出 Windows .exe 安装器

#### Scenario: 非本机平台明确报错
- **WHEN** 在 mac 上执行 `make package linux`（或反过来）
- **THEN** 报错提示"请在对应平台执行"，不产出无效产物

### Requirement: mac dmg 安装包
macOS 打包 MUST 在 `.app` 之后封装为 `.dmg` 安装包。

#### Scenario: 产出 dmg
- **WHEN** 在 macOS 执行 `make package`
- **THEN** 产出 `build/bin/<name>.dmg`（含 .app 的压缩映像）

### Requirement: 母版防呆
打包 MUST 在检测到 `__APP_NAME__` 占位符残留时拒绝执行。

#### Scenario: 占位符残留拒绝打包
- **WHEN** 在含 `__APP_NAME__` 占位符的母版上执行 `make package`
- **THEN** 报错提示"请先 scripts/init.sh 实例化"，不产出 `__APP_NAME__.app` 等怪名字产物
