# scaffold-init — 实施任务

## 1. 母版占位符化

- [x] 1.1 10 处文件的 `wails-bailer-starter` → `__APP_NAME__`（含 scripts/gen.sh、release.sh、smoke.sh，超出原 7 处清单）
- [x] 1.2 `grep -rn 'wails-bailer-starter'`（排除 archive/go.sum/settings.local）确认零残留

## 2. init 脚本

- [x] 2.1 建 `scripts/init.sh <name>`：校验 name 合法性 + `git clone` 母版 + 全局替换 `__APP_NAME__`
- [x] 2.2 清空 `openspec/changes/archive` + 重新 `git init` + 首次提交
- [x] 2.3 末尾断言 `grep -rn '__APP_NAME__'` 残留为空，非空报错
- [x] 2.4 打印"下一步"指引

## 3. 文档同步

- [x] 3.1 README 顶部加"实例化"章节、换肤章节改为"占位符 + init.sh"

## 4. 验证

- [x] 4.1 `scripts/init.sh testapp` 生成项目 → `make build` 通过（产出 testapp.app）
- [x] 4.2 生成项目无母版私货（无 .git 历史、无 settings.local.json、无 archive）
- [x] 4.3 生成项目保留治理基线（openspec/specs、AGENTS.md、deps.yaml、Makefile）
- [x] 4.4 `openspec validate scaffold-init` 通过
