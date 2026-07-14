# System Design and Architecture

## Overview

This document describes the architectural design of the Boxing Simulator backend system, including database schema, component interactions, and data flow.

## Architecture Principles

1. **Server-authoritative** - All simulation happens backend-side, clients only display state
2. **Event-driven** - Domain events (TRAINING_COMPLETE, FIGHT_START, FIGHT_COMPLETE) drive state changes
3. **Interface-based** - Use Go interfaces at module boundaries for testability and future extraction
4. **Domain-oriented** - Clean separation by business domain (boxer, fight, training, world)
5. **No circular dependencies** - Modules depend inward, never on each other

## Project Structure

```
/cmd
  /api          - HTTP server with REST endpoints
  /worker       - Background tick worker
  
/internal
  /auth         - JWT auth logic
  /world        - World clock and tick processing
  /boxer        - Boxer domain logic
  /training     - Training queue and stat calculations
  /store        - PostgreSQL repositories
  /events       - Domain event types
  /platform     - Database, Redis, logger adapters
  
/web           - Web UI files (HTML, CSS, JavaScript)
```

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

## Data Flow

1. **User Authentication**
   - User registers/login via `/auth/register` or `/auth/login`
   - JWT token is generated and returned
   - Subsequent requests include the JWT in Authorization header

2. **Boxer Management**
   - Users create, read, update boxers via `/boxers` endpoints
   - Boxer data is stored in PostgreSQL database
   - Redis cache may be used for frequently accessed data

3. **World Clock Processing**
   - Background worker runs every minute (tick)
   - Processes pending events in scheduled_events table
   - Updates boxer statuses and triggers events
   - Handles training completion, fight scheduling, etc.

4. **Training System**
   - Users queue training via `/training/queue` endpoint
   - Training is added to training_queue table
   - Tick worker processes queued training when time comes

5. **Fight System**
   - Users schedule fights via API endpoints
   - Fights are stored in fights table
   - Fight simulation occurs during tick processing
   - Results stored in fight_history table

## Domain Events

- `TRAINING_COMPLETE` - Triggered when training finishes
- `FIGHT_START` - Triggered when a fight begins
- `FIGHT_COMPLETE` - Triggered when a fight ends
- `BOXER_STATUS_UPDATE` - Triggered when boxer status changes

## Technology Stack

- **Backend**: Go 1.19+
- **Database**: PostgreSQL with connection pooling
- **Caching**: Redis integration
- **Configuration**: Environment-based configuration management
- **Routing**: Gorilla Mux router for HTTP endpoints
- **Logging**: Structured logging system
- **Deployment**: Docker Compose for service management

## Security Considerations

1. **Authentication**: JWT tokens for API authentication
2. **Authorization**: Role-based access control
3. **Data Validation**: Input validation at API boundaries
4. **Database Security**: Proper connection pooling and query parameterization
5. **Caching Security**: Redis connection security

## Performance Considerations

1. **Database Indexing**: Proper indexing on frequently queried fields
2. **Connection Pooling**: Efficient database connection management
3. **Redis Caching**: Strategic caching of frequently accessed data
4. **Background Processing**: Asynchronous event handling via tick worker
5. **API Response Optimization**: Efficient JSON serialization

## Scalability

The system is designed with scalability in mind:
- Modular architecture allows for microservice extraction
- Database and Redis connections are properly managed
- Background workers handle time-intensive operations
- Event-driven design reduces coupling between components