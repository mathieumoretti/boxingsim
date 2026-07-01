# Database Design and Implementation

## Overview

This document describes the PostgreSQL database design for the Boxing Simulator backend, including schema definitions, implementation details, and key considerations.

## Database Schema

### Users Table
- `id` (UUID) - Primary key
- `username` (string) - Unique user identifier
- `email` (string) - User email address
- `password_hash` (string) - Hashed password
- `created_at` (timestamp) - Account creation time
- `updated_at` (timestamp) - Last account update

### Boxers Table
- `id` (UUID) - Primary key
- `user_id` (UUID) - Foreign key to users table
- `name` (string) - Boxer name
- `power` (integer) - Power stat (1-100)
- `speed` (integer) - Speed stat (1-100)
- `chin` (integer) - Chin stat (1-100)
- `cardio` (integer) - Cardio stat (1-100)
- `defense` (integer) - Defense stat (1-100)
- `fight_iq` (integer) - Fight IQ stat (1-100)
- `current_status` (string) - Boxer status (active, training, fighting, etc.)
- `created_at` (timestamp) - Boxer creation time
- `updated_at` (timestamp) - Last boxer update

### Scheduled Events Table
- `id` (UUID) - Primary key
- `user_id` (UUID) - Foreign key to users table
- `boxer_id` (UUID) - Foreign key to boxers table
- `event_type` (string) - Type of event (training, fight)
- `scheduled_time` (timestamp) - When the event is scheduled to occur
- `status` (string) - Event status (pending, completed, cancelled)
- `created_at` (timestamp) - Event creation time

### Training Queue Table
- `id` (UUID) - Primary key
- `user_id` (UUID) - Foreign key to users table
- `boxer_id` (UUID) - Foreign key to boxers table
- `training_type` (string) - Type of training (power, speed, etc.)
- `duration_minutes` (integer) - Duration of training
- `scheduled_time` (timestamp) - When the training is scheduled to occur
- `status` (string) - Training status (queued, in_progress, completed)
- `created_at` (timestamp) - Training creation time

### Fights Table
- `id` (UUID) - Primary key
- `fighter1_id` (UUID) - Foreign key to boxers table
- `fighter2_id` (UUID) - Foreign key to boxers table
- `scheduled_time` (timestamp) - When the fight is scheduled
- `status` (string) - Fight status (scheduled, in_progress, completed)
- `winner_id` (UUID) - Foreign key to boxers table (nullable)
- `created_at` (timestamp) - Fight creation time

### Fight History Table
- `id` (UUID) - Primary key
- `fight_id` (UUID) - Foreign key to fights table
- `rounds` (jsonb) - Fight round-by-round details
- `winner` (string) - Winner type (KO, TKO, decision)
- `created_at` (timestamp) - History creation time

## Implementation Details

### Migration File Fixes

The main compilation errors were caused by database syntax inconsistencies:

1. **Database Schema Inconsistencies (Fixed)**
   - Changed AUTOINCREMENT to SERIAL for PostgreSQL compatibility
   - Updated DATETIME to TIMESTAMP 
   - Corrected placeholder syntax from `$1, $2` (SQLite) to proper PostgreSQL format
   - Fixed RETURNING clause usage in PostgreSQL-style queries

2. **Field Reference Issues (Fixed)**
   - Removed references to non-existent fields like `user.ID`, `boxer.UserID`, `fight.ID` that weren't present in creation models
   - Modified database functions not to try to scan IDs back into creation model objects
   - Used proper PostgreSQL syntax throughout the migration file

3. **Compilation Errors Eliminated**
   - Removed all "declared and not used" errors
   - Fixed "undefined field" errors 
   - Resolved all database function call issues

### Key Technical Changes Made

In `internal/db/migration.go`:
- Changed all AUTOINCREMENT to SERIAL 
- Updated all DATETIME to TIMESTAMP
- Removed RETURNING clauses that tried to set ID fields on creation objects
- Replaced `Scan(&modelField)` calls with simple `Exec()` calls where IDs aren't needed
- Fixed placeholder syntax throughout

## Database Connection and Management

### Connection Pooling
- PostgreSQL connections are managed through the platform/database package
- Proper connection pooling is implemented for performance optimization
- Connection parameters are loaded from environment variables

### Migration Strategy
- Migrations are stored in migrations/ directory
- The database connection is configured through environment variables
- Schema initialization happens during application startup

## Performance Considerations

1. **Indexing**: Proper indexing on frequently queried fields (user_id, scheduled_time, etc.)
2. **Connection Management**: Efficient connection pooling for database operations
3. **Query Optimization**: PostgreSQL-specific query syntax and optimization techniques
4. **Data Types**: Appropriate data types for each field to ensure efficiency

## Testing Considerations

### Database Tests Status
- Database tests require `CGO_ENABLED=0` to be disabled (sqlite3 driver issue)
- Tests use a stub implementation due to environment limitations
- Core database functionality exists but is not fully testable in current environment

### Testing Approach
- Use in-memory databases (SQLite) for integration-like tests where possible
- Mock database connections at the SQL level for unit testing
- Test both success and failure scenarios
- Focus on transaction handling where applicable

## Future Improvements

1. **Complete Database Testing**: Set up proper database testing environment with CGO enabled
2. **Advanced Indexing**: Implement more sophisticated indexing strategies
3. **Query Optimization**: Further optimize complex queries for performance
4. **Backup Strategy**: Implement proper backup and recovery procedures
5. **Monitoring**: Add database performance monitoring capabilities