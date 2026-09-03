.PHONY: help build run dev test lint fmt docker-up docker-down clean frontend-build frontend-dev seed db-create migrate seed-ref seed-dev world reset-dev test-db test-db-clean test-unit-only test-integration snapshot-save snapshot-load

.DEFAULT_GOAL := help

help:
	@echo "Boxing Simulator Development Commands"
	@echo "====================================="
	@echo "make build     - Build the application"
	@echo "make run       - Run the application directly with Go"
	@echo "make dev       - Run with hot reload using air (requires air to be installed)"
	@echo "make docker-up - Start all services using Docker Compose"
	@echo "make docker-down - Stop all Docker services"
	@echo "make test      - Run all tests"
	@echo "make lint      - Run linters (golangci-lint)"
	@echo "make fmt       - Format code with gofmt"
	@echo "make clean     - Clean build artifacts"
	@echo "make frontend-build - Build the frontend React app"
	@echo "make frontend-dev - Start frontend development server"
	@echo "make seed      - Seed the database with sample data"
	@echo "make db-create - Create database"
	@echo "make migrate   - Run database migrations"
	@echo "make seed-ref  - Seed reference data"
	@echo "make seed-dev  - Seed development data"
	@echo "make world     - Generate complete world"
	@echo "make reset-dev - Reset and reseed for development"
	@echo "make test-db   - Verify isolated test database connectivity"
	@echo "make test-db-clean - Remove orphaned test databases"
	@echo "make test-unit-only - Run unit tests only (fast, no database)"
	@echo "make test-integration - Run integration tests (uses BOXING_DATABASE_* or TEST_DB_* config)"
	@echo "make snapshot-save - Save current simulation state"
	@echo "make snapshot-load - Load saved simulation state"

build:
	go build -o bin/boxing cmd/server/main.go

run: build
	./bin/boxing

dev:
	air

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

test:
	gotestsum ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

clean:
	rm -rf bin/
	rm -rf dist/

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

test-db: ## Verify isolated test database connectivity (uses BOXING_DATABASE_* or TEST_DB_* config)
	@echo "Verifying isolated test database connectivity..."
	psql --host $(BOXING_DATABASE_HOST) --port $(BOXING_DATABASE_PORT) -U $(BOXING_DATABASE_USER) -d postgres -c 'SELECT 1;'

test-db-clean: ## Remove all orphaned test_* databases from PostgreSQL
	@echo "Cleaning up orphaned test databases..."
	psql --host $(BOXING_DATABASE_HOST) --port $(BOXING_DATABASE_PORT) -U $(BOXING_DATABASE_USER) -d postgres -c "DROP DATABASE IF EXISTS $(SELECT datname FROM pg_database WHERE datname LIKE 'test_%');"

test-unit-only: ## Run unit tests only (fast, no database required)
	@echo "Running unit tests only..."
	gotestsum --format=short-verbose `go list ./... | grep -v 'internal/integration$'`

test-integration: ## Run integration tests with isolated database per-test (uses BOXING_DATABASE_* or TEST_DB_* config)
	@echo "Running integration tests..."
	gotestsum --format=short-verbose -- -tags=integration ./internal/integration/...

snapshot-save:
	# This would be implemented for saving simulation state
	echo "Snapshot save command - placeholder"

snapshot-load:
	# This would be implemented for loading simulation state
	echo "Snapshot load command - placeholder"
