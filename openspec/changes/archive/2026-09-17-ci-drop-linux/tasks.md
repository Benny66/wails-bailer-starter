# ci-drop-linux — 实施任务

## 1. 定位 build job 失败根因

- [x] 1.1 取三条腿的结论：ubuntu 失败于 Build，macos/windows 成功
      → 确认是 Linux 专有，非本轮改动引入
- [x] 1.2 读 wails 源码：Linux 侧 cgo 按 build tag 选 webkit 版本
      （`#cgo !webkit2_41 pkg-config: webkit2gtk-4.0` / `#cgo webkit2_41 ...-4.1`）
- [x] 1.3 确认 wails CLI 不会自动加该标签（只有通用 `-tags`）
- [x] 1.4 Docker 实证：Debian 13 里 `webkit2gtk-4.0` 不存在、`webkit2gtk-4.1` 存在
      → 与 CI 安装的包正好错位

## 2. 撤出 Linux 编译矩阵

- [x] 2.1 矩阵改为 `[macos-latest, windows-latest]`，删除 Linux 构建依赖步骤
- [x] 2.2 `verify-gen` 从 ubuntu 腿移到 macos 腿
- [x] 2.3 保留 ubuntu 静态检查腿（Linux 的 Go 层覆盖不损失）
- [x] 2.4 workflow 结构核对（YAML 可解析、矩阵与 if 条件正确）

## 3. Linux 本地路径修好

- [x] 3.1 `package.sh` 的 linux 分支补 `-tags webkit2_41` + 原因注释
- [x] 3.2 `release.sh` 的 Linux 行同样补上
- [x] 3.3 刻意不做自动探测（避免把「猜环境」的逻辑写进构建脚本且无 CI 可验）

## 4. 文档

- [x] 4.1 `README.md`：平台依赖补标签说明；打包命令标注未验证；
      CI 说明改两平台；已知限制新增一条（含根因与代价）
- [x] 4.2 `docs/脚手架功能说明.md`：CI 条目与 verify-gen 腿同步

## 5. 验证与归档

- [x] 5.1 本地 `make test` / `make lint` / `make verify-gen` 通过
- [~] 5.2 **未完成**：Docker 内全量构建在 apt 安装阶段停滞（两次尝试，>20 分钟无输出），
      已终止。根因由【两端独立证据】支撑而非端到端复现：
      ① Wails 源码 `webkit2.go`（`//go:build linux`，无标签约束 = 默认路径）
      `#cgo !webkit2_41 pkg-config: webkit2gtk-4.0` / `#cgo webkit2_41 ...: webkit2gtk-4.1`；
      ② Docker 实测同代发行版（Debian 13）里 `-4.0` 不存在、`-4.1` 存在。
      ⚠ 一次早期探针出现不一致信号（不带标签的构建似乎也通过了），未能复现或推翻，
      如实记录以免读者把根因当成已被端到端证实。
- [x] 5.3 `openspec validate ci-drop-linux --strict` 通过 + 归档
- [x] 5.4 见报告（push 后轮询结果） push 后轮询 CI：静态检查腿 + macOS/Windows 两条腿均应转绿
