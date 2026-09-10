# package-release — 技术设计

## Context

Wails v2.15.0 的打包能力（已实测源码 `cmd/wails/build.go`）：
- macOS 只产 `.app` bundle，不产 `.dmg`；`hdiutil` 可封装 dmg。
- Windows 支持 `-nsis` 生成 .exe 安装器，且**可在 mac/linux 上交叉编译**（唯一不限制交叉编译的平台）。
- linux/darwin 交叉编译被硬限制（`Crosscompiling to Mac not currently supported`）。

## Goals / Non-Goals

**Goals:**
- `make package` 无参 = 当前平台真安装包（mac 产 .dmg，win 产 .exe 安装器）。
- `make package windows` 在 mac/linux 上交叉编译出 .exe。
- 母版占位符防呆：含 `__APP_NAME__` 时拒绝打包。
- `make build` 保留快速编译（.app 不封装）。

**Non-Goals:**
- 不做 linux/mac 交叉编译（Wails 硬限制，明确报错引导）。
- 不做 .pkg（需开发者证书签名，过重）。
- 不做代码签名/公证（v1.1）。

## Decisions

### D1: 参数化用 `make package [OS]`，OS ∈ {空, windows, macos, linux}
- 空 → 当前平台。
- `windows` → 交叉编译，唯一被 Wails 支持的跨平台目标。
- `macos`/`linux` → 仅当 `uname` 匹配时允许，否则报错"请在 X 平台执行"。

### D2: mac 安装包用 hdiutil 封装 .dmg
`.app` 打包后执行：
```bash
hdiutil create -volname "<name>" -srcfolder "build/bin/<name>.app" -ov -format UDZO "build/bin/<name>.dmg"
```
UDZO = 压缩只读映像，标准分发格式。零依赖（系统自带）。

### D3: 母版防呆用占位符残留检测
打包入口先 `grep -rq '__APP_NAME__' go.mod wails.json main.go`，命中则报错退出，提示"母版不能直接打包，请先 scripts/init.sh 实例化"。这是 init 之后母版"不可直接跑"的最后一层兜底（design D3 的落地）。

### D4: 拆 `scripts/package.sh` 承载逻辑，Makefile 薄封装
打包逻辑（平台判断 + 防呆 + dmg/nsis 分支）放 `scripts/package.sh`，Makefile 只 `bash scripts/package.sh $(os)`。理由：逻辑复杂（平台分支 + 错误处理），shell 脚本比 Makefile 内联可读、可测。

### D5: `make build` 保持 = `wails build`（快速 .app），`make package` = 真安装包
职责分离：build 开发用（快）、package 分发用（慢但完整）。README 明确二者区别。

## Risks / Trade-offs

- **[Risk] hdiutil 在非 mac 环境不存在** → package.sh 里 mac 分支才调用 hdiutil，且前置检查 `command -v hdiutil`。
- **[Risk] NSIS 交叉编译需 wine/额外工具** → Wails 内置 NSIS 生成，实测无需额外依赖；若 CI 失败，降级为"仅产裸 exe + 提示"。
- **[Risk] dmg 封装含 .app 符号链接，hdiutil 需保留结构** → 用 `-srcfolder`（保留内部结构），不用 `-srcfile`。
- **[Risk] 母版防呆误伤"已 init 但某处漏替换"的项目** → 防呆只在 `__APP_NAME__` 残留时触发，init 已有"替换后断言零残留"，两者一致。
