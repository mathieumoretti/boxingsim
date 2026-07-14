# Boxing Simulator Project Overview

## Context

Building a persistent multiplayer boxing management simulator with Go, PostgreSQL, and Redis. The game runs on accelerated time (1 real minute = 1 game hour) with a central world clock driving the simulation. We're starting with a modular monolith architecture designed for future extraction into microservices.

## Goal

Create a server-authoritative backend that can:
- Manage boxers with stats (Power, Speed, Chin, Cardio, Defense, Fight IQ)
- Handle training with fatigue and diminishing returns
- Run fights with round-by-round simulation
- Maintain world state through tick-driven events
- Persist all data to PostgreSQL and use Redis for caching

## Approach

Implement in three phases, each delivering playable functionality:

### Phase 1: Core Infrastructure & Boxer Management
**Deliverables:**
- Docker Compose setup (PostgreSQL + Redis)
- Project structure with Go modules
- Database schema (users, boxers, scheduled_events, training_queue)
- JWT authentication system
- Boxer CRUD operations with persistence
- World clock tick worker (runs every real minute)
- Training queue system
- Initial world clock processing (time passage)

**Key Components:**
1. **Project Structure**
```
/cmd/api          - HTTP server with REST endpoints
/cmd/worker       - Background tick worker
/internal/auth    - JWT auth logic
/internal/world   - World clock and tick processing
/internal/boxer   - Boxer domain logic
/internal/training - Training queue and stat calculations
/internal/store   - PostgreSQL repositories
/internal/events  - Domain event types
/internal/platform - Database, Redis, logger adapters
```

2. **Database Schema**
- `users` table - JWT tokens, user metadata
- `boxers` table - boxer stats, name, user_id, current_status
- `scheduled_events` table - future events (fights, training completion)
- `training_queue` table - queued training actions with game-time deadlines
- `fights` table - fight history (for Phase 2)
- `fight_history` table - fight history with outcomes (for Phase 2)

3. **API Endpoints**
- `POST /auth/register` - User registration
- `POST /auth/login` - JWT login
- `GET /boxers` - List user's boxers
- `POST /boxers` - Create new boxer
- `GET /boxers/:id` - Get boxer details
- `GET /world/time` - Current game time
- `GET /training/queue` - List queued training

### Phase 2: Fight System
**Deliverables:**
- AI opponent generation
- Fight scheduling API
- Round-by-round fight simulation
- Fight commentary generation
- Fight outcome determination (KO/TKO/decision)
- Fight history tracking
- Boxer ranking system

**New Components:**
- `internal/fight` - Fight simulation logic
- `internal/simulation` - Fight result calculation
- `internal/matchmaking` - AI opponent selection

### Phase 3: Advanced Features
**Deliverables:**
- Matchmaking API
- Multiplayer fight scheduling
- Economy (money, betting)
- Social features
- Advanced ranking with ELO system

## Architecture Principles

1. **Server-authoritative** - All simulation happens backend-side, clients only display state
2. **Event-driven** - Domain events (TRAINING_COMPLETE, FIGHT_START, FIGHT_COMPLETE) drive state changes
3. **Interface-based** - Use Go interfaces at module boundaries for testability and future extraction
4. **Domain-oriented** - Clean separation by business domain (boxer, fight, training, world)
5. **No circular dependencies** - Modules depend inward, never on each other

## Critical Files

- `cmd/api/main.go` - HTTP server entry point
- `cmd/worker/main.go` - Tick worker entry point
- `internal/world/ticker.go` - World clock and tick processing
- `internal/boxer/repository.go` - Boxer repository interface and implementation
- `internal/training/queue.go` - Training queue management
- `internal/platform/database/db.go` - PostgreSQL connection and migrations
- `internal/platform/redis/redis.go` - Redis connection
- `internal/events/events.go` - Domain event definitions
- `docker-compose.yml` - Service orchestration

## Implementation Order

1. Project scaffolding and Docker setup
2. Database schema and migrations
3. Authentication system
4. Boxer CRUD operations
5. World clock tick worker
6. Training queue system
7. Fight simulation (Phase 2)
8. Ranking system (Phase 2)

## Testing Strategy

- Unit tests for domain logic (training calculations, fight simulation)
- Integration tests for API endpoints
- Tick worker tested with mocked database/Redis
- Fight simulation tested with various stat combinations
- Use Go's testing package with table-driven tests

## Frontend Development Approaches

The backend provides a REST API that can be consumed by any frontend client. Recommended approaches:

### 1. Web Application (React/Vue/Angular)
Create a modern web-based UI that:
- Handles user authentication
- Displays boxer statistics and status
- Shows training queue and scheduled events
- Provides controls for managing boxers
- Visualizes the boxing world state

### 2. Mobile Application (React Native/Flutter)
Build a mobile app for:
- On-the-go boxer management
- Training scheduling
- Fight tracking
- Social features

### 3. Desktop Application (Electron)
Create a desktop client that:
- Provides full simulation controls
- Displays training progress
- Shows fight history and rankings