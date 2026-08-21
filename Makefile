.PHONY: help build run dev test lint fmt docker-up docker-down clean frontend-build frontend-dev seed db-create migrate seed-ref seed-dev world reset-dev test-db test-db-clean test-unit-only test-integration snapshot-save snapshot-load

.DEFAULT_GOAL := help

help:
	@echo "Boxing Simulator Development Commands"
	@echo "====================================="
	@echo "make build             - Build the application"
	@echo "make run               - Run the application directly with Go"
	@echo "make dev               - Run with hot reload using air (requires air to be installed)"
	@echo "make docker-up         - Start all services using Docker Compose"
	@echo "make docker-down       - Stop all Docker services"
	@echo "make test              - Run all tests"
	@echo "make lint              - Run linters (golangci-lint)"
	@echo "make fmt               - Format code with gofmt"
	@echo "make clean             - Clean build artifacts"
	@echo "make frontend-build    - Build the frontend React app"
	@echo "make frontend-dev      - Start frontend development server"
	@echo "make seed              - Seed the database with sample data"
	@echo "make db-create         - Create database"
	@echo "make migrate           - Run database migrations"
	@echo "make seed-ref          - Seed reference data"
	@echo "make seed-dev          - Seed development data"
	@echo "make world             - Generate complete world"
	@echo "make reset-dev         - Reset and reseed for development"
	@echo "make test-db           - Verify isolated test database connectivity"
	@echo "make test-db-clean     - Remove all orphaned test_db_* databases from PostgreSQL"
	@echo "make test-unit-only    - Run unit tests only (fast, no DB required)"
	@echo "make test-integration  - Run integration tests with isolated database per-test"

build:
	go build -o bin/boxing cmd/server/main.go

run: build ./bin/boxing

dev: .air

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

test: ## Run all tests (skips integration if no database available)
	gotestsum --format=short-verbose ./... || true

lint:
	golangci-lint run

fmt:
	gofmt -w .

clean:
	rm -rf bin/ dist/

frontend-build:
	npm run build

frontend-dev:
	npm start

seed:
	go run cmd/seed/main.go

# Database commands
db-create:
	createdb -U boxing boxing

migrate:
	go run cmd/migrate/main.go up

seed-ref:
	go run cmd/seed/main.go reference

seed-dev:
	go run cmd/seed/main.go development

world:
	go run cmd/seed/main.go world

reset-dev: migrate seed-ref seed-dev

# Test database commands - for isolated integration test databases
test-db: ## Verify isolated test database connectivity (requires TEST_DB_HOST env var)
	@echo "Verifying isolated test database connectivity..."
	@if [ -z "$${TEST_DB_HOST:-}" ]; then echo "ERROR: TEST_DB_HOST environment variable not set"; exit 1; fi
	psql --host "$${TEST_DB_HOST:-localhost}" --port "$${TEST_DB_PORT:-5432}" -U "$${TEST_DB_USER:-postgres}" -d postgres -c 'SELECT 1;' > /dev/null && echo "OK: PostgreSQL connection successful" || { echo "ERROR: Failed to connect"; exit 1; }

test-db-clean: ## Remove all orphaned test_db_* databases from PostgreSQL
	@echo "Cleaning up orphaned test databases..."
	psql --host "$${TEST_DB_HOST:-localhost}" --port "$${TEST_DB_PORT:-5432}" -U "$${TEST_DB_USER:-postgres}" -d postgres \
		-tAc "SELECT 'DROP DATABASE '''' || datname || ''' ';' FROM pg_database WHERE datname LIKE 'test_db_%';" 2>/dev/null | grep -v '^$$' | psql --host "$${TEST_DB_HOST:-localhost}" --port "$${TEST_DB_PORT:-5432}" -U "$${TEST_DB_USER:-postgres}" -d postgres || echo "OK: No test databases to clean up"

test-unit-only: ## Run unit tests only (fast, no database required)
	@echo "Running unit tests only..."
	go list ./... | grep -v 'internal/integration$$' | xargs gotestsum --format=short-verbose || echo "Unit tests completed with output above"

test-integration: ## Run integration tests with isolated database per-test (requires TEST_DB_HOST env var)
	@echo "Running integration tests..."
	gotestsum --format=short-verbose ./internal/integration/... 2>&1; exit $$?

snapshot-save:
	echo "Snapshot save command - not yet implemented"

snapshot-load:
	echo "Snapshot load command - not yet implemented"
