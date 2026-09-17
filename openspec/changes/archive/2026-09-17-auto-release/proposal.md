# auto-release — 打 tag 即自动发布（GitHub Releases）

## Why

仓库的 Releases 页一直是空的（`No releases published`）。发版目前只能靠本地
`scripts/release.sh`：它产出的 dmg/exe **只落在本机**，要发布得手工上传；而且
macOS 的 dmg 打包只能在 mac 上做、Windows 安装器需要 makensis——**没有任何一个人
能在一台机器上产出全套安装包**。

同时仓库已经有两块现成的基础设施可以用上：

- `scripts/package.sh` 已能产出 macOS 美化 dmg 与 Windows NSIS 安装器
- 版本号已有单一真相（`wails.json` 的 `info.productVersion`），且构建期会注入应用内

缺的只是把它们接到 GitHub 的事件上。

## What Changes

- **新增 `.github/workflows/release.yml`**：推 tag（`v*`）即自动
  两平台打包 → 创建 Release（变更说明由 GitHub 自动生成）→ 附加安装包。
- **新增 `scripts/set-version.sh`**：版本号写入的**唯一实现**（格式校验 + 写入 +
  回读），由本地 `release.sh` 与 CI 发布工作流共用——两处各写一份版本是漂移的温床。
- **`release.sh` 改用它**，删掉重复的校验与写入逻辑。
- **发版前先跑静态检查**：tag 可以被打在任意提交上，不能从不绿的提交发版。
- **母版识别**：本仓库自己是**母版**（含占位符，`package.sh` 会拒绝打包），
  故工作流检测到占位符时跳过打包，Release 仍照常创建（带 GitHub 自动附加的源码归档）。
  实例化后的项目则自动带上安装包——同一份工作流两种状态都对。

## Capabilities

### Modified Capabilities

- `package-release`: 发布流程接入 CI（tag 触发）；版本写入收敛为单一实现。
- `ci-pipeline`: 新增发布工作流（含发版前的静态检查门禁）。

## Non-Goals

- **不做自动打 tag / 自动递增版本**（何时发版是人的决定，不是流水线的）。
- **不做语义化版本推断**（从提交信息猜 major/minor/patch 属另一套约定，本项目未采用）。
- **不发 Linux 产物**（与编译矩阵一致：Linux 不在 CI 覆盖内）。
- **不上传到第三方分发平台**（只发 GitHub Releases；分发渠道是下游自己的事）。

## Impact

- **新增**：`.github/workflows/release.yml`、`scripts/set-version.sh`。
- **修改**：`scripts/release.sh`（复用 set-version）、`README.md`、`docs/map.md`、
  `docs/脚手架功能说明.md`。
- **破坏性**：**无**。不改任何构建产物的形态；`release.sh` 的用法与参数不变。
