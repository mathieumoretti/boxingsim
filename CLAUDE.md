This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

  Project Overview

  This is a Go-based REST API backend for a boxing simulation game. The system manages boxers, fights, users, and  game world state using PostgreSQL for data persistence and Redis for caching.

  Architecture and Structure

  The application follows a layered architecture pattern with clear separation of concerns:

  - cmd/: Entry point of the application
  - internal/: Core application logic organized by domain
    - model/: Data models and DTOs for API responses, requests, and database entities
    - service/: Business logic implementations
    - handler/: HTTP request handlers that coordinate between services and models
    - db/: Database operations and migrations
    - platform/: Platform-specific utilities (database, config, logger, redis)
    - auth/: Authentication logic
    - events/: Event handling system

  Key Components

  - Database: PostgreSQL with connection pooling managed through internal/platform/database
  - Caching: Redis integration for performance optimization (internal/platform/redis)
  - Configuration: Environment-based configuration management (internal/platform/config)
  - Routing: Gorilla Mux router for HTTP endpoints
  - Logging: Structured logging system (internal/platform/logger)

  Development Commands

  Building and Running

  - make build - Build the application
  - make run - Run the application directly with Go
  - make dev - Run with hot reload using air (requires installation)
  - make docker-up - Start all services using Docker Compose
  - make docker-down - Stop all Docker services

  Testing and Quality

  - make test - Run all tests
  - make lint - Run linters (go vet, golangci-lint)

  Database Operations

  - Migrations are stored in migrations/
  - The database connection is configured through environment variables

  Key Files to Understand

  - cmd/server/main.go - Main application entry point and server setup
  - internal/model/* - Data models defining the structure of entities
  - internal/service/* - Core business logic for boxing simulation
  - internal/handler/* - HTTP handlers that expose API endpoints
  - internal/platform/database/database.go - Database connection management
  - internal/platform/config/config.go - Configuration loading

  API Endpoints

  The system exposes:
  - Health check endpoint at /health
  - API v1 endpoints under /api/v1/* (currently stubbed)
  - Authentication endpoints in /auth/ (defined in internal/auth)

  Development Environment

  The project uses Docker Compose for service management with PostgreSQL and Redis. The development workflow
  supports hot reloading with air, and all environment variables are loaded through the configuration system. The
  application is designed to be run either directly or via Docker containers.

  Testing Strategy

  Tests are organized by package and include:
  - Unit tests for business logic in internal/service/*
  - Integration tests for database operations in internal/db/*
  - End-to-end API tests (not yet implemented)

  The testing framework uses Go's built-in testing capabilities with comprehensive test coverage of core
  functionality.

Incremental Build System

The project now includes an incremental build system to improve development workflow:
- make build - Incremental build of all packages 
- make test - Run tests for all packages
- make ci - Run complete CI pipeline (lint, build, test)
- make clean - Clean build artifacts

This system supports dependency-aware building and only rebuilds what's necessary, significantly speeding up development cycles.