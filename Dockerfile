# Boxing Simulator - Multi-stage Docker Build
# Optimized for small image size and security

# -----------------------------------------------------------------------------
# Stage 1: Builder
# -----------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies (git needed for go modules)
RUN apk add --no-cache git ca-certificates

# Copy go mod files first (better layer caching - changes to source won't rebuild deps)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build both server and worker with optimizations
# -ldflags="-s -w" strips debug symbols for smaller binary (~15MB savings)
# CGO_ENABLED=0 creates fully static binary (no libc dependency)
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -ldflags="-s -w" -o server ./cmd/server && \
    go build -a -installsuffix cgo -ldflags="-s -w" -o worker ./cmd/worker

# -----------------------------------------------------------------------------
# Stage 2: Runtime (Alpine with minimal packages)
# -----------------------------------------------------------------------------
FROM alpine:3.20 AS runtime

# Install runtime dependencies (ca-certificates for HTTPS, tzdata for timezones)
RUN apk --no-cache add ca-certificates tzdata && \
    # Create non-root user for security
    addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

WORKDIR /app

# Copy binaries from builder (using --chown to set ownership in single layer)
COPY --from=builder --chown=appuser:appgroup /app/server .
COPY --from=builder --chown=appuser:appgroup /app/worker .

# Switch to non-root user for security
USER appuser

# Expose port
EXPOSE 8080

# Health check endpoint (adjust path if needed)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/bin/sh", "-c", "wget -q --spider http://localhost:8080/health || exit 1"]

# Default command runs the server
CMD ["./server"]
