.PHONY: help dev-up dev-down build test clean setup-dev migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@egrep '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup-dev: ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Please edit .env file with your configuration"

dev-up: ## Start development environment
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Development environment is ready!"

dev-down: ## Stop development environment
	docker-compose -f docker-compose.dev.yml down

dev-logs: ## Show development environment logs
	docker-compose -f docker-compose.dev.yml logs -f

migrate: ## Run database migrations
	@echo "Running PostgreSQL migrations..."
	# We'll add migration tool later

build: ## Build all services
	@echo "Building all services..."
	@for service in services/*/; do \
		echo "Building $$service..."; \
		cd $$service && go build -o bin/$$(basename $$service) ./cmd/server && cd ../..; \
	done

test: ## Run tests for all services
	@echo "Running tests..."
	@go test ./...

clean: ## Clean up build artifacts
	@echo "Cleaning up..."
	@find . -name "bin" -type d -exec rm -rf {} +
	@docker-compose -f docker-compose.dev.yml down -v

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy
	@for service in services/*/; do \
		echo "Tidying $$service..."; \
		cd $$service && go mod tidy && cd ../..; \
	done