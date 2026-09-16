# frontend-pipeline-parity — 实施任务

## 1. 数据目录单一真相（Go）

- [x] 1.1 新增 `internal/appdir`：`Dir` / `File` + 5 个单测
      （路径落在用户配置目录下、目录自动创建、非法应用名/文件名拒绝）
- [x] 1.2 `internal/config`、`internal/database`、`internal/logging`、`internal/crash`
      四处改引用 `appdir`，删除各自的 `os.UserConfigDir()` 重复实现（落盘路径不变）

## 2. 数据目录可达（Go + 前端）

- [x] 2.1 新增 `internal/reveal`：跨平台打开目录（darwin `open` / windows `explorer` /
      linux `xdg-open`），未知平台返回明确错误；路径作为独立参数传入，不经 shell
- [x] 2.2 `app.go` 新增绑定 `GetDataDir()` / `OpenDataDir()`（经 `apperr.Wrap` 归一化）
- [x] 2.3 `Settings.vue` 增加「数据目录」一行：显示路径 + 打开按钮
- [x] 2.4 重新生成 wails bindings（`wails build`），`bindings_test` 护栏绿

## 3. 前端事件契约镜像

- [x] 3.1 `frontend/src/lib/event.ts`：`EventAction` 常量镜像 + `eventName()` +
      `EventProgress`/`EventResult` 类型 + `onEvent()`（返回取消函数，非 wails 环境安全退化）
- [x] 3.2 `frontend/src/composables/useEvent.ts`：组件卸载自动解绑
- [x] 3.3 文件头写入可抄的「长任务推进度」范例（Go `EventsEmit` + 前端 `useEvent`）

## 4. 前端分页契约镜像

- [x] 4.1 `frontend/src/lib/page.ts`：`PageRequest` / `PageResult<T>` 类型 + `PageSize`
      常量镜像 + `newPageRequest()`。**不**镜像 `Normalized()` 逻辑（一个规则两个实现
      比复制常量更糟）
- [x] 4.2 `frontend/src/composables/usePagedList.ts`：注入式 fetcher，收敛列表页状态与加载，
      加载成功后以 Go 回显的归一化值为准
- [x] 4.3 `_example/frontend/ExampleList.vue` 改用 `usePagedList`（页面脚本从 39 行降到 16 行）

## 5. 镜像一致性护栏

- [x] 5.1 `internal/guard/parity_test.go`：解析 Go 常量与 TS 镜像，双向断言一致
      （错误码 / 事件动作 / 页大小上下界）
- [x] 5.2 任一解析到 0 个结果即 Fatal，附「写法可能已变更，请同步更新护栏解析规则」
- [x] 5.3 破坏性验证（三种都实测会红）：取值漂移 / 镜像缺项 / 镜像多出项

## 6. 验证入口统一

- [x] 6.1 `frontend/package.json` 增加 `typecheck`；`make lint` 纳入 `vue-tsc`
- [x] 6.2 `make lint` 实测通过
- [x] 6.3 **顺手修复**：ESLint 基础 `no-unused-vars` 会把 TS 类型注解里的函数类型参数
      （`handler: (payload: T) => void`）误报为未使用 → 对 TS 关闭该规则，
      改由 `vue-tsc` 的 `noUnusedLocals`/`noUnusedParameters` 兜（带类型信息，更准；
      已破坏性验证会红）

## 7. 顺手修复：生成器产出不合规（既有 bug）

- [x] 7.1 复现：`make gen` 把 import 追加到 import 块尾部 → 顺序不符 gofmt →
      **`make gen` 之后 `make lint` 必红**，且报错指向 app.go，看不出是生成器的锅
      （已用 `git show HEAD:app.go` + 注入模拟证实为既有问题）
- [x] 7.2 `scripts/gen.sh` 产出后 `gofmt -w`（无 gofmt 时警告而非静默）
- [x] 7.3 `scripts/verify-gen.sh` 新增第 5 步「生成物格式合规」断言，防止回归

## 8. 文档同步

- [x] 8.1 `docs/map.md`：补 `internal/{appdir,reveal}`、`lib/{event,page}.ts`、`composables/`
- [x] 8.2 `docs/代码规范.md`：分页/事件章节补前端侧；新增「镜像纪律」一节；
      新增「命令与验证」一节（类型检查不可省）
- [x] 8.3 `docs/脚手架功能说明.md`：补数据目录可达、打开目录能力、前端契约辅助
- [x] 8.4 `AGENTS.md`：铁律 1 补 `appdir` 与镜像护栏；铁律 4 补 `make lint` 组成
- [x] 8.5 `README.md`：能力行、护栏清单、目录结构补新路径

## 9. 端到端验证与归档

- [x] 9.1 `make test` 通过（含全部护栏）
- [x] 9.2 `make lint` 通过（gofmt + 护栏 + vet + ESLint + vue-tsc）
- [x] 9.3 `make verify-gen` 端到端通过（连续生成 2 模块 → bindings → 编译 → gofmt → 护栏）
- [x] 9.4 `wails build` 通过（同时验证前端 `vue-tsc + vite build`）
- [x] 9.5 `openspec validate frontend-pipeline-parity --strict` 通过
- [x] 9.6 归档本变更；`pipeline-contract` 的规格基线随本变更补齐
      （原 `pipeline-contract` 变更未产出 specs delta，导致 `docs/代码规范.md`
      引用了不存在的 `openspec/specs/pipeline-contract/`）
