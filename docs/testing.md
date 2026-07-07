# Testing

This project uses PostgreSQL for all testing environments to ensure consistency between development, CI, and production.

## Running Tests

### Local Development

For local testing with PostgreSQL:

1. Start the test database services:
```bash
docker-compose -f docker-compose.test.yml up -d
```

2. Run tests:
```bash
make test-with-postgres
```

Or directly with environment variables:
```bash
DB_HOST=localhost DB_PORT=5433 DB_USER=test_user DB_PASSWORD=test_password DB_NAME=test_boxing go test ./...
```

### GitHub Actions

Tests automatically run in GitHub Actions using Docker Compose services for PostgreSQL and Redis.

## Test Environment Setup

The tests expect the following environment variables:

- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)  
- `DB_USER`: Database user (default: test_user)
- `DB_PASSWORD`: Database password (default: test_password)
- `DB_NAME`: Database name (default: test_boxing)
- `REDIS_ADDR`: Redis address (default: localhost:6379)
- `JWT_SECRET`: JWT secret for authentication (default: test-secret-key-change-in-production)

## Notes

All database tests now use PostgreSQL instead of SQLite to avoid CGO dependencies and ensure consistency with the production environment.