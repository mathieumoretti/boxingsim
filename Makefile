.PHONY: help build run dev test lint fmt docker-up docker-down clean frontend-build frontend-dev

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