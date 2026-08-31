# Integration Testing Guide

This guide covers the isolated test database infrastructure for running integration tests safely and efficiently.

## Overview

Integration tests in this project use **isolated PostgreSQL databases** that are:
- Created fresh for each test
- Automatically cleaned up after test completion
- Fully independent from development/production databases
- Named with crypto-random suffixes for uniqueness

## Setup

### Prerequisites

1. **PostgreSQL Instance**: You need a dedicated PostgreSQL instance for integration tests
   - Can be local Docker container, separate server, or CI database
   - Must allow creating/dropping databases (superuser privileges)

2. **Environment Variables**: Set the test database configuration:

```bash
# OPTION 1: Use main DATABASE config (recommended for local development)
# Tests will use BOXING_DATABASE_* env vars or config/development.yaml
# No additional setup needed!

# OPTION 2: Isolated test database (for CI/CD or separate test server)
# If ANY TEST_DB_* var is set, ALL must be set:
export TEST_DB_HOST="localhost"      # PostgreSQL host for tests
export TEST_DB_PORT="5433"           # PostgreSQL port  
export TEST_DB_USER="postgres"       # Database user with create/drop privileges
export TEST_DB_PASSWORD="password"   # Database password

# OPTIONAL - Override migrations directory (must be absolute path)
export TEST_MIGRATIONS_DIR="/absolute/path/to/db/migrations"
```

**Default Behavior**: Tests use the same database configuration as your development environment (from `config/development.yaml` or `BOXING_DATABASE_*` env vars). Each test still gets its own isolated database that's cleaned up after completion.

### Docker Setup (Recommended for Local Development)

```bash
# Start a dedicated test PostgreSQL instance
docker run -d --name boxing-test-db \
  -e POSTGRES_PASSWORD=testpassword123 \
  -p 5433:5432 \
  postgres:16-alpine

# Set environment variables
export TEST_DB_HOST="localhost"
export TEST_DB_PORT="5433"
export TEST_DB_USER="postgres"
export TEST_DB_PASSWORD="testpassword123"
```

### Verification

Before running tests, verify connectivity:

```bash
make test-db
```

Expected output: `OK: PostgreSQL connection successful`

## Running Tests

### Run All Integration Tests

```bash
make test-integration
```

This runs all tests in `internal/integration/` with isolated databases.

### Run Specific Integration Test

```bash
# Run specific test function
go test -v ./internal/integration/... -run TestDBConnection

# Run with timeout override (default: 10m per package)
go test -timeout=30m ./internal/integration/...
```

### Run Unit Tests Only (No Database Required)

```bash
make test-unit-only
```

This excludes `internal/integration/` and runs only fast unit tests.

### Run All Tests

```bash
make test
```

Runs both unit and integration tests.

## How It Works

### Database Lifecycle

1. **Test Start**: `FreshDatabaseWithMigrations()` is called
2. **Create Isolated DB**: New database with name like `test_integration_dbconn_a3f8b2c9d1`
3. **Run Migrations**: Apply schema from `db/migrations/`
4. **Execute Test**: Test runs against isolated database
5. **Auto Cleanup**: `t.Cleanup()` drops database when test completes

### Database Naming

Databases use crypto-random naming for uniqueness:
- Format: `test_<prefix>_<10-hex-chars>`
- Example: `test_integration_dbconn_f2a9b3c8d1`
- 64 bits of entropy prevents collisions in concurrent tests

### Safety Checks

The test helper includes critical safety checks:

1. **Required Configuration**: Fails fast if `TEST_DB_HOST` not set (prevents connecting to wrong server)
2. **Database Verification**: Queries `current_database()` to verify connection matches expected database
3. **Migration Path Resolution**: Uses `GetRepoRoot()` to find migrations regardless of CWD during test execution

## Test Helper Functions

### FreshDatabaseWithMigrations

Creates isolated database with migrations applied:

```go
func TestExample(t *testing.T) {
    db, cleanup := db.FreshDatabaseWithMigrations(t, "example", true)
    if db == nil {
        t.Skip("PostgreSQL not available")
    }
    defer cleanup()
    
    // Use db for tests
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
    assert.NoError(t, err)
}
```

### FreshDatabaseWithoutMigrations

Creates isolated database without migrations (for manual schema setup):

```go
func TestMigrationBehavior(t *testing.T) {
    db, cleanup := db.FreshDatabaseWithoutMigrations(t, "migration_test")
    defer cleanup()
    
    // Manually run InitializeSchema or test migration behavior
    err := db.InitializeSchema()
    assert.NoError(t, err)
}
```

## Migration Path Resolution

### Problem

During test execution, Go may change the working directory, causing relative paths like `db/migrations` to fail.

### Solution

The `GetMigrationsDir()` function:
1. Checks `TEST_MIGRATIONS_DIR` env var (if absolute path)
2. Walks up directory tree from CWD looking for `.git` or `go.mod`
3. Returns absolute path: `<repo-root>/db/migrations`
4. Falls back to relative path if repo root not found

This ensures migrations are found regardless of where tests run from.

## Cleanup

### Automatic Cleanup

Databases are automatically dropped when tests complete via `t.Cleanup()`. No manual cleanup needed.

### Manual Cleanup (Orphaned Databases)

If tests crash or leave orphaned databases:

```bash
# Remove all test_* databases
make test-db-clean
```

This safely removes all databases matching `test_%` pattern.

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Setup PostgreSQL
  uses: docker/setup-buildx-action@v3
  with:
    host: localhost
    port: 5433

- name: Run Integration Tests
  env:
    TEST_DB_HOST: localhost
    TEST_DB_PORT: 5433
    TEST_DB_USER: postgres
    TEST_DB_PASSWORD: ${{ secrets.TEST_DB_PASSWORD }}
  run: |
    make test-integration
```

## Troubleshooting

### "failed to load main config for tests"

**Cause**: Missing required `BOXING_JWT_SECRET` or invalid config  
**Fix**: 
1. Check `.env.local` has all required variables (see `.env.example`)
2. Verify `config/development.yaml` exists and is valid YAML

### "failed to connect to server"

**Cause**: PostgreSQL not running or wrong credentials  
**Fix**: 
1. Start PostgreSQL: `make docker-up`
2. Check credentials in `BOXING_DATABASE_*` env vars or `config/development.yaml`
3. Verify port is correct (Docker typically uses 5433, not 5432)

### "insufficient privilege to create database"

**Cause**: Database user lacks CREATEDB privilege  
**Fix**: Your dev database user needs superuser or CREATEDB privilege to create test databases

### "migrations failed: directory not found"

**Cause**: Migration path resolution failed  
**Fix**: 
1. Ensure running from within repository
2. Set `TEST_MIGRATIONS_DIR` to absolute path if needed
3. Verify `db/migrations/` exists

### Tests Leave Orphaned Databases

**Cause**: Test crash before cleanup runs  
**Fix**: Run `make test-db-clean` to remove all `test_%` databases

## Best Practices

1. **Use isolated databases**: Each test creates a fresh database that's cleaned up automatically
2. **For CI/CD**: Set `TEST_DB_*` env vars to use a separate test PostgreSQL instance
3. **Clean up regularly**: Run `make test-db-clean` if tests crash and leave orphaned databases
4. **Use descriptive prefixes**: Database names include test prefix for easy identification
5. **Skip gracefully**: Tests skip when PostgreSQL unavailable (check BOXING_DATABASE_* config)

## Security Notes

- **Main Config Fallback**: By default, tests use the same database as your dev environment
- **For CI/CD**: Set `TEST_DB_*` env vars to isolate test database from main development DB
- **Database Isolation**: Each test gets its own isolated database (e.g., `test_integration_dbconn_a3f8b2c9d1`)
- **Auto Cleanup**: Test databases are automatically dropped when tests complete
- **Production Safety**: Tests use `test_%` prefix - they won't accidentally touch production databases
