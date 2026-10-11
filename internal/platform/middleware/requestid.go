package middleware

import (
	"context"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"net/http"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const requestIDContextKey contextKey = "request_id"

// RequestIDMiddleware generates a unique request ID for each incoming HTTP request
// and injects it into the request context. It also adds the X-Request-ID header
// to both request and response for traceability.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (e.g., from upstream proxy)
		requestID := r.Header.Get("X-Request-ID")

		// Generate new ID if not present
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add request ID to context for downstream use
		ctx := ContextWithRequestID(r.Context(), requestID)

		// Set response header for client visibility
		w.Header().Set("X-Request-ID", requestID)

		// Continue with modified request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID creates a unique ID using crypto/rand
func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := cryptoRand.Read(bytes); err != nil {
		// Fallback to simple counter-based ID (should never happen in practice)
		return "req-" + hex.EncodeToString(make([]byte, 8))
	}
	return hex.EncodeToString(bytes)
}

// RequestIDFromContext extracts the request ID from the context.
// Returns empty string if no request ID is present.
func RequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDContextKey).(string); ok {
		return reqID
	}
	return ""
}

// RequestIDFromRequest is a convenience wrapper for RequestIDFromContext(r.Context()).
func RequestIDFromRequest(r *http.Request) string {
	return RequestIDFromContext(r.Context())
}

// ContextWithRequestID adds the request ID to the context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}
