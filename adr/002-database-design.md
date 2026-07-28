# 2. Database Design and Schema

## Status
Accepted

## Context
The boxing simulator needed a robust database design that could support:
- User management with authentication
- Boxer statistics and status tracking
- Training and fight scheduling
- Time-based event processing
- Scalability for future features

## Decision
We chose PostgreSQL as the primary database with the following schema:

**Core Tables**:
1. `users` - User accounts with JWT tokens and metadata
2. `boxers` - Boxer statistics (power, speed, defense, etc.) and status
3. `scheduled_events` - Future events (training, fights) with scheduling
4. `training_queue` - Queued training actions with game-time deadlines
5. `fights` - Fight records with participants and status
6. `fight_history` - Detailed fight outcomes and round-by-round results

We implemented a modular approach using UUID primary keys and proper foreign key relationships.

## Consequences
- **Pros**:
  - PostgreSQL's robust transaction support ensures data integrity
  - JSONB column in fight_history table supports flexible fight result storage
  - UUIDs provide better security and scalability than auto-incrementing IDs
  - Proper indexing on frequently queried fields (user_id, scheduled_time) improves performance
  - Foreign key constraints maintain referential integrity

- **Cons**:
  - Need to manage database migrations manually or with tools
  - More complex schema design than simpler solutions
  - Requires careful consideration of data types and relationships

## Alternatives Considered
1. SQLite for development/testing: Would simplify setup but lacks production features like connection pooling and concurrency
2. MongoDB: Would provide flexibility but lacks transactional consistency needed for game state
3. MySQL: Similar to PostgreSQL but with different feature set and performance characteristics
4. NoSQL approach with multiple document stores: Would be more complex and harder to maintain consistency

The PostgreSQL choice provides the reliability and features necessary for a persistent, multi-user boxing simulation system.