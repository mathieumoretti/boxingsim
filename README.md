# Boxing API

A Go-based REST API backend.

## Setup

### Prerequisites

- Go 1.21+
- Docker & Docker Compose

### Quick Start

**Using Docker:**
```bash
make docker-up
```

**Manual:**
```bash
make build
make run
```

## Development

**Run with hot reload (requires `air`):**
```bash
make dev
```

**Run tests:**
```bash
make test
```

**Clean build:**
```bash
make clean && make build
```

## Docker Services

- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Server API: `localhost:8080`

## Services Management

```bash
make docker-up       # Start all services
make docker-down     # Stop all services
make docker-logs     # View logs
```

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/*` - API v1 endpoints (to be implemented)