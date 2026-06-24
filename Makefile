.PHONY: build run dev test clean docker-up docker-down

APP_NAME=boxing
BINARY_NAME=$(APP_NAME)-server
BUILD_DIR=bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the application
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run: ## Build and run the application
	go run ./cmd/server

dev: ## Run with hot reload (requires air)
	air

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	go clean

docker-up: ## Start all services with Docker Compose
	docker-compose up -d postgres redis
	@echo "Waiting for database to be ready..."
	@sleep 5
	docker-compose up -d server

docker-down: ## Stop all Docker services
	docker-compose down

docker-logs: ## View logs from all services
	docker-compose logs -f

lint: ## Run go vet and linter
	go vet ./...
	golangci-lint run ./...

deps: ## Download dependencies
	go mod download
	go mod tidy