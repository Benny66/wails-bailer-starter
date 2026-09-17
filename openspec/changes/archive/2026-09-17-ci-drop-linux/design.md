# ci-drop-linux — 设计

## D1：为什么撤掉而不是修好

修好只需一行（在 Linux 构建时传 `-tags webkit2_41`）。之所以仍选择撤出矩阵：

- **失效来自环境而非代码**：cgo 要链哪个 webkit 取决于 runner 镜像装了哪个包。
  今天 24.04 只有 4.1，明天镜像换代或改回双版本，这条腿就会再红一次，
  而每次排查都要重新走一遍「查 wails cgo 指令 → 查 pkg-config → 对照镜像包」。
- **Linux 不是这个基座的目标平台**：打包脚本支持它，但主要使用路径是
  macOS / Windows（Windows 还能从 macOS 交叉编译）。
- **Go 层覆盖没有损失**：静态检查腿仍在 ubuntu 上跑，测试 / 护栏 / lint /
  生成器端到端全都覆盖 Linux 的纯 Go 行为。

**代价说清楚**：Linux 产物从此无 CI 验证。这是有意接受的——用一个「经常红、
又要花时间判断是不是环境问题」的腿，换来的信任度低于它消耗的注意力。

## D2：Linux 本地路径必须修，不能只撤

撤出 CI 不等于放任 `make package os=linux` 报一句 「找不到 webkit2gtk-4.0」。
故脚本里补上标签，并把原因写在旁边：

```bash
bash scripts/build.sh -clean -tags webkit2_41
```

**为何不做自动探测**（检测到只有 4.1 才加标签）：那等于把「猜环境」的逻辑写进
构建脚本，且无 CI 可验证——正是我们刚刚决定不去维护的那类东西。
一个明确的标签 + 一段说明，比一段无人验证的探测代码更可靠。

## D3：verify-gen 落在哪条腿

它需要 `rsync`（Windows runner 没有）+ bash + wails CLI + 前端依赖。

- ubuntu 腿已撤 → 不能放那里
- Windows 腿无 rsync → 放不了
- macOS 腿：三样都有 ✓

故移到 macos 腿，仍只跑一次（检查内容与平台无关）。

## D4：验证方式

| 项 | 手段 |
|---|---|
| workflow 结构 | `yaml.safe_load` 解析 + 打印矩阵与各步 `if` 条件 |
| Linux 根因 | Docker 内 `pkg-config --exists` 直接问有没有 4.0 / 4.1 |
| 标签确实有效 | Docker 内带 `-tags webkit2_41` 构建根包（见 tasks） |
| 撤出后 CI 转绿 | push 后由 GitHub 触发，用 API 轮询各腿结论 |
