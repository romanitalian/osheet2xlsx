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

##@ Security

.PHONY: security-gosec
security-gosec: ## Run gosec security scanner
	curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.19.0
	gosec ./...

.PHONY: security-govulncheck
security-govulncheck: ## Run govulncheck for vulnerabilities
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

.PHONY: security-nancy
security-nancy: ## Run nancy for dependency vulnerabilities
	go install github.com/sonatype-nexus-community/nancy@latest
	go list -json -deps ./... | nancy sleuth

.PHONY: security-all
security-all: ## Run all security checks
	@make security-gosec
	@make security-govulncheck
	@make security-nancy

##@ Clean

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f mem.prof
	rm -f security-report.json

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

##@ CI/CD

.PHONY: ci-test
ci-test: ## Run tests for CI
	go test -v ./internal/... ./cmd/...

.PHONY: ci-test-race
ci-test-race: ## Run race tests for CI
	go test -race -v ./internal/... ./cmd/...

.PHONY: ci-test-coverage
ci-test-coverage: ## Run coverage tests for CI
	go test -coverprofile=coverage-report.out ./internal/... ./cmd/...

.PHONY: ci-build
ci-build: ## Build for CI
	go build -o bin/osheet2xlsx main.go

.PHONY: ci-build-race
ci-build-race: ## Build with race detector for CI
	go build -race -o bin/osheet2xlsx main.go

.PHONY: ci-install
ci-install: ## Install for CI
	go install

.PHONY: ci-lint
ci-lint: ## Run linter for CI
	golangci-lint run ./...

.PHONY: ci-examples
ci-examples: ## Build examples for CI
	@mkdir -p outputs
	go run main.go convert examples/sample.osheet --out outputs/sample.xlsx --overwrite
	go run main.go convert examples/typed.osheet --out outputs/typed.xlsx --overwrite

.PHONY: ci-examples-test
ci-examples-test: ## Test examples for CI
	go run main.go convert examples/sample.osheet --out outputs/sample_test.xlsx --overwrite
	go run main.go convert examples/typed.osheet --out outputs/typed_test.xlsx --overwrite

.PHONY: ci-examples-clean
ci-examples-clean: ## Clean examples for CI
	rm -rf outputs/

##@ Homebrew (local)

.PHONY: brew-install-local
brew-install-local: ## Install via Homebrew from local Formula (HEAD, build-from-source)
	brew install --HEAD --build-from-source ./Formula/osheet2xlsx.rb | cat

.PHONY: brew-reinstall-local
brew-reinstall-local: ## Reinstall via Homebrew from local Formula (HEAD)
	brew reinstall --HEAD --build-from-source ./Formula/osheet2xlsx.rb | cat

.PHONY: brew-uninstall
brew-uninstall: ## Uninstall Homebrew formula
	-brew uninstall osheet2xlsx | cat

.PHONY: brew-audit-local
brew-audit-local: ## Audit local formula (strict)
	brew audit --strict --online ./Formula/osheet2xlsx.rb | cat

.PHONY: brew-test
brew-test: ## Run brew test block for the formula
	brew test osheet2xlsx | cat

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

.PHONY: s
s: ## Run all security checks
	@make security-all

