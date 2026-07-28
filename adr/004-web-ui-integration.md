# 4. Web UI Integration Approach

## Status
Accepted

## Context
The boxing simulator needed to serve both API endpoints and web UI files from a single application for:
- Development convenience
- Simplified deployment process
- Consistent user experience
- Unified port access

## Decision
We chose to integrate the web UI directly into the Go server by serving static files from the `web/` directory through the same HTTP server that serves API endpoints.

**Implementation Details**:
- Modified `cmd/server/main.go` to include a file server handler that serves static files from the `web/` directory
- Both API endpoints and UI files are served from the same port (8080)
- This approach allows for easy development with a single running server

## Consequences
- **Pros**:
  - Simplified deployment: Single server handles both API and UI
  - Development convenience: No need to run separate servers during development
  - Reduced complexity: Fewer moving parts to manage
  - Consistent port access: All traffic goes through one port
  - Clean, integrated experience that maintains all existing API functionality

- **Cons**:
  - In production, it's often better to separate API and UI for independent scaling
  - Less granular security control over UI vs API policies
  - Different caching requirements for static assets vs API responses
  - Large UI files could impact API server performance in some configurations

## Alternatives Considered
1. Separate web server (nginx/Apache) for UI with reverse proxy to API: Better for production but more complex deployment
2. CDN for static assets with API on separate server: More scalable but adds infrastructure complexity
3. Microservices architecture with dedicated UI service: Overkill for current scope but would scale well
4. Client-side routing with single-page application: Would require additional frontend framework complexity

The current approach is suitable for development and demonstration purposes, providing a clean integrated experience while maintaining all existing API functionality.