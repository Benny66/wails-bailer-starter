# pipeline-contract

Go↔TS 管道契约：错误协议、分页协议、事件协议。

## ADDED Requirements

### Requirement: 错误协议
绑定方法（app.go 的 App 方法）返回的业务错误 MUST 为结构化错误，携带机器可读的
`code` 与面向用户的中文 `message`；前端 MUST 能据 `code` 区分错误类别而不解析文案。

#### Scenario: 业务错误可区分
- **WHEN** 绑定方法因记录不存在而失败
- **THEN** 前端捕获到 `code = "not_found"` 与中文 `message`，可据此分流展示

#### Scenario: 未知错误有兜底
- **WHEN** 绑定方法返回未包装的原始错误（如数据库驱动错误）
- **THEN** 前端归一化后得到 `code = "internal"` 与中文兜底文案，不泄漏原始英文堆栈给用户

#### Scenario: 上游形态不外泄
- **WHEN** Wails 传递 error 的底层形态（字符串或对象）发生变化
- **THEN** 前端归一化包装吸收该差异，下游拿到的形态保持稳定

### Requirement: 分页协议
列表类绑定方法 MUST 返回结构化分页结果，包含数据列表与总数；分页请求参数 MUST 经归一化
（页码下界、页大小上下界）。页大小上下界常量 MUST 单一真相。

#### Scenario: 列表返回分页结果
- **WHEN** 前端请求某模块列表（带 page / page_size）
- **THEN** 返回含 `list` 与 `total` 的分页结果，前端可据此渲染分页控件

#### Scenario: 分页参数越界归一化
- **WHEN** 请求页码 < 1 或页大小超过上限
- **THEN** 归一化为合法值（页码夹到 1、页大小夹到上界），不报错也不返回全表

#### Scenario: 页大小上下界单一真相
- **WHEN** service 使用默认页大小或上限
- **THEN** 取值来自唯一定义的常量，不在多处硬编码

### Requirement: 事件协议
Go 侧向前端推送事件时，事件名 MUST 遵循 `<domain>:<action>` 约定，且 domain 与业务模块名
一致；进度类 payload MUST 含 `done`/`total`，结束类 MUST 含 `ok`。

#### Scenario: 长任务推进度
- **WHEN** Go 侧执行长任务并向前端推进度
- **THEN** 事件名为 `<domain>:progress`，payload 含 `done` 与 `total`，前端据此渲染进度

#### Scenario: 任务结束
- **WHEN** 长任务完成或失败
- **THEN** 发出 `<domain>:done`（含 `ok`）或 `<domain>:error`（含 `ok=false` 与 message）

### Requirement: 契约护栏
契约中可判定的部分 MUST 编译为会失败的检查（对齐架构护栏铁律）：
分页结构体字段齐全、页大小常量存在且有夹取测试、生成器产出的列表方法返回分页结果。

#### Scenario: 分页结构被改动
- **WHEN** 有人从分页结果结构体中删除 `total` 字段
- **THEN** 护栏失败并提示契约形状已变更

#### Scenario: 护栏感知自己瞎了
- **WHEN** 护栏解析到 0 个目标（如绑定文件缺失）
- **THEN** 护栏失败并提示「写法可能已变更，请同步更新护栏解析规则」，而非静默通过

## MODIFIED Requirements

### Requirement: 规范文档单一真相
`docs/代码规范.md` 描述的分层与调用方式 MUST 与本仓实际一致，不得残留其他项目的架构遗骸。

#### Scenario: 无架构遗骸
- **WHEN** 阅读 `docs/代码规范.md`
- **THEN** 不出现本仓不存在的层（如 controller）或不存在的文件（如 `api/index.js`）
