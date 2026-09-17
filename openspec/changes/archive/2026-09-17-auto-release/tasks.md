# auto-release — 实施任务

## 1. 版本写入收敛为单一实现

- [x] 1.1 新增 `scripts/set-version.sh`：格式校验（数字点分）→ 写入 wails.json → 回读确认
- [x] 1.2 `scripts/release.sh` 改为调用它，删掉重复的校验/写入逻辑
- [x] 1.3 验证：合法版本写入成功；非法版本（`v1.2.3`）被拒绝**且不改动文件**

## 2. 自动发布工作流

- [x] 2.1 `verify` job：发版前跑前端构建 + `make test` + `make lint`（不从不绿的提交发版）
- [x] 2.2 `package` job：macos/windows 两条腿；Windows 先确保 makensis（缺了只会降级成裸 exe）
- [x] 2.3 `release` job：`needs: [verify, package]`，单点创建 Release + 附加产物
      （避免两条腿抢建；且打包失败时不会留下空发布）
- [x] 2.4 版本取自 tag（去 `v` 前缀）→ `set-version.sh` 写入单一真相
- [x] 2.5 母版识别：占位符**拆写**检测（`.yml` 不在 init.sh 替换范围内，连续字面量会误判），
      命中则跳过打包、Release 照常创建
- [x] 2.6 产物按「应用名-版本-平台」重命名后上传
- [x] 2.7 `--generate-notes`：变更说明由 GitHub 按提交自动生成

## 3. 验证

- [x] 3.1 全部脚本 `bash -n` 通过；两个工作流 YAML 可解析、job 依赖正确
- [x] 3.2 母版检测逻辑：本地 `git grep` 确认母版命中
- [x] 3.3 **`package.sh macos` 真跑**（此前从未执行过的路径）：在一次性实例上产出
      6.2MB dmg，美化生效、版本回读一致
- [x] 3.4 `make test` / `make lint` 通过；22 个规格全绿
- [ ] 3.5 **未验证**：Windows NSIS 安装器路径、真实 Release 创建——两者只能在推 tag 后
      由 GitHub 执行（创建公开 Release 属对外动作，未由 agent 发起）

## 4. 文档

- [x] 4.1 `README.md` 发布章节：打 tag 即发（含母版行为说明）
- [x] 4.2 `docs/map.md`：`set-version.sh` 与 `.github/workflows/`
- [x] 4.3 `docs/脚手架功能说明.md`：自动发布条目

## 5. 归档

- [ ] 5.1 归档 + 由使用者推送（推 tag 会创建公开 Release，需使用者发起）
