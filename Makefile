.PHONY: build test ci clean

# Build targets
build: ## Build the application incrementally
	@echo "Building all packages incrementally..."
	./mini-build.sh build

test: ## Run tests
	@echo "Running all tests..."
	./mini-build.sh test

ci: ## Run CI pipeline (lint, build, test)
	@echo "Running CI pipeline..."
	./mini-build.sh ci

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	./mini-build.sh clean
	@rm -f ./boxing-server

# Individual package builds
build-model:
	go build -o build/model ./internal/model

build-service: build-model
	go build -o build/service ./internal/service

build-db: build-model
	go build -o build/db ./internal/db

build-platform: build-model
	go build -o build/platform ./internal/platform

docker-up: ## Start all services with Docker Compose
	docker-compose up -d postgres redis
	@echo "Waiting for database to be ready..."
	@sleep 5
	docker-compose up -d server
build-auth: build-model build-platform
	go build -o build/auth ./internal/auth

docker-down: ## Stop all Docker services
	docker-compose down
build-handler: build-service build-model
	go build -o build/handler ./internal/handler

docker-logs: ## View logs from all services
	docker-compose logs -f
build-server: build-auth build-db build-handler build-platform
	go build -o boxing-server ./cmd/server

# Individual package tests
test-model:
	go test ./internal/model

test-service:
	go test ./internal/service

test-db:
	go test ./internal/db

test-handler:
	go test ./internal/handler

test-auth:
	go test ./internal/auth

lint: ## Run go vet and linter
	go vet ./...
	golangci-lint run ./...

lint-dev: ## Run go vet and linter
	go vet ./...
	bin/golangci-lint.exe run ./...

deps: ## Download dependencies
	go mod download
	go mod tidy

# Run development server
dev: ## Run development server (no hot reload)
	@echo "Building and running development server..."
	@make build-server
	@echo "Starting server on port 8080..."
	@./boxing-server

web-dev: ## Run the web UI server separately
	@echo "Starting web UI server on port 8081..."
	@go run web-server.go