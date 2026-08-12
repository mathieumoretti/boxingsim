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

## File Path Rules
* Always use Windows-style backslashes (`\`) for file operations.
* Always use absolute paths with drive letters (e.g., `C:\path\to\project\main.go`).
* NEVER use relative paths or forward slashes.

## Go Writing Rules
* If a file write or edit fails due to string matching, do not try to use the Edit tool again.
* Fall back immediately to writing the file using a shell block via powershell: `New-Item -Force` or a redirect script.

## Makefile Operations
* Use `make <command>` for common operations instead of running commands directly.
* Available make targets (see `make help`):
  * **build** - Build the application (`go build`)
  * **run** - Run the built application
  * **dev** - Run with hot reload using air
  * **test** - Run all tests with gotestsum
  * **lint** - Run golangci-lint
  * **fmt** - Format code with gofmt
  * **migrate** - Run database migrations
  * **seed-ref** - Seed reference data
  * **seed-dev** - Seed development data  
  * **world** - Generate complete world (full seeding)
  * **reset-dev** - Reset and reseed for development (migrate + seed-ref + seed-dev)
* If a needed operation doesn't exist in the Makefile, add a new make target following the existing conventions.
