# Configuration Guide

Comprehensive guide for configuring the Boxing Simulation application.

## Quick Start

```bash
# 1. Copy the example environment file
cp .env.example .env.local

# 2. Edit .env.local with your secrets (JWT_SECRET, DATABASE_PASSWORD)

# 3. Start PostgreSQL via Docker
make docker-up

# 4. Run the application
make dev
```

For detailed configuration options, continue reading below.

---

## Table of Contents

1. [Configuration Hierarchy](#configuration-hierarchy)
2. [Environment Variables Reference](#environment-variables-reference)
3. [Secrets Management](#secrets-management)
4. [Config Files Structure](#config-files-structure)
5. [Integration Test Database Setup](#integration-test-database-setup)
6. [Docker Compose Configuration](#docker-compose-configuration)
7. [Troubleshooting](#troubleshooting)

---

## Configuration Hierarchy

The application loads configuration from multiple sources in the following priority order (highest to lowest):

1. **Environment Variables** - `BOXING_*` prefixed variables (highest priority)
2. **YAML Config Files** - `config/{environment}.yaml` files
3. **Default Values** - Hardcoded fallbacks in `applyDefaults()` (lowest priority)

### Loading Flow Diagram

```
┌─────────────────────────────────────────────────┐
│ 1. Viper loads YAML file                       │
│    (e.g., config/development.yaml)              │
└──────────────┬──────────────────────────────────┘
               │
               v
┌─────────────────────────────────────────────────┐
│ 2. Environment variables override YAML         │
│    (BOXING_DATABASE_HOST, etc.)                 │
└──────────────┬──────────────────────────────────┘
               │
               v
┌─────────────────────────────────────────────────┐
│ 3. applyDefaults() fills remaining empty values│
└─────────────────────────────────────────────────┘
```

### Example: How a Value Gets Resolved

```bash
# Scenario: Database port configuration

config/development.yaml has:   database.port = 5433
applyDefaults() provides:      5432 (standard PostgreSQL)
BOXING_DATABASE_PORT env var:  not set

Result: Port = 5433 (from YAML file)

---

# Change scenario: Set environment variable
export BOXING_DATABASE_PORT=5434

Result: Port = 5434 (env var overrides YAML)
```

---

## Environment Variables Reference

All application environment variables use the `BOXING_` prefix.

### Core Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOXING_ENV` | No | `development` | Environment name; determines YAML file to load (`development`, `test`, `production`) |

### Server Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOXING_SERVER_PORT` | No | `8080` | HTTP server port |
| `BOXING_SERVER_HOST` | No | `localhost` | HTTP server host binding |

### Database Configuration (PostgreSQL)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOXING_DATABASE_HOST` | No | `localhost` | PostgreSQL host address |
| `BOXING_DATABASE_PORT` | No | `5432` | PostgreSQL port (use `5433` for Docker-mapped) |
| `BOXING_DATABASE_USER` | No | `boxing` | Database username |
| `BOXING_DATABASE_PASSWORD` | **Yes** | `boxing123` ⚠️ | Database password (**change in production!**) |
| `BOXING_DATABASE_NAME` | No | `boxing` | Database name |

> ⚠️ **Security Warning**: The default password `boxing123` is only for local development convenience. Always set a strong password via `BOXING_DATABASE_PASSWORD` in `.env.local`.

### Redis Configuration (Optional)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOXING_REDIS_ADDR` | No | `localhost:6379` | Redis server address |
| `BOXING_REDIS_PASSWORD` | No | (empty) | Redis password (if authentication enabled) |

### JWT Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOXING_JWT_SECRET` | **Yes** | `default-jwt-secret-*` ⚠️ | Secret key for JWT token signing (**REQUIRED!**) |

> 🔴 **CRITICAL**: Change `BOXING_JWT_SECRET` to a secure random string in production. Generate one with:
> ```bash
> openssl rand -base64 32
> ```

### Logging Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOXING_LOGGING_LEVEL` | No | `info` | Log level (`debug`, `info`, `warn`, `error`) |

---

## Secrets Management

### Security Principles

1. **NEVER commit secrets** - `.env` and `.env.local` are in `.gitignore`
2. **Use environment variables** - All secrets come from env vars, not code defaults
3. **Strong passwords** - Use cryptographically random strings for JWT_SECRET and DATABASE_PASSWORD
4. **Different secrets per environment** - Development, staging, and production should use different secrets

### File Structure

```
.env.example    # Template with placeholder values (COMMIT THIS)
.env.local      # Your local development secrets (NEVER COMMIT)
.env            # Alternative name for .env.local (NEVER COMMIT)
```

### .gitignore Rules

The project's `.gitignore` ensures:
- ✅ `.env.example` is tracked (safe template)
- ❌ `.env`, `.env.local`, `*.env` are ignored (contain secrets)

### Generating Secure Secrets

```bash
# JWT Secret (32 bytes, base64 encoded)
openssl rand -base64 32

# Database Password (16 bytes)
openssl rand -base64 16

# Alternative using Python
python -c "import secrets; print(secrets.token_urlsafe(32))"
```

---

## Config Files Structure

YAML config files live in `config/` directory and are loaded based on `BOXING_ENV`.

### development.yaml

```yaml
database:
  host: "localhost"
  port: 5433        # Docker-mapped PostgreSQL
  user: "postgres"
  name: "boxing_dev"

server:
  port: 8080

logging:
  level: "debug"
```

**Purpose**: Local development with Docker containers

### test.yaml

```yaml
database:
  host: "localhost"
  port: 5432        # Standard PostgreSQL (not Docker-mapped)

server:
  port: 8081        # Different port to avoid conflicts

logging:
  level: "warn"     # Less verbose for tests

redis:
  addr: "localhost:6380"
```

**Purpose**: Test environment with reduced logging

### production.yaml

Create this file for production deployments with:
- Production database credentials
- Production Redis settings
- `logging.level: "error"` or `"warn"`
- Secure JWT secret (via env var only!)

---

## Integration Test Database Setup

For integration tests that require a real database connection, use the `TEST_DB_*` environment variables.

### Why Separate Test Database?

Integration tests may modify/delete data. Using a separate database prevents:
- Accidental damage to development/production data
- Test pollution from leftover data
- Race conditions when running tests in parallel

### Required Variables (NO DEFAULTS)

| Variable | Description | Example |
|----------|-------------|---------|
| `TEST_DB_HOST` | Test database host | `localhost` |
| `TEST_DB_PORT` | Test database port | `5433` |
| `TEST_DB_USER` | Test database user | `testuser` |
| `TEST_DB_PASSWORD` | Test database password | `testpass123` |

### Safety Mechanism

**Tests will fail with a clear error if any `TEST_DB_*` variable is missing.** This prevents accidental connection to the wrong database.

```bash
# Example: Running integration tests
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5433
export TEST_DB_USER=testuser
export TEST_DB_PASSWORD=testpass123

go test ./internal/integration/...
```

### Creating Test Database (PostgreSQL)

```sql
-- Connect to PostgreSQL
psql -U postgres

-- Create test user and database
CREATE USER testuser WITH PASSWORD 'testpass123';
CREATE DATABASE testdb OWNER testuser;
GRANT ALL PRIVILEGES ON DATABASE testdb TO testuser;
```

---

## Docker Compose Configuration

The project uses Docker Compose for local development dependencies.

### docker-compose.yml Overview

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    ports:
      - "5433:5432"    # Host:Container mapping
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: boxing_dev

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### Port Mapping Explained

- **Container port**: `5432` (PostgreSQL default inside container)
- **Host port**: `5433` (accessed from your machine)
- **Why different?** Prevents conflict with native PostgreSQL installations

### Docker Commands

```bash
# Start services
make docker-up

# Stop services
make docker-down

# Restart services
make docker-down && make docker-up

# View logs
docker-compose logs -f postgres

# Execute SQL in container
docker-compose exec postgres psql -U postgres -d boxing_dev
```

---

## Troubleshooting

### "role 'boxing' does not exist"

**Cause**: Database user not created or mismatch with config.

**Solution**:
```bash
# Check current config
cat .env.local | grep BOXING_DATABASE_USER

# Connect to database
docker-compose exec postgres psql -U postgres

# List users
\du

# Create missing user
CREATE USER boxing WITH PASSWORD 'boxing123';
GRANT ALL PRIVILEGES ON DATABASE boxing_dev TO boxing;
```

### Tests fail with "TEST_DB_HOST environment variable is required"

**Cause**: Integration tests need isolated database configuration.

**Solution**: Add to `.env.local`:
```bash
TEST_DB_HOST=localhost
TEST_DB_PORT=5433
TEST_DB_USER=testuser
TEST_DB_PASSWORD=testpass123
```

Then export them:
```bash
set -a
source .env.local
set +a
```

### JWT validation errors after config change

**Cause**: Old tokens signed with previous `BOXING_JWT_SECRET`.

**Solution**: Users need to re-login with new secret.

### YAML file not being loaded

**Check**:
1. File exists: `ls config/${BOXING_ENV}.yaml`
2. Correct env: `echo $BOXING_ENV` (default is `development`)
3. Valid YAML syntax: Use a YAML linter

### Port already in use

**Error**: `listen tcp :8080: bind: Address already in use`

**Solution**:
```bash
# Find process
netstat -ano | findstr :8080

# Or change port
export BOXING_SERVER_PORT=8081
```

---

## Related Documentation

- [Architecture Decision Records](../adr/) - Technical decisions and rationale
- [Development Setup](development-setup.md) - Getting started guide
- [Testing Strategy](testing-strategy.md) - Test configuration and practices
- [Database Design](database-design.md) - Schema and migrations

---

## Security Checklist

Before deploying to production:

- [ ] Changed `BOXING_JWT_SECRET` to random string
- [ ] Changed `BOXING_DATABASE_PASSWORD` to strong password
- [ ] Verified `.env.local` is not in git
- [ ] Set `BOXING_LOGGING_LEVEL` to `"warn"` or `"error"`
- [ ] Created separate production database
- [ ] Configured firewall/security groups for database access
- [ ] Enabled TLS for production PostgreSQL connection
- [ ] Set up secret management (AWS Secrets Manager, Vault, etc.)

---

## Appendix: Complete .env.example Template

```bash
# Boxing Simulation Environment Variables
# ========================================
# Copy this to .env.local and modify as needed
# NEVER commit .env or .env.local!

# Environment Selection
BOXING_ENV=development

# Server Configuration
BOXING_SERVER_PORT=8080
BOXING_SERVER_HOST=localhost

# Database Configuration (PostgreSQL)
BOXING_DATABASE_HOST=localhost
BOXING_DATABASE_PORT=5433
BOXING_DATABASE_USER=postgres
BOXING_DATABASE_NAME=boxing_dev
BOXING_DATABASE_PASSWORD=change-me-in-production

# Redis Configuration
BOXING_REDIS_ADDR=localhost:6379
# BOXING_REDIS_PASSWORD=  # Uncomment if needed

# JWT Configuration [REQUIRED]
BOXING_JWT_SECRET=generate-a-secure-random-string

# Logging
BOXING_LOGGING_LEVEL=debug

# Test Database (Integration Tests Only)
# TEST_DB_HOST=localhost
# TEST_DB_PORT=5433
# TEST_DB_USER=testuser
# TEST_DB_PASSWORD=testpassword
```
