# release-integrity — 让产物说真话 + 让端到端检查接上电

## Why

上一轮探查实测出两个问题，都属「没有任何症状」类：

**1. 版本号与 bundle id 是双源漂移的，且都不可控。**

- macOS 从 `Info.plist` 读到的 `CFBundleShortVersionString` 恒为 **`1.0.0`**——
  不是空值，是 Wails 在 `info.productVersion` 缺失时的默认常量，**永不变化**。
  而 `app.log` 首行显示的是构建期注入的另一个版本。用户从「显示简介」看到的版本，
  与应用自己记录的版本永远对不上。
- `CFBundleIdentifier` 模板硬编码为 `com.wails.{{safeBundleID .Name}}`——
  等于把 Wails 当厂商，且所有用本脚手架的项目共享同一命名空间
  （macOS 的登录项、权限授予、文件关联都按 bundle id 记账）。
- 根因是「版本」有四处落点（macOS plist / Windows exe 版本资源 / NSIS 注册表 /
  应用内），前三个由 wails 从 `wails.json` 渲染、第四个只能靠 `-ldflags`，
  而此前两边的取值来源不同（一个用 Wails 默认值，一个用 `git describe`）。

**2. `verify-gen` 不在 CI 里。**

它是本仓**唯一**的端到端检查（连生 2 个模块 → bindings → 编译 → gofmt → 护栏），
注释里自己记着已抓到 3 个真实 bug。但 CI 只跑 `make test` / `make lint` / `make build`
——最有用的回归网完全靠开发者记得手动跑。

## What Changes

- **版本号单一真相**：定为 `wails.json` 的 `info.productVersion`（唯一能同时到达
  macOS/Windows/NSIS 三处产物的字段），并新增 `scripts/build.sh` 作为**唯一注入点**，
  把同一个值经 `-ldflags` 注入应用内；`make build` / `package.sh` / `release.sh`
  一律经它构建。
- **提交号单独注入**：版本号受 NSIS `VIProductVersion` 限制必须是数字点分格式，
  没有位置放提交信息，故新增 `InjectedCommit` 与版本分开注入以保追溯性。
- **构建后回读校验**（macOS）：断言产物 `Info.plist` 的版本确实等于单一真相——
  版本「没生效」此前只能靠用户看「显示简介」才发现。
- **bundle id 项目化**：`Info.plist` 的厂商前缀改为占位符，`init.sh` 生成
  `com.<你的域名>.<项目名>`（默认 `com.example`，并在结尾提示替换）。
- **`release.sh` 的版本参数真正落地**：写进 `wails.json` 的 `info.productVersion`
  （这就是「发布新版本」这个动作），并校验数字点分格式。
- **CI 接入 `make verify-gen`**（ubuntu 腿；Windows runner 无 rsync 且检查与平台无关）。
- **护栏**：`wails.json` 必须有合法的 `info.productVersion` / `productName`；
  `Info.plist` 不得沿用 `com.wails.*`；版本注入目标必须真实存在（改扫 `build.sh`）。

## 顺带修复的既有隐患

`init.sh` 的残留断言依赖 bash「在替换自身之前已把整个脚本读进内存」这一缓冲行为
才勉强成立——脚本一旦超过读取块大小，断言就可能读到被改写后的内容而失效。
改用仓库既有的拆写手法（`package.sh` 的 `PLACEHOLDER="__APP_""NAME__"`），
并新增 bundle 前缀的回读校验。

## Capabilities

### Modified Capabilities

- `package-release`: 版本号单一真相与注入点唯一化；bundle id 项目化；构建后回读校验。
- `scaffold-init`: 实例化时生成 bundle id 前缀，并把新占位符纳入残留断言。
- `ci-pipeline`: 生成器端到端验证接入 CI。
- `architecture-guardrails`: 新增版本/bundle id 单一真相护栏。

## Non-Goals

- **不引入版本号自动递增**（无 tag 时如何递增是发布策略问题，不是脚手架该定的规则）。
- **不改 `wails.json` 的 author 字段**（用户信息，属实例化后自行维护）。
- **不动 Windows NSIS 的打包流程**（本次只保证它拿到的版本是单一真相）。
- **不做依赖升级 / 体积优化 / 前端测试**（上一轮探查里的 P3–P6，另行评估）。

## Impact

- **修改代码**：`scripts/{build,package,release,init}.sh`、`Makefile`、`wails.json`、
  `build/darwin/Info.plist`、`internal/appinfo`、`internal/guard/wiring_test.go`、
  `.github/workflows/build.yml`。
- **新增资源**：`scripts/build.sh`。
- **破坏性**：**无**。版本号从 `1.0.0`（假的）变为真实值；实例化项目新增 bundle 前缀
  参数但保持默认值向后兼容。
