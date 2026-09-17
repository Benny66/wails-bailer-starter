# ci-drop-linux — Linux 产物编译撤出 CI 矩阵

## Why

修好 `check` job 后，`build` job 首次真正执行（此前因 `needs: check` 每次被跳过），
三条腿里只有 ubuntu 失败在 `Build`：

```
build (ubuntu-latest)  failure  ['Build']
build (macos-latest)   success
build (windows-latest) success
```

**根因（已在 Docker 实证）**：Wails 在 Linux 上按 webkit2gtk 版本做 cgo 链接，
默认找 `webkit2gtk-4.0`；而 Ubuntu 24.04+ / Debian 13+ 只提供 `webkit2gtk-4.1`，
必须带 `-tags webkit2_41` 才切得过去。实测确认：

| 检查项 | 结果 |
|---|---|
| `pkg-config --exists webkit2gtk-4.0`（wails 默认要找的） | ✗ 不存在 |
| `pkg-config --exists webkit2gtk-4.1`（CI 装的那个） | ✓ 存在 |

CI 装了 4.1 的包却没传标签，于是 cgo 去找不存在的 4.0 → 编译失败。

这类失效的特点是**取决于 runner 镜像装了什么包**：今天修好标签，将来镜像换代
还会再坏一次。对一个由单人维护的脚手架，这种环境耦合的维护成本高于收益
——**尤其当 Linux 并不是这个基座的主要目标平台**。

## What Changes

- **编译矩阵撤掉 Linux**，只保留 macOS / Windows。
- **生成器端到端验证（`verify-gen`）从 ubuntu 腿移到 macos 腿**（Windows runner 无 rsync）。
- **静态检查腿仍是 ubuntu**：Linux 的 Go 层覆盖（测试 / 护栏 / lint / 生成器端到端）
  一点没少，只是不再承诺「Linux 产物能编译」。
- **Linux 本地打包路径修好**而非放任：`package.sh` / `release.sh` 的 Linux 构建补
  `-tags webkit2_41`，使 `make package os=linux` 在现代发行版上真的可用（但**无 CI 背书**）。

## Capabilities

### Modified Capabilities

- `ci-pipeline`: 编译矩阵从三平台收敛为 macOS/Windows；Linux 产物编译不在 CI 承诺内。

## Non-Goals

- **不删 Linux 支持**：`package.sh os=linux`、`internal/tray` 的 Linux 实现、
  Linux 打包说明都保留，只是标注「无 CI 验证」。
- **不做 webkit 版本的自动探测**：那等于把「猜 runner 装了什么」的逻辑写进构建脚本，
  与撤出 CI 的初衷相悖。
- **不引入 `continue-on-error` 之类的软化**：让腿「红着但不拦」比直接撤掉更糟
  ——失败会长期无人处理。

## Impact

- **修改文件**：`.github/workflows/build.yml`（矩阵 + verify-gen 腿）、
  `scripts/package.sh`、`scripts/release.sh`（Linux 构建标签）、`README.md`、
  `docs/脚手架功能说明.md`。
- **破坏性**：**无**（CI 覆盖范围收窄，本地能力不变）。
