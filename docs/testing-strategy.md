# BoxingSim Unit Testing Plan

## Overview
This document outlines a systematic approach to adding unit tests package by package with proper mocking for complex dependencies.

## Testing Strategy
- Use dependency injection patterns where possible (interfaces)
- Mock external dependencies like database connections using interfaces
- Focus on business logic and service layer testing
- Keep integration tests separate for complex scenarios

## Package-by-Package Testing Checklist

### 1. `internal/boxer` - Boxer Business Logic
**Status: [x]**
- [x] Create mock repository interface for testing
- [x] Test `CreateBoxer` method with various inputs and error conditions
- [x] Test `GetBoxer` method 
- [x] Test `GetBoxersByUser` method
- [x] Test `UpdateBoxer` method with partial updates
- [x] Test error handling scenarios

### 2. `internal/service` - Service Layer
**Status: [x]**
- [x] Test service wrapper methods (should be minimal)
- [x] Test proper delegation to boxer package
- [x] Test error propagation

### 3. `internal/db` - Database Operations
**Status: [ ]**
- [ ] Test database operations with in-memory SQLite
- [ ] Test error conditions (not found, duplicates)
- [ ] Test transaction handling where applicable
- [ ] Test schema initialization

### 4. `internal/model` - Data Models
**Status: [x]**
- [x] Test model creation and validation if any
- [x] Test JSON marshaling/unmarshaling
- [x] Test model methods if any

### 5. `internal/handler` - HTTP Handlers
**Status: [ ]**
- [ ] Test handler logic with mocked services
- [ ] Test error responses
- [ ] Test request parsing and validation
- [ ] Test response formatting

### 6. `internal/auth` - Authentication
**Status: [x]**
- [x] Test authentication logic
- [x] Test token generation/verification
- [x] Test middleware behavior
- [x] Test error cases (invalid credentials, expired tokens)

### 7. `internal/platform/database` - Database Connection
**Status: [x]**
- [x] Test connection establishment
- [x] Test connection pool configuration
- [x] Test database operations with mocked SQL

### 8. `internal/platform/redis` - Redis Integration
**Status: [x]**
- [x] Test Redis connection
- [x] Test basic Redis operations (set/get)
- [x] Test error handling for Redis failures

### 9. `internal/platform/config` - Configuration Management
**Status: [x]**
- [x] Test configuration loading from environment
- [x] Test default values
- [x] Test validation logic if any

### 10. `internal/events` - Event System
**Status: [ ]**
- [ ] Test event publishing
- [ ] Test event handling
- [ ] Test error propagation in events

## Testing Approach Details

### For Packages with Database Dependencies:
1. Create repository interfaces (already partially done)
2. Use dependency injection to inject mocked repositories
3. Mock the database connection at the SQL level or use in-memory SQLite
4. Test both success and failure scenarios

### For Service Layer:
1. Focus on business logic validation
2. Test method delegation to repository
3. Test error handling and propagation

### For Handlers:
1. Use mocked service dependencies
2. Test HTTP status codes and response bodies
3. Test middleware behavior
4. Test request parsing/validation

## Tools and Techniques:
- Use Go's built-in testing framework
- Use `gomock` or similar for mocking interfaces
- Use in-memory databases (SQLite) for integration-like tests
- Test both positive and negative cases
- Use table-driven tests where appropriate

## Implementation Steps:
1. Start with `internal/boxer` package
2. Create mock repository for testing
3. Add comprehensive unit tests for boxer business logic
4. Repeat for each package in the checklist
5. Ensure proper error handling coverage
6. Document test cases and expected behavior