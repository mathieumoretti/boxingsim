# Boxing Simulator

A Go-based REST API backend for a boxing simulation game with a modern React frontend.

## Project Structure

- `cmd/` - Entry points for the application
- `internal/` - Core application logic organized by domain:
  - `model/` - Data models and DTOs
  - `service/` - Business logic implementations
  - `handler/` - HTTP request handlers
  - `db/` - Database operations and migrations
  - `platform/` - Platform-specific utilities (database, config, logger, redis)
  - `auth/` - Authentication logic
  - `store/` - Data access layer for repositories
- `web/` - Legacy web UI files (deprecated, replaced by React app in `frontend/`)
- `frontend/` - Modern React frontend with Webpack bundling

## Features Implemented

### Core Components
- **Boxer Management**: Create, update, and manage boxers with stats (strength, defense, agility)
- **User Authentication**: Registration and login with JWT token generation
- **Fight System**: Fight logic between boxers with combat mechanics
- **Database Integration**: PostgreSQL storage with migrations
- **Caching Layer**: Redis for performance optimization
- **Configuration Management**: Viper with strongly-typed config structs ([docs/configuration.md](docs/configuration.md))

### API Endpoints

#### Health Check
- `GET /health` - Returns server status

#### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login

#### Boxer Operations
- `POST /boxers` - Create a new boxer
- `GET /boxers/{id}` - Get boxer details
- `PUT /boxers/{id}` - Update boxer stats
- `DELETE /boxers/{id}` - Delete a boxer

## Development Commands

### Backend Setup

```bash
# Install Go dependencies
go mod tidy

# Install air hot-reloading tool (for development)
go install github.com/air-verse/air@latest

# Start database services
make docker-up
```

### Frontend Setup

```bash
# Install npm dependencies
npm install

# Run development server (with hot reload)
npm start

# Build for production
npm run build
```

## Running the Application

### Development Mode

1. Start the backend server:
   ```bash
   make dev
   ```

2. In another terminal, start the frontend development server:
   ```bash
   npm start
   ```

3. Open your browser and navigate to `http://localhost:3000` (frontend) or `http://localhost:8080` (backend API)

### Production Mode

1. Build the frontend:
   ```bash
   npm run build
   ```

2. Start the backend server:
   ```bash
   make run
   ```

## Testing and Quality

- `make test` - Run all tests (builds and lints first)
- See [docs/testing-strategy.md](docs/testing-strategy.md) for detailed testing information

## Gotestsum Installation

To get enhanced test output with gotestsum:

```bash
go install github.com/gotesttools/gotestsum@latest
```

If you encounter git authentication issues in CI environments:
```bash
GO111MODULE=on go install github.com/gotesttools/gotestsum@latest
```

## Database Operations

- Migrations are stored in db/migrations/ directory
- The database connection is configured through environment variables
- Sample data can be seeded using `make seed` command
- Database management commands:
  - `make db-create` - Create database
  - `make migrate` - Run all migrations
  - `make seed-ref` - Seed reference data (static game data)
  - `make seed-dev` - Seed development data (fake data for dev)
  - `make world` - Generate complete boxing world
  - `make reset-dev` - Reset and reseed for development

## Database Setup Troubleshooting

If you encounter "role 'boxing' does not exist" error:

1. Ensure Docker containers are running:
   ```bash
   make docker-up
   ```

2. Check database logs:
   ```bash
   docker-compose logs boxing-postgres
   ```

3. Restart the database service:
   ```bash
   make docker-down
   make docker-up
   ```

4. Run migrations again:
   ```bash
   make migrate
   ```

## Architecture

The application follows a layered architecture pattern with clear separation of concerns:
1. **Presentation Layer**: HTTP handlers in `/handler`
2. **Business Logic Layer**: Services in `/service` 
3. **Data Access Layer**: Repositories in `/store` and `/db`
4. **Platform Layer**: Database, Redis, configuration utilities

## Configuration

The application uses Viper for configuration management with a priority hierarchy:

1. **Environment Variables** (highest priority) - `BOXING_*` prefixed variables
2. **YAML Config Files** - `config/{environment}.yaml` files  
3. **Default Values** (lowest priority) - Hardcoded fallbacks

### Quick Setup

```bash
# 1. Copy the example environment file
cp .env.example .env.local

# 2. Edit .env.local with your secrets:
#    - BOXING_JWT_SECRET (generate with: openssl rand -base64 32)
#    - BOXING_DATABASE_PASSWORD (use a strong password)

# 3. Start PostgreSQL via Docker
make docker-up

# 4. Run the application  
make dev
```

### Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOXING_ENV` | No | `development` | Environment name (loads corresponding YAML) |
| `BOXING_SERVER_PORT` | No | `8080` | HTTP server port |
| `BOXING_DATABASE_HOST` | No | `localhost` | PostgreSQL host |
| `BOXING_DATABASE_PORT` | No | `5432` | PostgreSQL port (use `5433` for Docker) |
| `BOXING_DATABASE_PASSWORD` | **Yes** | `boxing123`⚠️ | Database password (**change in prod!**) |
| `BOXING_JWT_SECRET` | **Yes** | *default*⚠️ | JWT signing key (**REQUIRED!**) |
| `BOXING_LOGGING_LEVEL` | No | `info` | Log level (debug/info/warn/error) |

⚠️ **Security Warning**: Default passwords are for development only. Always set strong secrets via `.env.local`.

**Full Documentation**: See [docs/configuration.md](docs/configuration.md) for complete configuration guide, including:
- Configuration loading hierarchy details
- Complete environment variables reference
- Secrets management best practices
- Docker Compose configuration
- Integration test database setup
- Troubleshooting guide

## Getting Started

1. Clone the repository
2. Set up environment variables in `.env`
3. Run `make docker-up` to start dependencies
4. Run `make run` to start the server

The API is now ready to handle boxing simulation requests with full CRUD operations for boxers and authentication.