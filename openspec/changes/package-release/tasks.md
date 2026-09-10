# package-release — 实施任务

## 1. 打包脚本

- [ ] 1.1 建 `scripts/package.sh [os]`：母版防呆（`__APP_NAME__` 检测）+ 平台判断 + 错误处理
- [ ] 1.2 mac 分支：`wails build` 后 `hdiutil create` 封装 .dmg
- [ ] 1.3 windows 分支：`wails build -nsis` 产 .exe 安装器（含交叉编译）
- [ ] 1.4 linux 分支：`wails build` 产二进制（当前平台限定）

## 2. Makefile 接线

- [ ] 2.1 `make package` 改为 `bash scripts/package.sh $(os)`（参数化 + 薄封装）
- [ ] 2.2 `make build` 保持 `wails build`（快速编译，不封装）

## 3. 文档

- [ ] 3.1 README 打包章节更新：产物位置、`make package [os]` 用法、交叉编译限制（mac/linux 不可交叉）

## 4. 验证

- [ ] 4.1 母版防呆：当前母版（含占位符）执行 `make package` 报错
- [ ] 4.2 init 实例化后的项目 `make package` 产 .dmg（mac）或 .exe（windows）
- [ ] 4.3 `make package windows` 在 mac 上交叉编译出 .exe
- [ ] 4.4 `openspec validate package-release` 通过
