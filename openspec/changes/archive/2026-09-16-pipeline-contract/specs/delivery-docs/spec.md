# delivery-docs (delta)

## MODIFIED Requirements

### Requirement: 交付文档齐全
README MUST 覆盖环境准备、开发、打包、目录结构、命令表、换肤与部署说明；
其余交付文档（`docs/`）MUST 与仓库实际交付一致，不得承诺未实现的代码，也不得残留其他项目的架构。

#### Scenario: 文档覆盖
- **WHEN** 查看 README
- **THEN** 包含：环境准备（Go/Node/wails）、`make` 命令表、目录结构（指向 docs/map.md）、换肤说明（改 name/logo/primary）、Windows WebView2 部署说明

#### Scenario: 文档与实际交付一致
- **WHEN** 阅读 `docs/脚手架功能说明.md` 的能力清单
- **THEN** 其中每一项在代码中真实存在，未实现的承诺已被删除或标注为「不含」

#### Scenario: 导航地图完整
- **WHEN** 按 AGENTS.md 指示先读 `docs/map.md`
- **THEN** 已落地路径（runtime/service/config 等）均已列出，不留「后续补充」的 TODO
