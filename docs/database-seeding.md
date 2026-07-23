# Database Seeding System

This document describes the database seeding implementation for the Boxing Simulator.

## Overview

The database seeding system provides functionality to populate the application with sample data for demonstration purposes. This includes sample users and boxers that create a realistic boxing simulation world from the start.

## Implementation Details

### Files Created

1. **`internal/db/seed.go`** - Main seeding logic
2. **`cmd/seed/main.go`** - Dedicated command to run seeding
3. **`SEEDING.md`** - User documentation for running seeding
4. **Updated `Makefile`** - Added `make seed` command

### Seeding Logic

The seeding system:
- Creates 3 sample users with realistic usernames and passwords
- Creates 5 sample boxers with boxing attributes (strength, defense, agility)
- Uses the existing password hashing approach (bcrypt) for user passwords
- Follows the existing database schema structure
- Handles errors gracefully without stopping the entire process

### Password Hashing

The system uses bcrypt to hash passwords as required by the authentication system. When seeding data:
1. If a password is already hashed, it's used as-is
2. If a password is plain text, it gets hashed using `auth.AuthService.HashPassword()`
3. This maintains consistency with how passwords are handled in the actual authentication flow

### Sample Data

#### Users
- `boxingfan` with email `boxingfan@example.com`
- `champ` with email `champ@example.com`
- `puncher` with email `puncher@example.com`

#### Boxers
The system creates 5 notable boxers with realistic attributes:
1. **Mike Tyson** - "The Baddest Man on the Planet"
2. **Muhammad Ali** - "The Greatest" 
3. **Floyd Mayweather** - "The Matrix"
4. **Sugar Ray Leonard** - "The Lion"
5. **Joe Frazier** - "The Executioner"

Each boxer has:
- Name and nickname (optional)
- Position coordinates (x, y)
- Boxing attributes: strength, defense, agility
- Default health and energy values

## Usage

### Running the Seeding Process

From the project root:

```bash
make seed
```

Or directly:

```bash
go run cmd/seed/main.go
```

### Integration with Development Flow

The seeding process is designed to:
- Not overwrite existing data (idempotent)
- Continue if individual records fail
- Work with the existing database connection setup
- Be easily runnable during initial setup or reset scenarios

## Design Considerations

1. **Separation of Concerns**: The seeding logic is separated from the main server code for clarity and maintainability.

2. **Error Handling**: The system handles errors gracefully - if one user or boxer fails to create, it continues with others.

3. **Reusability**: The same seeding data can be used for testing, development, and demonstration purposes.

4. **Consistency**: Uses the existing authentication and database patterns to maintain code consistency.

5. **Idempotency**: Running the seed multiple times should not duplicate data (though it's designed to be safe).

## Testing

The seeding functionality has been tested to ensure:
- Database connections work properly
- Passwords are hashed correctly
- Sample data is inserted into appropriate tables
- Errors in individual records don't prevent overall process completion
- The system works with existing database schema

This seeding system provides a complete demonstration environment for the boxing simulator right from the start.