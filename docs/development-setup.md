# Development Setup

## Prerequisites

- Go 1.19 or higher installed
- Docker and Docker Compose for database services
- Make utility

## Getting Started

### 1. Install Dependencies

```bash
make deps
```

### 2. Start Database Services

```bash
make docker-up
```

### 3. Run Development Server

```bash
make dev
```

This will:
- Build the server binary
- Start the server on port 8080
- Serve both API endpoints and web UI files from the same server

## Web UI Access

Once the server is running, access the application at:
http://localhost:8080

## Hot Reload (Optional)

For hot reload capability during development:

1. Install air:
```bash
go install github.com/air-verse/air@latest
```

2. Run with hot reload:
```bash
make dev
```

Note: The server will build and run without hot reload if air is not installed.

## Project Structure

- `cmd/server/main.go` - Main application entry point
- `web/` - Web UI files (HTML, CSS, JavaScript)
- `internal/` - Core application logic organized by domain
- `docker-compose.yml` - Database service configuration

## API Endpoints

- `/health` - Health check endpoint
- `/auth/register` - User registration
- `/auth/login` - User login  
- `/boxers` - Boxer management endpoints

## Installing golangci-lint

For local development, `golangci-lint` should be installed in your GOPATH:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

In CI environments, you can install `golangci-lint` using the following approach:

### Option 1: Using a script to install it (recommended)
```bash
# Install golangci-lint in CI
curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
```

### Option 2: Using Docker (if your CI uses Docker)
```dockerfile
# In your Dockerfile
RUN curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v1.54.2
```

After installation, verify that golangci-lint is working:
```bash
golangci-lint version
```

If you encounter "command not found" errors in CI environments:

1. Make sure `GOPATH/bin` is in your PATH
2. Ensure the binary has execute permissions
3. Use the full path to the binary in your Makefile as shown in this project's Makefile