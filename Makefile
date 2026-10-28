# Toolbit 工程入口
#
# 仓库分为两个独立构建单元：
#   backend/   Go 模块（github.com/blue-cloud-net/Toolbit/backend）
#   frontend/  pnpm 项目（Vue 3 + TypeScript）
#
# 常用目标：
#   make lint         静态检查（Go + Markdown + 文档一致性 + 前端 ESLint 与类型检查）
#   make test         Go 单元测试与契约测试（带 -race）
#   make test-web     前端单元测试
#   make build-web    前端生产构建
#   make check-docs   文档一致性检查
#   make check-budget 前端体积预算检查
#   make verify       提交前全量校验（vet + test + test-web + lint + build-web + check-budget）

SHELL := /bin/sh
.SHELLFLAGS := -eu -c

BACKEND    ?= backend
FRONTEND   ?= frontend
GO         ?= go
GOLANGCI_LINT ?= golangci-lint
MARKDOWNLINT  ?= markdownlint-cli2
NODE          ?= node
PNPM          ?= pnpm

# golangci-lint 内嵌的 go/types 版本**不**得高于它所解析的标准库版本，
# 否则会以 "file requires newer Go version" 崩溃。
# 这里从 `golangci-lint version` 里读出它构建时使用的 Go 版本，并让分析用同一版本。
GOTOOLCHAIN ?= $(shell $(GOLANGCI_LINT) version 2>/dev/null | sed -n 's/.*built with \(go[0-9][0-9.]*\).*/\1/p')

.DEFAULT_GOAL := help

.PHONY: help
help: ## 显示可用目标
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*## "}; {printf "  %-14s %s\n", $$1, $$2}'

.PHONY: build
build: ## 编译后端所有包
	cd $(BACKEND) && $(GO) build ./...

.PHONY: vet
vet: ## go vet 静态检查
	cd $(BACKEND) && $(GO) vet ./...

.PHONY: test
test: ## 单元测试与契约测试（带竞态检测）
	cd $(BACKEND) && $(GO) test ./... -race -count=1

.PHONY: fmt
fmt: ## 格式化 Go 源码
	cd $(BACKEND) && $(GO) fmt ./...

.PHONY: lint-go
lint-go: ## Go 静态检查（golangci-lint）
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || { \
		echo "错误：未找到 $(GOLANGCI_LINT)。"; \
		echo "安装：go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; \
		exit 1; }
	cd $(BACKEND) && GOTOOLCHAIN=$(GOTOOLCHAIN) $(GOLANGCI_LINT) run

.PHONY: lint-md
lint-md: ## Markdown 风格检查
	npx --yes $(MARKDOWNLINT) "**/*.md"

.PHONY: check-docs
check-docs: ## 文档一致性检查（catalog / 错误码 / 配置项 / 状态行 / 禁词）
	$(NODE) scripts/check-docs.mjs

.PHONY: lint
lint: lint-go lint-md check-docs lint-web ## 全部静态检查

.PHONY: lint-web
lint-web: ## 前端 ESLint 与类型检查
	cd $(FRONTEND) && $(PNPM) lint
	cd $(FRONTEND) && $(PNPM) type-check

.PHONY: test-web
test-web: ## 前端单元测试
	cd $(FRONTEND) && $(PNPM) test

.PHONY: build-web
build-web: ## 前端生产构建
	cd $(FRONTEND) && $(PNPM) build

.PHONY: check-budget
check-budget: ## 前端体积预算（初始 JS 与单工具 chunk）
	$(NODE) scripts/check-budget.mjs

.PHONY: verify
verify: vet test test-web lint build-web check-budget ## 提交前全量校验
	@echo "✅ 校验通过"

