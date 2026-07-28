# 5. Seeding System

## Status
Accepted

## Context
The boxing simulator needed a way to populate the database with sample data for demonstration purposes and testing without requiring manual data entry.

## Decision
We implemented a database seeding system that:

**Core Features**:
1. Creates sample users with realistic usernames, emails, and passwords
2. Creates sample boxers with names, attributes (strength, defense, agility), and positions
3. Uses the existing password hashing approach (bcrypt)
4. Handles database connection errors gracefully
5. Continues seeding even if individual records fail
6. Follows the existing database schema
7. Preserves existing data - doesn't overwrite or delete anything

**Implementation Details**:
- Seeding functionality in `cmd/seed/main.go`
- Sample data includes realistic boxer names (Mike Tyson, Muhammad Ali, etc.)
- Boxers have realistic attributes and start at level 1 with default health and energy
- Can be run once during initial setup using `make seed` or `go run cmd/seed/main.go`

## Consequences
- **Pros**:
  - Provides immediate value to users by showing a populated demo environment
  - Simplifies onboarding process for new developers
  - Makes the application feel like a real boxing simulation world from start
  - Easy to run with simple commands (`make seed` or direct go run)
  - Graceful error handling ensures partial success even if some records fail
  - Preserves existing data while adding sample content

- **Cons**:
  - Adds complexity to the development workflow
  - Requires careful management of sample data to avoid conflicts
  - Could potentially be confusing if users don't understand it's demo data
  - Additional maintenance overhead for keeping sample data current

## Alternatives Considered
1. Manual data entry: Would require significant time investment and wouldn't provide immediate value
2. Sample data in migration files: Would complicate the migration process and make it harder to distinguish between schema and data changes
3. Separate test data generation tool: Would add more complexity without clear benefits
4. No seeding at all: Would leave new users with an empty application, reducing initial engagement

The seeding system provides immediate value for demonstration and development while maintaining a clean separation from the core application logic.