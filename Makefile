# HomePhotos Makefile
# Unified commands for managing the full stack

.PHONY: help setup dev build generate test lint db clean

CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RESET := \033[0m

##@ General

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(CYAN)Usage:$(RESET)\n  make $(GREEN)<target>$(RESET)\n"} /^[a-zA-Z_0-9\/-]+:.*?##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup

setup: setup/backend setup/frontend ## Install dependencies for all projects
	@echo "$(GREEN)All projects set up$(RESET)"

setup/backend: ## Install backend dependencies and tools
	@echo "$(CYAN)Setting up backend...$(RESET)"
	cd backend && $(MAKE) setup

setup/frontend: ## Install frontend dependencies
	@echo "$(CYAN)Setting up frontend...$(RESET)"
	cd frontend && npm install

##@ Development

dev: ## Run backend + frontend
	@echo "$(CYAN)Starting backend and frontend...$(RESET)"
	@./scripts/dev.sh backend frontend

dev/backend: ## Start backend with air
	@echo "$(CYAN)Starting backend...$(RESET)"
	cd backend && air

dev/frontend: ## Start frontend dev server
	@echo "$(CYAN)Starting frontend...$(RESET)"
	cd frontend && npm run dev

##@ Code Generation

generate: bundle-openapi generate/backend generate/frontend ## Generate all API code
	@echo "$(GREEN)All code generated$(RESET)"

generate/backend: ## Generate backend REST server from OpenAPI
	@echo "$(CYAN)Generating backend code...$(RESET)"
	cd backend && $(MAKE) generate

generate/frontend: ## Generate frontend API client from OpenAPI
	@echo "$(CYAN)Generating frontend code...$(RESET)"
	cd frontend && npm run api:generate

bundle-openapi: validate-openapi ## Bundle split OpenAPI spec
	@echo "$(CYAN)Bundling OpenAPI spec...$(RESET)"
	cd frontend && npm run bundle-openapi
	@echo "$(GREEN)OpenAPI spec bundled$(RESET)"

validate-openapi: ## Validate OpenAPI spec
	@echo "$(CYAN)Validating OpenAPI spec...$(RESET)"
	cd frontend && npm run validate-openapi

##@ Testing

test: test/backend test/frontend ## Run all tests
	@echo "$(GREEN)All tests passed$(RESET)"

test/backend: ## Run backend tests
	@echo "$(CYAN)Running backend tests...$(RESET)"
	cd backend && $(MAKE) test

test/frontend: ## Run frontend tests
	@echo "$(CYAN)Running frontend tests...$(RESET)"
	cd frontend && npm run check

##@ Linting

lint: lint/backend lint/frontend ## Run all linters
	@echo "$(GREEN)All linting passed$(RESET)"

lint/backend: ## Run backend linters
	@echo "$(CYAN)Linting backend...$(RESET)"
	cd backend && $(MAKE) lint

lint/frontend: ## Run frontend linter
	@echo "$(CYAN)Linting frontend...$(RESET)"
	cd frontend && npm run check

##@ Database

db/migrate: ## Run database migrations
	@echo "$(CYAN)Running migrations...$(RESET)"
	@DB_PATH=$$(grep DB_PATH backend/.env.local 2>/dev/null | cut -d'=' -f2- || echo "./backend/homephotos.db"); \
	cd backend && migrate -source "file://database/sqlite/migrations" -database "sqlite://$$DB_PATH" up

db/migrate/down: ## Revert last migration
	@DB_PATH=$$(grep DB_PATH backend/.env.local 2>/dev/null | cut -d'=' -f2- || echo "./backend/homephotos.db"); \
	cd backend && migrate -source "file://database/sqlite/migrations" -database "sqlite://$$DB_PATH" down 1

db/migrate/create: ## Create new migration (name=MigrationName)
ifndef name
	$(error name is required: make db/migrate/create name=AddUsersTable)
endif
	cd backend && $(MAKE) migration/create name=$(name)

##@ Utilities

clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning build artifacts...$(RESET)"
	rm -rf backend/tmp frontend/.svelte-kit frontend/build
	@echo "$(GREEN)Clean complete$(RESET)"
