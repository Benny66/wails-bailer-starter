# release-integrity — 实施任务

## 1. 版本号单一真相

- [x] 1.1 `wails.json` 增加 `info` 块（productName / productVersion=0.1.0 / comments）
- [x] 1.2 新增 `scripts/build.sh`：读单一真相 → 格式校验 → 注入 `-ldflags`（版本 + 提交号）
      → 透传参数给 `wails build` → macOS 回读 Info.plist 校验
- [x] 1.3 `Makefile` 的 `build` 改走 `build.sh`
- [x] 1.4 `package.sh`：删除自算的 MODULE/VERSION/LDFLAGS，4 处 `wails build` 全改走 `build.sh`
- [x] 1.5 `release.sh`：版本参数写进 `wails.json`（发布动作）+ 格式校验 + 走 `build.sh`
- [x] 1.6 `appinfo`：新增 `InjectedCommit` 与 `Commit()`，版本与提交分开注入
- [x] 1.7 验证：**macOS 产物** `Info.plist` 版本 = 0.1.0（独立用 `plutil` 复读确认）
- [x] 1.8 验证：**Windows 产物** exe 版本资源含 `0.1.0`（UTF-16LE 检索确认）
- [x] 1.9 验证：三条失败路径均退出 1（版本非法 / 缺 productVersion / release 参数非法）
- [ ] 1.10 未实跑：NSIS 安装器路径（本机无 makensis）——机制同 exe 版本资源，已由 1.8 佐证

## 2. bundle id 项目化

- [x] 2.1 `build/darwin/Info.plist`：厂商前缀改为 `__BUNDLE_PREFIX__` 占位符（附原因注释）
- [x] 2.2 `init.sh`：新增可选第二参数（bundle 前缀，默认 `com.example`）+ 合法性校验
- [x] 2.3 `init.sh`：替换范围补 `*.plist`；结尾回读校验 bundle id 真的进了 plist
- [x] 2.4 验证：实例化 `testapp com.acme` → 产物 `CFBundleIdentifier = com.acme.testapp`

## 3. 顺带修复：init.sh 自替换的缓冲依赖

- [x] 3.1 占位符改拆写形式（`PH_APP="__APP_""NAME__"`），并把断言改用变量
- [x] 3.2 注释里的裸占位符字面量一并改写（否则实例化后被替换成无意义的句子）
- [x] 3.3 验证：`grep` 确认脚本内已无连续占位符字面量；实例化全流程复跑通过

## 4. 护栏

- [x] 4.1 注入目标护栏改扫 `scripts/build.sh`（注入点唯一化）
- [x] 4.2 新增：`wails.json` 必须有合法 `info.productVersion` 与非空 `productName`
- [x] 4.3 新增：`Info.plist` 不得沿用 `com.wails.*`，且必须用 `{{safeBundleID .Name}}`
- [x] 4.4 破坏性验证三种（缺版本号 / 版本格式非法 / bundle id 退回默认）均确认会红

## 5. CI

- [x] 5.1 `build.yml` 的 ubuntu 腿加 `make verify-gen`（附：为何只跑一腿、为何排除 Windows）
- [x] 5.2 本地复跑 `make verify-gen`（含新护栏）通过
- [ ] 5.3 未实跑：CI 上的真实执行（需 push 后由 GitHub 触发）

## 6. 文档

- [x] 6.1 `README.md`：实例化的 bundle 前缀参数与说明；版本单一真相一段；
      CI 说明补 verify-gen；已知限制去掉「`make build` 显示 dev」并改为版本格式约束
- [x] 6.2 `docs/map.md`：补 `scripts/build.sh`
- [x] 6.3 `docs/脚手架功能说明.md`：版本单一真相四处同源 / build.sh / CI 接入

## 7. 验证与归档

- [x] 7.1 `make test` / `make lint` 通过
- [x] 7.2 `make build` 通过（Makefile 已改走 build.sh，含回读校验）
- [x] 7.3 `make verify-gen` 通过
- [x] 7.4 `openspec validate release-integrity --strict` 通过
- [ ] 7.5 归档 + 提交推送
