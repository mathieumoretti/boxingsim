.PHONY: build test ci clean

# Build targets
build: ## Build the application
	@echo "Building application..."
	go build -o boxing-server ./cmd/server

test: ## Run tests
	@echo "Running all tests with gotestsum..."
	gotestsum --format=short-verbose ./...

ci: ## Run CI pipeline (lint, build, test)
	@echo "Running CI pipeline..."
	go vet ./...
	golangci-lint run ./...
	$(MAKE) build
	$(MAKE) test

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f ./boxing-server

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

# Run development server
dev: ## Run development server
	@echo "Building and running development server..."
	$(MAKE) build
	@echo "Starting server on port 8080..."
	@./boxing-server

web-dev: ## Run the web UI server separately
	@echo "Starting web UI server on port 8081..."
	@go run web-server.go