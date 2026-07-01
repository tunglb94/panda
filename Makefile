## FAIRRIDE — Root Makefile
## Usage: make <target>
## Run `make help` for available targets.

SHELL := /bin/bash
.DEFAULT_GOAL := help

BACKEND_DIR := backend
INFRA_DIR   := infra/docker
SERVICES    := identity user driver trip dispatch geo pricing wallet payment promotion notification review analytics admin

# ─── Colours ─────────────────────────────────────────────────────────────────
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RESET  := \033[0m

## ──────────────────────────────────────────────────────────────────────────────
## Development environment
## ──────────────────────────────────────────────────────────────────────────────

.PHONY: infra-up
infra-up: ## Start PostgreSQL, Redis, and Kafka (Docker Compose)
	@echo -e "$(GREEN)Starting infrastructure...$(RESET)"
	@cd $(INFRA_DIR) && docker compose --env-file .env up -d
	@echo -e "$(GREEN)Infrastructure is up.$(RESET)"

.PHONY: infra-down
infra-down: ## Stop infrastructure containers
	@cd $(INFRA_DIR) && docker compose down

.PHONY: infra-logs
infra-logs: ## Tail infrastructure logs
	@cd $(INFRA_DIR) && docker compose logs -f

.PHONY: infra-reset
infra-reset: ## Stop infrastructure and DELETE all data volumes
	@echo -e "$(YELLOW)WARNING: This will delete all local data volumes.$(RESET)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	@cd $(INFRA_DIR) && docker compose down -v

## ──────────────────────────────────────────────────────────────────────────────
## Go workspace
## ──────────────────────────────────────────────────────────────────────────────

.PHONY: deps
deps: ## Download and tidy all Go dependencies
	@echo -e "$(GREEN)Tidying Go modules...$(RESET)"
	@cd $(BACKEND_DIR)/shared && go mod tidy
	@for svc in $(SERVICES); do \
		echo "  → $$svc"; \
		cd $(CURDIR)/$(BACKEND_DIR)/services/$$svc && go mod tidy; \
	done
	@cd $(BACKEND_DIR) && go work sync
	@echo -e "$(GREEN)Done.$(RESET)"

.PHONY: build
build: ## Build all services
	@echo -e "$(GREEN)Building all services...$(RESET)"
	@cd $(BACKEND_DIR) && go build github.com/fairride/shared/...
	@for svc in $(SERVICES); do \
		echo "  → $$svc"; \
		go build github.com/fairride/$$svc/...; \
	done
	@echo -e "$(GREEN)Build successful.$(RESET)"

.PHONY: build-svc
build-svc: ## Build a single service (usage: make build-svc SVC=identity)
	@if [ -z "$(SVC)" ]; then echo "Usage: make build-svc SVC=<service-name>"; exit 1; fi
	@cd $(BACKEND_DIR) && go build github.com/fairride/$(SVC)/...

.PHONY: test
test: ## Run all unit tests
	@echo -e "$(GREEN)Running tests...$(RESET)"
	@cd $(BACKEND_DIR)/shared && go test -race -count=1 ./...
	@for svc in $(SERVICES); do \
		cd $(CURDIR)/$(BACKEND_DIR)/services/$$svc && go test -race -count=1 ./... 2>/dev/null || true; \
	done

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	@cd $(BACKEND_DIR)/shared && go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@cd $(BACKEND_DIR)/shared && go tool cover -html=coverage.out -o coverage.html
	@echo -e "$(GREEN)Coverage report: backend/shared/coverage.html$(RESET)"

.PHONY: test-svc
test-svc: ## Test a single service (usage: make test-svc SVC=identity)
	@if [ -z "$(SVC)" ]; then echo "Usage: make test-svc SVC=<service-name>"; exit 1; fi
	@cd $(BACKEND_DIR)/services/$(SVC) && go test -race -v ./...

.PHONY: lint
lint: ## Run golangci-lint across the backend workspace
	@cd $(BACKEND_DIR)/shared && golangci-lint run ./...
	@for svc in $(SERVICES); do \
		cd $(CURDIR)/$(BACKEND_DIR)/services/$$svc && golangci-lint run ./... 2>/dev/null || true; \
	done

.PHONY: vet
vet: ## Run go vet
	@cd $(BACKEND_DIR)/shared && go vet ./...
	@for svc in $(SERVICES); do \
		cd $(CURDIR)/$(BACKEND_DIR)/services/$$svc && go vet ./...; \
	done

## ──────────────────────────────────────────────────────────────────────────────
## Run services locally
## ──────────────────────────────────────────────────────────────────────────────

.PHONY: run
run: ## Run a service (usage: make run SVC=identity)
	@if [ -z "$(SVC)" ]; then echo "Usage: make run SVC=<service-name>"; exit 1; fi
	@cd $(BACKEND_DIR) && go run ./services/$(SVC)/cmd/server

## Convenience targets — one per service
define RUN_TARGET
.PHONY: run-$(1)
run-$(1): ## Run the $(1) service
	@cd $(BACKEND_DIR) && go run ./services/$(1)/cmd/server
endef
$(foreach svc,$(SERVICES),$(eval $(call RUN_TARGET,$(svc))))

## ──────────────────────────────────────────────────────────────────────────────
## Code generation
## ──────────────────────────────────────────────────────────────────────────────

.PHONY: generate
generate: ## Run go generate across the entire workspace
	@cd $(BACKEND_DIR) && go generate ./...

## ──────────────────────────────────────────────────────────────────────────────
## Utilities
## ──────────────────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove build artifacts and coverage files
	@rm -rf $(BACKEND_DIR)/bin $(BACKEND_DIR)/coverage.out $(BACKEND_DIR)/coverage.html
	@find . -name '*.test' -delete

.PHONY: check
check: vet test ## Run vet + tests (fast pre-commit check)

.PHONY: help
help: ## Print this help message
	@echo ""
	@echo "FAIRRIDE — available make targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
