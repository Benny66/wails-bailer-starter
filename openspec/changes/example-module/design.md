# example-module — 技术设计

## Context

地基已有 model 注册表（`AllModels()` 空）、database 迁移框架、design-system（本 change 依赖其 tokens/组件）。需要交付"能长出模块"的能力：**最小干净的 `_example/` 模板 + `make gen` 生成器**，开箱零业务代码。

> **方向修正（实施中确认）**：最初设计"资产表作为第一个真实模块留在代码里 + 提炼模板"，这会导致脚手架开箱带 Asset 业务代码 = 脏代码。正确形态是**干净基座**——正式代码只留模板 + 生成器，用户用 `make gen` 产出自己的模块。模板正确性用「生成→编译→删除」验证，验证产物不残留。

## Goals / Non-Goals

**Goals:**
- `_example/` 最小干净模板（model/service/绑定/前端页面四段 + TODO 锚点）。
- `make gen name=<tool>` 生成器（锚点注入 + 幂等 + TODO）。
- 护栏断言锚点存在。
- 用「生成→编译→删除」验证模板可用，且不残留。

**Non-Goals:**
- 不做任何真实业务模块（不新增 Asset 等模型/服务/绑定/页面）。
- 不做前端 CRUD 全自动生成（生成器生成骨架，业务手填 TODO）。

## Decisions

### D1: 干净基座——模板与生成器是交付物，示例模块不是
`_example/` 是"代码范例"（最小、干净、带 `// TODO` 锚点），是唯一交付物。没有"资产表作为可运行能力"这一层——那是脚手架**作者**开发时的验证路径，不是交付形态。验证用「`make gen` 生成临时模块 → 编译 → 删除」，验证完不留任何示例业务代码。

### D2: 生成器用 shell + sed，锚点注入
`scripts/gen.sh name=<tool>` 流程：
1. **fail-fast 前置校验**：锚点是否存在（`grep -qF '// gen:model' internal/model/model.go` 等）、目标资源是否被占用、目标文件是否已存在。
2. **锚点注入**：在 `AllModels()` 注入 `&Tool{}`、`app.go` 注入绑定方法、路由/菜单注入。
3. **幂等/拒绝覆盖**：目标文件已存在则 `exit 1`，绝不覆盖业务代码。
4. **占位符替换**：`Example`/`example` → `Tool`/`tool`（PascalCase/snake_case）。
5. **`// TODO` 锚点**：生成文件带 `// TODO: 业务逻辑`，AI/人只填锚点处。

### D3: 锚点是"刻意的耦合"，护栏断言锚点存在
`// gen:model` / `// gen:bind` / `// gen:route` 锚点既给生成器定位，也给护栏断言"锚点存在"。删了锚点 → 生成器坏 + 护栏红，两边一起暴露。

### D4: 占位符替换避开 `sed -i` 平台差异
用 `sed 's/.../.../' > tmp && mv` 而非 `sed -i`（macOS/BSD 与 GNU 语法不同）。

### D5: 模板前端页面复用 design-system tokens/组件
`_example/` 的前端页面用 tokens 变量与 EP 组件，不硬编码色值，作为"如何用设计系统"的活范例。

## Risks / Trade-offs

- **[Risk] 生成器中途失败留下不一致状态** → fail-fast 前置校验把"锚点/占用"全查完再动任何文件，避免"后端已生成、前端未注入"。
- **[Risk] 模板被写成具体业务（照抄某个真实工具）后变脏** → 护栏断言 `_example/` 无业务残留（无具体字段名），模板由人手工保持最小。
- **[Risk] 生成器与护栏规则耦合，改生成器忘改护栏** → 单一真相：锚点清单抽成 `scripts/anchors.sh` 供生成器与护栏共同引用。
- **[Risk] sed 转义在特殊字符上出错** → 用固定占位符 `Example`/`example`（无正则特殊字符），替换目标也是纯字母数字。
