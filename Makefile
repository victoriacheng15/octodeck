.DEFAULT_GOAL := help

.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Go

.PHONY: lint-go vet format-go test
lint-go: ## Run Go static analysis (go vet)
	go vet ./...

vet: lint-go ## Alias for lint-go

format-go: ## Format Go source code
	gofmt -s -w .

test: ## Run tests
	go test -v -race ./...

##@ Markdown

.PHONY: lint-md format-md
lint-md: ## Lint Markdown files
	npx -y markdownlint-cli "**/*.md"

format-md: ## Format Markdown files with auto-fix
	npx -y markdownlint-cli "**/*.md" --fix

##@ General

.PHONY: lint format
lint: lint-go lint-md ## Run all linters

format: format-go format-md ## Run all formatters


