package auth

import (
	"context"
	"net/http"

	"github.com/mormm/boxing/internal/model"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const userContextKey contextKey = "user"

// UserFromContext extracts the authenticated user from request context.
// Returns nil if no user is present (e.g., unauthenticated endpoint).
func UserFromContext(ctx context.Context) *model.User {
	user, ok := ctx.Value(userContextKey).(*model.User)
	if !ok {
		return nil
	}
	return user
}

// UserFromRequest is a convenience wrapper for UserFromContext(r.Context()).
func UserFromRequest(r *http.Request) *model.User {
	return UserFromContext(r.Context())
}

// WithUser adds the authenticated user to the request context.
func WithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
