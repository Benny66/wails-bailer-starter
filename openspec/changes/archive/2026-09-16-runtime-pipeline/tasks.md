# runtime-pipeline — 实施任务

## 1. 窗口几何持久化

- [x] 1.1 `internal/config`：新增 `WindowWidth` / `WindowHeight` / `WindowMaximised` 字段
- [x] 1.2 `Config.WindowSize()` 取值夹取（未记录/负数/过小 → 默认）+ 下界常量单一真相
- [x] 1.3 `internal/config/config_test.go`：7 组夹取用例 + 旧配置（无窗口字段）可加载 +
      落盘回读 + snake_case 字段名断言
- [x] 1.4 `main.go`：用配置值作为创建期 `Width/Height/WindowStartState`（首帧即正确）
- [x] 1.5 `app.go` shutdown：采集窗口几何回写；**最大化时不读尺寸**（否则"还原"尺寸
      会变成屏幕尺寸）+ 落盘前记一条「窗口几何已记录」日志
- [x] 1.6 **实跑验证**：配置 900×600 → 启动 → 干净退出 → 日志
      `窗口几何已记录 width=900 height=600`（从活窗口读回，证明创建期尺寸生效）
      + `config.json` 正确回写

## 2. 跨实例通信

- [x] 2.1 `app.go` 绑定 `GetLaunchArgs()`（返回副本，不外发 os.Args 全局切片）
- [x] 2.2 `main.go` `OnSecondInstanceLaunch` → `app:second-instance` 事件
      （ctx 未就绪只记日志不 panic）
- [x] 2.3 `ActionSecondInstance` 定义为 Go 侧常量，纳入镜像护栏
- [x] 2.4 `frontend/src/lib/app.ts`：`AppEvent` 常量 + `SecondInstancePayload` +
      `onSecondInstance()`

## 3. 可观测性

- [x] 3.1 `internal/logging/wails.go`：Wails `logger.Logger` 适配器（7 方法 → slog，
      `source=wails` 标记，`Fatal` **不** os.Exit）
- [x] 3.2 `internal/logging/wails_test.go`：各级别路由 + 标识 + 不退出 + 消息原样透传
- [x] 3.3 `main.go` 注入 `Logger` + `LogLevel` / `LogLevelProduction`
- [x] 3.4 `internal/logging`：启动首行记录版本 / 平台 / 日志文件路径
- [x] 3.5 `frontend/src/lib/log.ts`：日志出口（非框架环境退化 console）+ 单行化 + 截断
- [x] 3.6 `frontend/src/lib/log.ts`：全局错误兜底（`error` / `unhandledrejection`
      + 供 Vue 挂载的 `createVueErrorHandler`）+ 幂等安装与卸载
- [x] 3.7 `frontend/src/main.ts`：挂载**前**安装兜底；挂载后记「前端已挂载」
- [x] 3.8 **实跑验证**：`app.log` 出现 `msg=前端已挂载 source=wails`
      ——前端日志确实经该链路落到文件

## 4. 数据导出

- [x] 4.1 `internal/service/export.go`：`ExportDatabase`（数据目录内拒绝 →
      `VACUUM INTO` 同目录中转文件 → 原子改名 → 失败清理）
- [x] 4.2 单测 5 项：导出产物可被独立连接打开且数据完整 / 覆盖已有文件 / 不残留中转文件 /
      拒绝数据目录内目标 / 拒绝空路径

## 5. 版本与诊断

- [x] 5.1 `internal/appinfo`：`Resolve()` 分层回退 + `Platform()` + `Current()` + `String()`
- [x] 5.2 `app.go` 绑定 `GetAppInfo()`（返回 `appinfo.Info`，bindings 生成类型安全的
      `models.ts`）
- [x] 5.3 `scripts/package.sh`：注入 `git describe` 结果（模块路径从 go.mod 现读）
- [x] 5.4 `scripts/release.sh`：同样注入——让该脚本的 VERSION 参数**真正生效**
      （此前只用于产物文件名）

## 6. 护栏

- [x] 6.1 `internal/guard/wiring_test.go`：断言 `options.App` 设置了 `Logger`
- [x] 6.2 同上：断言打包脚本的 `-X` 注入目标在 appinfo 包真实存在
- [x] 6.3 两处均「解析到 0 个即 Fatal」+ 破坏性验证会红（注释掉 Logger 接线；
      把注入符号改成不存在的名字）

## 7. 顺手修复的真实 bug

- [x] 7.1 **生产态日志级别从未生效**：`main.go` 读环境变量 `WAILS_PRODUCTION`，
      而该变量在 Wails 全库中不存在（已全库确认）——判断永远不成立，生产构建也在
      按 Debug 输出，调试噪音会快速吃掉 5MB×3 的轮转窗口。
      改为按 Wails 真正使用的 **build tag `production`** 判定（`buildmode_*.go`），
      并在两种标签下各跑一次测试断言取值正确
- [x] 7.2 **ESLint `no-undef` 误报 DOM 类型**：基础规则不认识 `ErrorEvent` /
      `PromiseRejectionEvent` 等 TS 类型名。按官方建议对 TS 关闭该规则
      （TS 本身会报 `Cannot find name`，且知道得更多），与既有的 `no-unused-vars` 同理

## 8. 文档同步

- [x] 8.1 `docs/map.md`：补 `internal/appinfo`、`lib/{app,log}.ts`
- [x] 8.2 `docs/代码规范.md`：新增「前端日志与异常」一节
- [x] 8.3 `docs/脚手架功能说明.md`：补 6 项新能力（12–17）
- [x] 8.4 `AGENTS.md` 铁律 5：补推论「没有任何症状的失效必须有护栏」
- [x] 8.5 `README.md`：能力清单、护栏清单、已知限制（窗口位置不持久化 /
      `make build` 版本显示 dev / 裸 `go build -tags production` 链接失败）

## 9. 验证与归档

- [x] 9.1 `make test` 通过
- [x] 9.2 `make lint` 通过（含 vue-tsc）
- [x] 9.3 `make verify-gen` 通过
- [x] 9.4 `wails build` 通过（bindings + models.ts 重新生成）
- [x] 9.5 实跑构建产物：窗口几何 + 前端日志落盘均已验证（见 1.6 / 3.8）
- [x] 9.6 `openspec validate runtime-pipeline --strict` 通过
- [x] 9.7 归档
