# Boxing Simulation Game API

A Go-based REST API backend for a boxing simulation game. The system manages boxers, fights, users, and game world state using PostgreSQL for data persistence and Redis for caching.

## Features Implemented

### Core Components
- **Boxer Management**: Create, update, and manage boxers with stats (strength, defense, agility)
- **User Authentication**: Registration and login with JWT token generation
- **Fight System**: Fight logic between boxers with combat mechanics
- **Database Integration**: PostgreSQL storage with migrations
- **Caching Layer**: Redis for performance optimization
- **Configuration Management**: Environment-based configuration

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

## Project Structure

```
cmd/              # Entry point of the application
internal/         # Core application logic organized by domain
├── model/       # Data models and DTOs for API responses, requests, and database entities
├── service/     # Business logic implementations  
├── handler/     # HTTP request handlers that coordinate between services and models
├── db/          # Database operations and migrations
├── platform/    # Platform-specific utilities (database, config, logger, redis)
├── auth/        # Authentication logic
└── store/       # Data access layer for repositories

build/            # Build artifacts and scripts
migrations/       # Database schema migrations
```

## Development Commands

- `make build` - Build the application
- `make run` - Run the application directly with Go
- `make dev` - Run with hot reload using air (requires installation)
- `make docker-up` - Start all services using Docker Compose
- `make docker-down` - Stop all Docker services
- `make lint` - Run linters (golangci-lint)
- `make fmt` - Format code with gofmt

## Testing and Quality

- `make test` - Run all tests (builds and lints first)
- See [docs/testing-strategy.md](docs/testing-strategy.md) for detailed testing information

## Database Operations

- Migrations are stored in migrations/
- The database connection is configured through environment variables

## Architecture

The system follows a layered architecture pattern:
1. **Presentation Layer**: HTTP handlers in `/handler`
2. **Business Logic Layer**: Services in `/service` 
3. **Data Access Layer**: Repositories in `/store` and `/db`
4. **Platform Layer**: Database, Redis, configuration utilities

## Getting Started

1. Clone the repository
2. Set up environment variables in `.env`
3. Run `make docker-up` to start dependencies
4. Run `make run` to start the server

The API is now ready to handle boxing simulation requests with full CRUD operations for boxers and authentication.