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

- Migrations are stored in migrations/
- The database connection is configured through environment variables

## Architecture

The application follows a layered architecture pattern with clear separation of concerns:
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