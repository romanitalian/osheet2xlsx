SHELL:=bash

.DEFAULT_GOAL := help

.PHONY: help
help: ## Available commands
	@clear
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[0;33m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""

##@ Build

.PHONY: build
build: ## Build application
	go build -o bin/osheet2xlsx main.go

.PHONY: build-race
build-race: ## Build application with race detector
	go build -race -o bin/osheet2xlsx main.go

.PHONY: install
install: ## Install application
	go install

##@ Run

.PHONY: run
run: ## Run application
	go run main.go

.PHONY: run-example
run-example: ## Run with example file
	go run main.go convert examples/sample.osheet --out outputs/sample.xlsx --overwrite

.PHONY: run-typed
run-typed: ## Run with typed example file
	go run main.go convert examples/typed.osheet --out outputs/typed.xlsx --overwrite

##@ Test

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-race
test-race: ## Run tests with race detector
	go test -race -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-bench
test-bench: ## Run benchmark tests
	go test -bench=. ./...

##@ Lint & Format

.PHONY: lint
lint: ## Run linter (golangci-lint)
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix: ## Fix linting issues automatically
	golangci-lint run --fix ./...

.PHONY: lint-new
lint-new: ## Run lint on new files only
	git diff --name-only --cached | grep '\.go$$' | xargs -r golangci-lint run

.PHONY: lint-report
lint-report: ## Generate lint report
	golangci-lint run --out-format=html --out=lint-report.html ./...

.PHONY: lint-config
lint-config: ## Show current lint configuration
	golangci-lint config

.PHONY: format
format: ## Format code
	go install golang.org/x/tools/cmd/goimports@latest
	goimports -l -w .

.PHONY: fmt
fmt: ## Format code (alias for format)
	@make format

.PHONY: vet
vet: ## Run go vet
	go vet ./...

##@ Clean

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

##@ Development

.PHONY: deps
deps: ## Download dependencies
	go mod download
	go mod tidy

.PHONY: deps-update
deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

.PHONY: generate
generate: ## Generate code (if needed)
	go generate ./...

##@ Examples

.PHONY: examples
examples: ## Build all examples
	@mkdir -p outputs
	go run main.go convert examples/sample.osheet --out outputs/sample.xlsx --overwrite
	go run main.go convert examples/typed.osheet --out outputs/typed.xlsx --overwrite

.PHONY: examples-clean
examples-clean: ## Clean example outputs
	rm -rf outputs/

##@ Aliases

.PHONY: r
r: ## Run app
	@make run

.PHONY: t
t: ## Run tests
	@make test

.PHONY: l
l: ## Run linter
	@make lint

.PHONY: lf
lf: ## Fix linting issues automatically
	@make lint-fix

.PHONY: ln
ln: ## Run lint on new files only
	@make lint-new

.PHONY: lr
lr: ## Generate lint report
	@make lint-report

.PHONY: lc
lc: ## Show current lint configuration
	@make lint-config

.PHONY: f
f: ## Format code
	@make format

.PHONY: c
c: ## Clean build artifacts
	@make clean

.PHONY: b
b: ## Build application
	@make build

.PHONY: i
i: ## Install application
	@make install
