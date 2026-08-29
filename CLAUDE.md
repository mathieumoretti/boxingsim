# Boxing Simulation Project

## Overview
A Go-based boxing simulation game with REST API backend and React frontend. The project implements boxer management, authentication, fight mechanics, and tournament seeding systems.

## Repository Structure

```
.
├── adr/                    # Architecture Decision Records
│   ├── 001-technology-stack.md
│   ├── 002-database-design.md
│   ├── 003-authentication-system.md
│   ├── 004-web-ui-integration.md
│   ├── 005-seeding-system.md
│   ├── 006-architecture-patterns.md
│   ├── 007-docker-deployment.md
│   └── 008-viper-configuration-management.md
├── cmd/
│   ├── seed/               # Seeding functionality for tournament setup
│   └── server/             # Main server binary
├── config/                 # Environment-specific YAML configs
│   ├── development.yaml
│   ├── test.yaml
│   └── production.yaml
├── docs/                   # Documentation
│   ├── configuration.md    # Configuration guide
│   ├── database-design.md
│   ├── testing-strategy.md
│   └── ...
├── internal/
│   ├── boxer/              # Boxer entity and domain logic
│   ├── auth/               # Authentication handlers
│   ├── database/           # Database connection utilities
│   ├── handler/            # HTTP request handlers
│   ├── model/              # Data models and DTOs
│   ├── platform/           # Platform abstractions
│   │   ├── config/         # Viper configuration management
│   │   ├── database/       # Database utilities
│   │   ├── logger/         # Logging abstraction
│   │   └── redis/          # Redis client utilities
│   ├── seeding/            # Tournament seeding logic
│   ├── service/            # Business logic services
│   └── store/              # Data access layer / repositories
├── frontend/               # React frontend application
├── public/                 # Static assets served by backend
└── web/                    # Legacy web UI (deprecated)
```

## Key Components

### Configuration Management (Viper)
- **Package**: `internal/platform/config`
- **Pattern**: Strongly-typed structs with multi-source loading
- **Priority Chain**: Environment variables → YAML files → Defaults
- **Environment Prefix**: All env vars use `BOXING_` prefix
- **Files**:
  - `.env.example` - Template (committed)
  - `.env.local` - Local secrets (gitignored)
  - `config/{env}.yaml` - Environment-specific defaults

**Key Concepts**:
```go
// Config loading priority:
// 1. BOXING_DATABASE_PORT=9999 (env var) → uses 9999
// 2. development.yaml has port: 5433 → uses 5433 if no env var
// 3. applyDefaults() sets port: 5432 → final fallback
```

### Seeding System
- **Location**: `cmd/seed/main.go` and `internal/seeding/`
- **Purpose**: Organize boxers into tournament bracket structure
- **Dependencies**: Uses boxer service for data management

### Boxer Management
- **Package**: `internal/boxer/` - Entity and domain logic
- **Service**: `internal/service/` - Business operations
- **Store**: `internal/store/` - Data access layer
- **Features**: CRUD operations, stats management (strength, defense, agility)

### Authentication
- **Type**: JWT-based authentication
- **Package**: `internal/auth/`
- **Endpoints**: `/auth/register`, `/auth/login`
- **Secret Management**: JWT secret from `BOXING_JWT_SECRET` env var (REQUIRED)

### Server Components
- **Binary**: `cmd/server/boxing-server`
- **Port**: Configurable via `BOXING_SERVER_PORT` (default: 8080)
- **Features**: REST API + static file serving for frontend

### Database
- **Type**: PostgreSQL 15
- **Migrations**: Stored in `db/migrations/`
- **Local Dev**: Docker Compose on port 5433 (mapped from container's 5432)
- **Connection**: Configured via `BOXING_DATABASE_*` environment variables

## Development Workflow

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Node.js 18+ (for frontend)
- `make` utility

### Initial Setup

```bash
# 1. Start dependencies
make docker-up

# 2. Set up configuration
cp .env.example .env.local
# Edit .env.local with your secrets:
#   - BOXING_JWT_SECRET (generate: openssl rand -base64 32)
#   - BOXING_DATABASE_PASSWORD

# 3. Run migrations
make migrate

# 4. Seed data (optional)
make seed-ref    # Reference data
make seed-dev    # Development data
```

### Available Make Targets

| Command | Description |
|---------|-------------|
| `make build` | Build the application |
| `make run` | Run the built application |
| `make dev` | Run with hot reload (air) |
| `make test` | Run all tests (builds + lints first) |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with gofmt |
| `make migrate` | Run database migrations |
| `make seed-ref` | Seed reference data (static game data) |
| `make seed-dev` | Seed development data (fake data) |
| `make world` | Generate complete boxing world |
| `make reset-dev` | Reset and reseed for development |
| `make docker-up` | Start PostgreSQL and Redis containers |
| `make docker-down` | Stop Docker containers |

### Frontend Development

```bash
# Install dependencies
npm install

# Development server (hot reload)
npm start

# Production build
npm run build
```

## Testing Strategy

### Test Types
- **Unit Tests**: Pure function tests without external dependencies
- **Integration Tests**: Real database connections with isolated test database
- **Configuration Tests**: Verify config loading from different sources

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/platform/config/...

# Integration tests (requires TEST_DB_* env vars)
go test ./internal/integration/...
```

### Test Database Configuration

Integration tests use separate `TEST_DB_*` environment variables to prevent production database access:

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5433
export TEST_DB_USER=testuser
export TEST_DB_PASSWORD=testpass123
```

**Safety**: Tests fail with clear error if any `TEST_DB_*` variable is missing.

## Configuration Details

### Environment Variables (BOXING_ prefix)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOXING_ENV` | No | `development` | Loads corresponding YAML file |
| `BOXING_SERVER_PORT` | No | `8080` | HTTP server port |
| `BOXING_DATABASE_HOST` | No | `localhost` | PostgreSQL host |
| `BOXING_DATABASE_PORT` | No | `5432` | PostgreSQL port (use 5433 for Docker) |
| `BOXING_DATABASE_USER` | No | `boxing` | Database username |
| `BOXING_DATABASE_PASSWORD` | **Yes** | `boxing123`⚠️ | **Change in production!** |
| `BOXING_REDIS_ADDR` | No | `localhost:6379` | Redis address |
| `BOXING_JWT_SECRET` | **Yes** | *placeholder* | **Required for auth** |
| `BOXING_LOGGING_LEVEL` | No | `info` | debug/info/warn/error |

### Secrets Management

- **NEVER commit secrets**: `.env`, `.env.local`, `*.env` are gitignored
- **Template file**: `.env.example` is committed with placeholder values
- **Generate secure secrets**:
  ```bash
  openssl rand -base64 32  # JWT secret
  openssl rand -base64 16  # Database password
  ```

**Full Guide**: [docs/configuration.md](docs/configuration.md)

## Documentation

- **[docs/configuration.md](docs/configuration.md)** - Complete configuration guide
- **[docs/database-design.md](docs/database-design.md)** - Schema and ERD
- **[docs/testing-strategy.md](docs/testing-strategy.md)** - Testing approach
- **[adr/](adr/)** - Architecture Decision Records
  - [008-viper-configuration-management.md](adr/008-viper-configuration-management.md)

## API Endpoints

### Health
- `GET /health` - Server status

### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login (returns JWT)

### Boxers
- `POST /boxers` - Create boxer
- `GET /boxers/{id}` - Get boxer details
- `PUT /boxers/{id}` - Update boxer stats
- `DELETE /boxers/{id}` - Delete boxer

## Common Issues

### "role 'boxing' does not exist"
```bash
make docker-up
docker-compose exec postgres psql -U postgres
# Then: CREATE USER boxing WITH PASSWORD 'boxing123';
```

### Configuration not loading
Check `BOXING_ENV` environment variable and ensure `config/{env}.yaml` exists.

### Test failures with "TEST_DB_HOST required"
Set the four `TEST_DB_*` environment variables before running integration tests.

## Current Status

### Completed
- ✅ Viper configuration management (MAT-62)
- ✅ Secrets management with .env.example (MAT-65)
- ✅ Configuration documentation (MAT-70)
- ✅ PostgreSQL database with migrations
- ✅ Redis caching layer
- ✅ JWT authentication
- ✅ Boxer CRUD operations
- ✅ Tournament seeding system

### In Progress
- Integration test database isolation (MAT-46) - enabled by config system

## File Path Rules
* Always use Windows-style backslashes (`\`) for file operations.
* Always use absolute paths with drive letters (e.g., `C:\path\to\project\main.go`).
* NEVER use relative paths or forward slashes.

## Go Writing Rules
* If a file write or edit fails due to string matching, do not retry with Edit tool.
* Fall back immediately to writing the file using PowerShell: `New-Item -Force` or redirect.
