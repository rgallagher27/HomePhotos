.PHONY: setup generate generate/rest generate/sqlite test lint migration/create

##@ Setup

setup: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/air-verse/air@latest
	@echo "All tools installed."

##@ Code Generation

generate: generate/rest generate/sqlite ## Generate all code

generate/rest: ## Generate REST server from OpenAPI
	oapi-codegen -config=./ports/rest/oapi-codegen.yml ../openapi.yaml

generate/sqlite: ## Generate sqlc code
	@cd database/sqlite && sqlc generate

##@ Database

migration/create: ## Create new migration (name=MigrationName)
ifndef name
	$(error name is required)
endif
	@name_snake=$$(echo '$(name)' | sed -e 's/\([a-z0-9]\)\([A-Z]\)/\1_\2/g' | tr '[:upper:]' '[:lower:]'); \
	migrate create -ext sql -dir ./database/sqlite/migrations -seq $$name_snake

##@ Testing

test: ## Run all tests
	go test -v ./... -short

##@ Linting

lint: ## Run linters
	go vet ./...
