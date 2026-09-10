# example-module — 黄金范例与代码生成器

## Why

复用型基座的核心能力是"别人拿来就能长出模块"。但基座开箱**不能带业务代码**——Asset 这类示例模块一旦留在正式代码里，就是用户拿来时的"脏代码"。正确形态：只交付**最小干净的 `_example/` 模板 + `make gen` 生成器**，用户用生成器产出自己的模块。模板正确性用「生成→编译→删除」验证，验证产物不残留。

## What Changes

- **`_example/` 黄金范例**：最小、干净、带 `// TODO` 锚点的模板，四段——model + service + 绑定方法 + 前端页面。不含任何真实业务字段/逻辑。
- **`make gen name=<tool>` 生成器**：锚点注入（`// gen:model` / `// gen:bind` / `// gen:route` 等）+ fail-fast 前置校验 + 幂等/拒绝覆盖 + 占位符替换（Example→Tool）。
- **护栏断言锚点存在**：锚点删了会坏生成器，护栏同步断言锚点存在（刻意的耦合）。
- **验证但不残留**：用 `make gen` 生成一个临时模块 → 编译通过 → 删除，证明生成器与模板可用，但正式代码不留下任何示例模块。

## Capabilities

### New Capabilities

- `code-generator`: `_example/` 黄金范例 + `make gen` 生成器（锚点注入、幂等、TODO 锚点）。

### Modified Capabilities

- `project-scaffold`: `make gen` 从"未实现占位"改为"真实生成器"。

## Impact

- **新增代码**：`_example/` 模板（4 段）、`scripts/gen.sh`、护栏锚点断言。
- **依赖**：无新外部依赖（生成器用 shell + sed）。
- **护栏同步**：`internal/guard/` 新增锚点存在性断言。
- **破坏性**：无。`AllModels()` 保持空注册表（仅含 `// gen:model` 锚点）。
- **明确不含**：不新增任何 Asset 等业务模型/服务/绑定/页面；开箱为零业务代码。
