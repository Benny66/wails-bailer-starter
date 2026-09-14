# 统一命令入口——AI 与人只记 make <target>，不记零散脚本路径。
# 未实现的能力显式失败（exit 1），而非静默通过，避免误导。

# wails 装在 $GOPATH/bin 下（可能不在 PATH），动态解析而非硬编码。
WAILS := $(shell go env GOPATH)/bin/wails

# 项目自身的 Go 包（排除 node_modules 里被 npm 拉进来的 Go 代码，如 flatted）
GO_PKGS := $(shell go list ./... | grep -v node_modules)

.PHONY: help dev build test lint smoke package gen verify-gen

help: ## 列出所有命令及用途
	@echo "可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  make %-12s %s\n", $$1, $$2}'

dev: ## 启动开发态（wails dev，含前端热更新）
	$(WAILS) dev

build: ## 编译当前平台产物（快速，.app 不封装 dmg）
	$(WAILS) build

test: ## 运行 Go 测试（含迁移框架冒烟测试）
	go test $(GO_PKGS)

lint: ## 静态检查（gofmt + Go 护栏 + go vet + 前端 ESLint）
	@test -z "$$(gofmt -l . | grep -v node_modules)" || (echo "gofmt 未通过，请运行 gofmt -w 修复:" && gofmt -l . | grep -v node_modules && exit 1)
	go test ./internal/guard/ && go vet $(GO_PKGS) && cd frontend && npx eslint "src/**/*.{vue,ts,js}"

smoke: ## 冒烟测试（构建 → 启动 → 断言 → 清理）
	bash scripts/smoke.sh

package: ## 打包真安装包（make package os=windows|macos|linux [scope=user|machine]）
	bash scripts/package.sh $(os) $(scope)

gen: ## 生成新模块（锚点注入 + 幂等）—— make gen name=<module>
	bash scripts/gen.sh $(name)

verify-gen: ## 生成器端到端验证（临时沙箱：生成 2 个模块 → 编译 → 护栏）
	bash scripts/verify-gen.sh
