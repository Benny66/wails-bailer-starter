# scaffold-init Specification

## Purpose
TBD - created by archiving change scaffold-init. Update Purpose after archive.
## Requirements
### Requirement: 母版占位符化
脚手架母版 MUST 将模块名/品牌名统一为占位符 `__APP_NAME__`，不含任何具体项目名残留。

#### Scenario: 无具体名残留
- **WHEN** 在母版中 grep 模块名
- **THEN** 所有原本写 `wails-bailer-starter` 的位置均为 `__APP_NAME__`，无具体项目名硬编码

### Requirement: 一键实例化
`scripts/init.sh <name>` MUST 从母版生成一个可编译的新项目，替换零遗漏、不含母版私货。

#### Scenario: 生成可编译项目
- **WHEN** 执行 `scripts/init.sh myapp`
- **THEN** 生成 `myapp` 项目，`make build` 通过，模块名/import/wails.json/index.html title 全部为 `myapp`

#### Scenario: 无母版私货
- **WHEN** 检查 init 生成的项目
- **THEN** 不含母版的 `.git` 历史、`.claude/settings.local.json`、归档历史 `openspec/changes/archive`

#### Scenario: 非法名拒绝
- **WHEN** 执行 `scripts/init.sh` 传入空值或含非法字符的名称
- **THEN** 脚本报错退出，不生成任何项目

### Requirement: 保留治理基线
init 生成的项目 MUST 保留 OpenSpec 治理结构（specs 能力基线 + AGENTS/CLAUDE 宪法 + deps.yaml + Makefile），仅清空母版开发历史。

#### Scenario: 治理基线在
- **WHEN** 检查 init 生成的项目
- **THEN** 存在 `openspec/specs/`（能力基线）、`AGENTS.md`、`CLAUDE.md`、`deps.yaml`、`Makefile`、`_example/`

