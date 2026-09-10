# package-release — 实施任务

## 1. 打包脚本

- [x] 1.1 建 `scripts/package.sh [os]`：母版防呆（`__APP_NAME__` 检测）+ 平台判断 + 错误处理
- [x] 1.2 mac 分支：`wails build` 后 `hdiutil create` 封装 .dmg
- [x] 1.3 windows 分支：`wails build` 产 .exe（有 makensis 时 `-nsis`，无则降级裸 exe）
- [x] 1.4 linux 分支：`wails build` 产二进制（当前平台限定）

## 2. Makefile 接线

- [x] 2.1 `make package` 改为 `bash scripts/package.sh $(os)`（参数化 + 薄封装）
- [x] 2.2 `make build` 保持 `wails build`（快速编译，不封装）

## 3. 文档

- [ ] 3.1 README 打包章节更新：产物位置、`make package os=...` 用法、交叉编译限制

## 4. 验证

- [x] 4.1 母版防呆：当前母版（含占位符）执行 `make package` 报错
- [x] 4.2 init 实例化后的项目 `make package` 产 .dmg（mac）
- [x] 4.3 `make package os=windows` 在 mac 上交叉编译出 .exe（含 makensis 降级）
- [x] 4.4 `openspec validate package-release` 通过
