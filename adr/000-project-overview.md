# 0. Project Overview and Architecture Decision Summary

## Status
Accepted

## Context
The boxing simulator project required a comprehensive approach to architecture, technology selection, and implementation patterns that would support:
- A robust, scalable backend for game simulation
- User management with authentication
- Database persistence for boxer data and game state
- Web UI integration for user interaction
- Development workflow that supports rapid iteration
- Production deployment capabilities

## Decision
This project implements a modular monolith architecture using Go as the primary language with PostgreSQL database and Redis caching. The system is designed with server-authoritative approach where all simulation happens backend-side, with clients only displaying state.

The core components include:
- JWT-based authentication system with bcrypt password hashing
- PostgreSQL database with well-defined schema for users, boxers, events, and training
- Docker Compose for consistent development and deployment environments
- Web UI integration through single server serving both API and static files
- Database seeding functionality for demonstration purposes
- Event-driven architecture using domain events

## Consequences
- **Pros**:
  - Clean separation of concerns with domain-oriented packages
  - Scalable architecture that supports future microservice extraction
  - Consistent development and deployment workflow
  - Robust authentication and data integrity
  - Easy to set up and run locally with Docker
  - Comprehensive documentation covering all aspects

- **Cons**:
  - Requires learning curve for Go developers unfamiliar with the language
  - More complex than simple monolithic approaches
  - Need to maintain multiple configuration files and tooling
  - Database schema changes require careful migration management

## Architecture Summary
The system follows a modular approach where:
1. `cmd/` directory contains application entry points (server, seed)
2. `internal/` directory contains core business logic organized by domain
3. `web/` directory holds static web UI files
4. `docker-compose.yml` defines service dependencies and configuration

This architecture provides a solid foundation for building a complete boxing simulation platform with features like fighter management, competitive fights, training systems, and event scheduling.