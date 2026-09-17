# auto-release — 设计

## D1：版本号取自 tag，但仍写进单一真相

打 tag（`v0.1.0`）时，版本号的来源是 tag 而非仓库里的值。做法是：
**去掉前缀 v → 经 `scripts/set-version.sh` 写入 `wails.json`**，再由既有链路派生
（wails 渲染 macOS Info.plist / Windows exe 版本资源 / NSIS 注册表；
`build.sh` 经 `-ldflags` 注入应用内）。

**为何不直接把 tag 传给 wails**：`wails build` 没有覆盖版本的开关（前一轮已查证），
`info.productVersion` 是唯一入口。走这条既有链路，CI 与本地发布的行为就是同一个。

CI 里不需要还原 `wails.json`（runner 是一次性的）。

## D2：`set-version.sh` 抽出来，而不是在 CI 里再写一遍

`release.sh` 里原本内联着「校验格式 → 写 wails.json → 回读确认」。如果 CI 再实现一份，
就出现两处写同一个字段——**版本号是单一真相，写它的代码也该是**。
故抽成脚本，两处共用。

格式校验不可省：NSIS 的 `VIProductVersion` 要求数字点分，非数字版本（`v1.2.3` 的前缀、
`git describe` 的 `abc1234`）会让 Windows 打包在最后一步失败。**早失败好过打包打到一半才炸**。

## D3：三个 job，而不是两个——避免两条腿抢建 Release

两条平台腿并行，若各自执行「没有就创建、有就上传」，两边可能同时判定「不存在」
然后同时创建，其中一个必失败。

故：

```
verify(ubuntu) → package(macos ∥ windows) → release(ubuntu)
```

- 产物经 `actions/upload-artifact` 在 job 间传递，最后**单点**创建 Release 并附加全部产物
- 顺带的好处：打包失败时不会留下一个没有产物的 Release

## D4：发版前跑静态检查

tag 可以打在任意提交上（包括没走过 CI 的分支）。从红提交发版、把有问题的安装包
发给用户，代价远高于多跑两分钟检查。故 `verify` job 是 `package` 的前置。

它需要与 `build.yml` 的 check job 相同的步骤（前端构建 + 测试 + lint），
因为根包用 `//go:embed` 嵌入了 `frontend/dist`。这段重复是刻意的：
跨工作流抽 composite action 会让「读一遍就知道 CI 干了什么」这件事变难。

## D5：母版跳过打包，但 Release 照发

本仓库自己是母版，`package.sh` 的防呆会拒绝打包（这是对的：母版没有可发布的产物）。
工作流因此分两种状态：

| 状态 | 行为 |
|---|---|
| 母版（含占位符） | 跳过打包；Release 只带 GitHub 自动附加的源码归档 |
| 实例化项目 | 正常打包，安装包作为 Release 资产 |

检测用占位符，且**必须拆写**（`PH='__APP_''NAME__'`）：`.yml` 不在 `init.sh` 的替换
文件类型列表里，写成连续字面量会在实例化后依然匹配，导致「永远是母版」的误判。
这与 `package.sh` 防呆用的是同一个坑与同一个对策。

## D6：Release 的默认形态

`gh release create` + `--generate-notes`（变更说明由 GitHub 按提交/PR 自动生成），
**不加 `--draft`**——目标是让 Releases 页真的出现条目。

> 若希望「先审后发」，在 `release.yml` 的 `gh release create` 上加 `--draft` 即可：
> 产物照旧构建上传，但需人工点 Publish 才公开。本设计选择全自动，因为发版动作本身
> 由人推 tag 触发，已经是显式意图。

## D7：验证边界（哪些能本地验、哪些只能靠首次 tag）

| 项 | 手段 |
|---|---|
| `set-version.sh` | 本地合法/非法两路都跑（非法须拒绝且不改文件） |
| 工作流 YAML 结构 | `yaml.safe_load` + 打印 job 依赖 |
| 母版检测 | 本地 `git grep` 验证母版命中 |
| **`package.sh macos` 路径** | 在一次性实例上真跑（该路径此前从未执行过） |
| Windows NSIS / 真实 Release 创建 | **只能靠首次推 tag 验证**（创建公开 Release 属对外动作，不由 agent 发起） |
