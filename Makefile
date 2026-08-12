.PHONY: help build run dev test lint fmt docker-up docker-down clean frontend-build frontend-dev seed db-create migrate seed-ref seed-dev world reset-dev test-db snapshot-save snapshot-load

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
	@echo "make test-db   - Setup isolated test database"
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

test-db:
	# This would be implemented to create isolated test database
	echo "Test database setup command - placeholder"

snapshot-save:
	# This would be implemented for saving simulation state
	echo "Snapshot save command - placeholder"

snapshot-load:
	# This would be implemented for loading simulation state
	echo "Snapshot load command - placeholder"