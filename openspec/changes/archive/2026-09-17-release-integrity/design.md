# release-integrity — 设计

## D1：版本号单一真相 = `wails.json` 的 `info.productVersion`

**问题**：版本号要到达四处，但只有一处是 wails 管的。

| 落点 | 来源 | 用户在哪看到 |
|---|---|---|
| macOS `Info.plist` | wails 渲染 `{{.Info.ProductVersion}}` | 显示简介、活动监视器 |
| Windows exe 版本资源 | wails 渲染 `build/windows/info.json` | 文件属性 → 详细信息 |
| NSIS 注册表 `DisplayVersion` | wails 渲染 `wails_tools.nsh` | 控制面板 → 程序和功能 |
| 应用内（日志/`GetAppInfo`） | 只能靠 `-ldflags -X` | `app.log` 首行 |

前三个都渲染自 `wails.json` 的 `info.productVersion`，第四个 wails 不管。
而 `wails build` **没有**覆盖版本的命令行开关（查过 `--help`），
所以「让四处同源」只有一个办法：**以 `info.productVersion` 为准，由构建脚本把同一个值
再喂给 `-ldflags`**。

**为何不用 `git describe` 当版本号**：`wails.json` 的版本会被渲染进 Windows 的
`VIProductVersion "${INFO_PRODUCTVERSION}.0"`，而它**要求数字点分格式**——
`v1.2.3` 的 `v` 前缀、`git describe` 在无 tag 仓库下给出的短哈希，都会让
Windows 打包在最后一步失败。实测确认过格式约束，故：

- **版本号**（数字点分，人工/release.sh 维护）→ 四处产物
- **提交号**（`git rev-parse --short`，构建期自动）→ 单独注入，补追溯性

两者结合后，报障时既有「用户看到的版本」也有「精确到提交的定位」。

**校验前置**：`build.sh` 与 `release.sh` 都拒绝非数字点分版本。早失败好过
打包到 Windows 那一步才炸（那时已经花了几分钟编译）。

## D2：新增 `scripts/build.sh` 作为唯一注入点

此前版本注入散在 `package.sh` 与 `release.sh` 里各算一份（而 `make build` 谁都没走），
这正是双源漂移的温床。现在：

```
make build ─┐
package.sh ─┼─→ scripts/build.sh ─→ 读 wails.json 版本
release.sh ─┘                      ├→ 注入 -ldflags（应用内）
                                   ├→ wails 渲染产物（Info.plist / exe / NSIS）
                                   └→ 回读产物校验
```

`build.sh` 透传所有参数（`-clean` / `-platform ...` / `-nsis` …），
故调用方语法不变。

**回读校验只做 macOS**：其它平台缺可靠的读取手段（exe 版本资源要解析 PE、NSIS 需
makensis）。macOS 用 `plutil -p` 读回产物 `Info.plist`，不一致即失败——
「版本没生效」此前只能等用户看「显示简介」才发现，本地构建期就拦住它。
交叉编译时**跳过**该校验：`-platform windows/...` 时 `build/bin` 里可能残留上一次的
`.app`，读了会得出错误结论（故先解析 `-platform` 参数判断）。

## D3：bundle id 厂商前缀只有 init 期能决定

`build/darwin/Info.plist` 的模板硬编码了 `com.wails.{{safeBundleID .Name}}`，
而 `wails.json` 的 `info` 结构里**没有** bundle id 字段（查过源码：
`Info` 只有 companyName / productName / productVersion / copyright / comments /
fileAssociations / protocols）。所以厂商前缀只能在**实例化时**改模板本身。

于是引入第二个占位符 `__BUNDLE_PREFIX__`，由 `init.sh` 替换（默认 `com.example`，
可选第二参数）。选 `com.example` 而非保留 `com.wails` 的理由：前者是 IANA 保留给示例的
域名，**明显是占位符**，用户会去改；后者看起来像真值，没人会动它。

`init.sh` 结尾还会回读校验 bundle id 真的进了 plist——占位符「替换没生效」比
「替换错」更难发现（`__BUNDLE_PREFIX__.myapp` 要到 macOS 侧栏或权限授予时才显形）。

## D4：`init.sh` 自替换的缓冲依赖（既有隐患）

`init.sh` 自己也在替换范围内。此前它把断言字符串写成裸 `__APP_NAME__`，
能工作纯属侥幸：bash 恰好在替换发生前已把整个脚本读进内存（脚本小于读取块）。
脚本一变大，就可能读到被改写后的内容——断言变成「搜项目名」，永远报残留。

**修法**：用仓库既有的拆写手法（`package.sh` 已有先例并注明原因）：

```bash
PH_APP="__APP_""NAME__"      # 文件里不存在连续字面量，替换不会碰它
```

顺带把 `find` 的替换范围加上 `*.plist`（否则新占位符在 plist 里替换不到）。

## D5：CI 加 verify-gen 的落点

`verify-gen` 需要 wails CLI + GTK 依赖 + rsync：

- `check` job：无 wails、无 GTK → 跑不了
- `build` job：三样都有，但 **windows runner 没有 rsync**，且该检查与平台无关

故加在 `build` job 的 ubuntu 腿（`if: matrix.os == 'ubuntu-latest'`）——
跑一次即可，不重复三遍。
