# 统一命令入口——AI 与人只记 make <target>，不记零散脚本路径。
# 未实现的能力显式失败（exit 1），而非静默通过，避免误导。

# wails 装在 $GOPATH/bin 下（可能不在 PATH），动态解析而非硬编码。
WAILS := $(shell go env GOPATH)/bin/wails

# 项目自身的 Go 包（排除 node_modules 里被 npm 拉进来的 Go 代码，如 flatted）
GO_PKGS := $(shell go list ./... | grep -v node_modules)

.PHONY: help dev build test lint smoke package gen

help: ## 列出所有命令及用途
	@echo "可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  make %-10s %s\n", $$1, $$2}'

dev: ## 启动开发态（wails dev，含前端热更新）
	$(WAILS) dev

build: ## 编译当前平台产物（不打包安装包）
	$(WAILS) build

test: ## 运行 Go 测试（含迁移框架冒烟测试）
	go test $(GO_PKGS)

lint: ## 静态检查（Go 护栏 + go vet + 前端 ESLint）
	go test ./internal/guard/ && go vet $(GO_PKGS) && cd frontend && npx eslint "src/**/*.{vue,ts,js}"

smoke: ## 冒烟测试（构建后 IPC/绑定握手）—— 待 ci-release change 实现
	@echo "未实现：smoke 待 ci-release change 落地冒烟脚本" && exit 1

package: ## 打包安装包（当前平台，如 .app/.exe）
	$(WAILS) build -clean

gen: ## 生成新模块（锚点注入 + 幂等）—— 待 example-module change 实现
	@echo "未实现：gen 待 example-module change 落地 _example 模板与生成器" && exit 1
