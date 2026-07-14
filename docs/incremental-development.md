# Incremental Building and Testing in Go

## Overview

The Boxing API project is structured to support incremental building and testing of individual packages, which allows developers to build working components step by step. This document explains how to work incrementally with the codebase.

## Package Structure Overview

The project follows a layered architecture with clearly separated concerns:
- `cmd/` - Entry points for the application
- `internal/` - Core application logic organized by domain
  - `model/` - Data models and DTOs
  - `service/` - Business logic implementations  
  - `handler/` - HTTP request handlers
  - `db/` - Database operations and migrations
  - `platform/` - Platform-specific utilities (database, config, logger, redis)
  - `auth/` - Authentication logic
  - `events/` - Event handling system

## Compiling Individual Packages

Go's package system allows you to compile individual components without building the entire application:

### Compile a specific package:
```bash
# Compile an internal package
go build ./internal/model

# Compile a command (entry point)
go build ./cmd/server

# Compile with specific output name
go build -o boxing-server ./cmd/server
```

### Test individual packages:
```bash
# Run tests for a specific package
go test ./internal/model

# Run tests with verbose output
go test -v ./internal/service

# Run tests with coverage
go test -cover ./internal/db
```

## Incremental Development Workflow

1. **Start with models and services** - These are the core business logic components
2. **Test them independently** - Ensure they work before integrating with other components  
3. **Build handlers step by step** - HTTP interfaces that coordinate services
4. **Integrate database operations** - Test with real database connections
5. **Final integration testing** - Full application compilation and end-to-end tests

## Example Package-by-Package Building

### Step 1: Build models
```bash
go build ./internal/model
```

### Step 2: Build services (dependent on models)
```bash  
go build ./internal/service
```

### Step 3: Build handlers (dependent on services and models)
```bash
go build ./internal/handler
```

### Step 4: Build platform utilities
```bash
go build ./internal/platform/database
go build ./internal/platform/redis
```

### Step 5: Build command entry points
```bash
go build ./cmd/server
```

## Testing Individual Components

The project is already set up with comprehensive tests for each package:

```bash
# Run all tests in the db package
go test ./internal/db

# Run specific test function
go test -run TestGetUserByID ./internal/db

# Run tests with race detector (useful for concurrent code)
go test -race ./internal/service

# Run tests and show coverage
go test -coverprofile=coverage.out ./internal/db && go tool cover -html=coverage.out
```

## Working with Dependencies

Go's dependency management makes it easy to work incrementally:
- Each package can be built independently if dependencies are satisfied
- The `go.mod` file manages all dependencies
- Go automatically resolves and downloads required packages

## Best Practices for Incremental Development

1. **Start with the core domain logic** - Build services first, then handlers
2. **Test frequently** - Run individual package tests as you develop
3. **Use build tags** - For conditional compilation when needed  
4. **Leverage Go's tooling** - Use `go vet`, `golangci-lint` for code quality
5. **Keep packages focused** - Each package should have a single responsibility

## Running the Full Application

When ready to test everything together:
```bash
# Build the full application
make build

# Run tests for the entire project  
make test

# Run with hot reload (if using air)
make dev

# Or run directly
go run ./cmd/server
```

## Development Benefits of Incremental Approach

### 1. Faster Feedback Loop
- Quick compilation and testing of individual components
- Early detection of issues in specific packages
- Reduced time to identify and fix bugs

### 2. Better Understanding of Dependencies
- Clear visibility into how packages depend on each other
- Easier to understand the flow of data through the system
- Helps identify potential circular dependencies

### 3. Modular Testing Strategy
- Unit tests can be run in isolation for specific functionality
- Integration tests can target specific components or groups
- Easier to maintain test coverage as the codebase grows

### 4. Reduced Complexity
- Smaller, focused changes are easier to review and understand
- Allows for more controlled development process
- Makes it easier to collaborate on different parts of the system

## Troubleshooting Incremental Development

### Common Issues and Solutions

1. **Dependency Errors**
   - Ensure all required dependencies are in `go.mod`
   - Run `go mod tidy` to clean up dependencies
   - Check that package imports are correct

2. **Build Failures in Specific Packages**
   - Test packages individually to isolate the issue
   - Verify that all functions and methods have proper signatures
   - Ensure database connection is properly mocked for unit tests

3. **Test Failures**
   - Run specific tests with verbose output to see detailed failures
   - Use `go test -race` to detect concurrency issues
   - Check that test fixtures are properly set up

This approach allows you to develop, test, and build working components incrementally rather than trying to compile everything at once.