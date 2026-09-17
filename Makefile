# 统一命令入口——AI 与人只记 make <target>，不记零散脚本路径。
# 未实现的能力显式失败（exit 1），而非静默通过，避免误导。

# wails 装在 $GOPATH/bin 下（可能不在 PATH），动态解析而非硬编码。
WAILS := $(shell go env GOPATH)/bin/wails

# 项目自身的 Go 包（排除 node_modules 里被 npm 拉进来的 Go 代码，如 flatted）。
#
# 刻意用递归展开（=）而非立即展开（:=）：go list 会编译根包，而根包依赖
# frontend/dist；干净检出里该目录不存在，立即展开会让**任何** make 目标
# （包括 make help）都先喷一句 `pattern all:frontend/dist: no matching files found`。
# 递归展开只在使用到它的 recipe 里求值，那时 require-dist 前置检查已通过。
GO_PKGS = $(shell go list ./... | grep -v node_modules)

.PHONY: help dev build test lint smoke package gen verify-gen require-dist

help: ## 列出所有命令及用途
	@echo "可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  make %-12s %s\n", $$1, $$2}'

dev: ## 启动开发态（wails dev，含前端热更新）
	$(WAILS) dev

# 走 scripts/build.sh 而非裸 wails build：版本号需同时注入产物（Info.plist /
# Windows 版本资源）与应用内（app.log / GetAppInfo），后者只能靠 ldflags，
# 且两者必须同源。见 scripts/build.sh 头部说明。
build: ## 编译当前平台产物（快速，.app 不封装 dmg）
	bash scripts/build.sh

# 前置条件：根包（main.go）用 //go:embed 嵌入前端产物，而 frontend/dist 被 gitignore
# ——干净检出里没有它，于是编译根包会失败。报错原文是
#   `main.go:24:12: pattern all:frontend/dist: no matching files found`
# 再配上 `FAIL <模块名> [setup failed]`，看不出真正原因（CI 因此长期红着，
# 本地首次 clone 也会撞上）。故在此显式拦一道并给出修复命令。
require-dist:
	@test -d frontend/dist || { \
	  echo "错误：frontend/dist 不存在。"; \
	  echo "      根包用 //go:embed 嵌入前端产物，编译它需要先构建前端。"; \
	  echo "      修复：cd frontend && npm install && npm run build"; \
	  exit 1; \
	}

test: require-dist ## 运行 Go 测试（含架构护栏）
	go test $(GO_PKGS)

# 前端类型检查不可省：开发态 vite 只剥离类型不做检查，缺了这步类型错误会漂到打包才炸。
# 注意 recipe 内的 `cd frontend` 只影响该行（每行一个 shell），不会泄漏给后续行。
lint: require-dist ## 静态检查（gofmt + Go 护栏 + go vet + 前端 ESLint + TS 类型）
	@test -z "$$(gofmt -l . | grep -v node_modules)" || (echo "gofmt 未通过，请运行 gofmt -w 修复:" && gofmt -l . | grep -v node_modules && exit 1)
	go test ./internal/guard/ && go vet $(GO_PKGS)
	cd frontend && npx eslint "src/**/*.{vue,ts,js}" && npm run typecheck

smoke: ## 冒烟测试（构建 → 启动 → 断言 → 清理）
	bash scripts/smoke.sh

package: ## 打包真安装包（make package os=windows|macos|linux [scope=user|machine]）
	bash scripts/package.sh $(os) $(scope)

gen: ## 生成新模块（锚点注入 + 幂等）—— make gen name=<module>
	bash scripts/gen.sh $(name)

verify-gen: ## 生成器端到端验证（临时沙箱：生成 2 个模块 → 编译 → 护栏）
	bash scripts/verify-gen.sh
