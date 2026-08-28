# Integration Tests

This directory contains integration tests for the Boxing API system. These tests verify that different components work together correctly.

## Test Categories

### Database Tests
- Database connection validation
- Schema initialization
- User creation and retrieval

### Authentication Tests
- Password hashing and verification
- JWT token generation and validation
- Complete authentication flow

### End-to-End Tests
- Complete system flow from authentication to database operations

## Running Integration Tests

To run integration tests:

```bash
# Run all integration tests
go test ./internal/integration/ -v

# Run specific integration test
go test ./internal/integration/ -run TestAuthIntegration -v

# Run with parallel execution
go test ./internal/integration/ -parallel 4 -v
```

## Environment Setup

Integration tests require:
1. PostgreSQL database running locally or accessible via environment variables
2. Proper configuration in `.env` file

Example `.env` configuration:
```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=boxing_test
DB_USER=testuser
DB_PASSWORD=testpassword
JWT_SECRET=test-secret-key-for-integration
```

## Test Structure

Tests are organized by functionality:
- `TestDBConnection`: Validates database connectivity
- `TestAuthIntegration`: Tests complete authentication flow
- `TestDatabaseOperations`: Tests database operations with actual tables
- `TestCompleteFlow`: Tests complete system integration