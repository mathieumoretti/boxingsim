# Configuration Files

This directory contains environment-specific configuration files for the Boxing Simulation application.

## File Structure

```
config/
├── development.yaml  # Local development defaults
├── test.yaml        # Test configuration
└── production.yaml  # Production defaults
```

## How Configuration Works

Configuration is loaded in the following order (later sources override earlier ones):

1. **Default values** in code (`config.go:applyDefaults()`)
2. **Config file** based on `BOXING_ENV` environment variable
3. **Environment variables** with `BOXING_` prefix

## Environment-Specific Configs

### Development (`development.yaml`)

- Used when `BOXING_ENV=development` or not set (default)
- Debug logging enabled
- PostgreSQL on port 5433 (Docker-mapped)
- Database name: `boxing_dev`

### Test (`test.yaml`)

- Used when `BOXING_ENV=test`
- Warning-level logging
- Separate Redis port (6380) to avoid polluting dev instance
- Test JWT secret

### Production (`production.yaml`)

- Used when `BOXING_ENV=production`
- Info-level logging
- Server binds to `0.0.0.0` for container networking
- Database name: `boxing_prod`

## Environment Variable Overrides

All configuration values can be overridden using environment variables with the `BOXING_` prefix:

```bash
# Server
BOXING_SERVER_PORT=8080
BOXING_SERVER_HOST=localhost

# Database
BOXING_DATABASE_HOST=localhost
BOXING_DATABASE_PORT=5433
BOXING_DATABASE_USER=postgres
BOXING_DATABASE_PASSWORD=your-password-here
BOXING_DATABASE_NAME=boxing_dev

# Redis
BOXING_REDIS_ADDR=localhost:6379
BOXING_REDIS_PASSWORD=  # Optional

# JWT (REQUIRED for authentication)
BOXING_JWT_SECRET=your-secret-key-here

# Logging
BOXING_LOGGING_LEVEL=debug
```

## Local Development Setup

1. Copy the example environment file:
   ```bash
   cp .env.example .env.local
   ```

2. Edit `.env.local` and set your secrets (database password, JWT secret)

3. Start your services (PostgreSQL, Redis) via Docker or locally

4. Run the application:
   ```bash
   make run
   # or
   go run cmd/server/main.go
   ```

## Security Notes

- **Never commit** `.env.local` or files with actual secrets
- Config files only contain **non-secret defaults** (hosts, ports, database names)
- All **secrets must come from environment variables**:
  - Database passwords
  - JWT secrets
  - Redis passwords
  - API keys

## Example: Override Just a Few Values

You can load the development config and override specific values:

```bash
BOXING_ENV=development BOXING_LOGGING_LEVEL=info make run
```

This loads `config/development.yaml` but sets logging to "info" instead of "debug".
