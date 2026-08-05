# Boxing Simulation Project

## Overview
This is a boxing simulation project that appears to include:
- A seeding system for organizing boxers
- Boxer management services
- A web UI integration
- Database setup and migration capabilities

## Repository Structure
```
.
├── cmd/
│   ├── seed/           # Seeding functionality
│   └── server/         # Server components
├── internal/
│   ├── boxer/          # Boxer-related logic
│   ├── database/       # Database handling
│   └── seeding/        # Seeding logic
├── public/             # Public assets (frontend)


```

## Key Components

### Seeding System
- The seeding functionality is in `cmd/seed/main.go`
- This appears to be responsible for organizing boxers into a seeding structure
- Related to the "seeding beginning" commit

### Boxer Management
- `internal/boxer/` package contains boxer-related logic
- The `cmd/seed/main.go` file likely uses this for managing boxer data during seeding

### Server Components
- `cmd/server/boxing-server` appears to be the main server binary
- This is likely where the web API and UI integration happens

### Database
- Database setup instructions in `DATABASE_SETUP.md`
- Migration information in `MIGRATION_SUMMARY.md`

## Recent Changes
The repository has recent commits related to:
- Seeding functionality (starting with "seeding beginning")
- Boxer management services
- Dashboard development
- Database migrations

## Development Notes
- The project appears to be using Go as the primary language
- There's a focus on boxing simulation and management
- Frontend integration seems to be part of the scope