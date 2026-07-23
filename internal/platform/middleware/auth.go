package middleware

import (
	"net/http"
	"strings"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/logger"
)

// AuthMiddleware checks if a request has valid authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logger.New("auth-middleware")

		// Get the token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Error("Missing authorization header")
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			logger.Error("Invalid authorization header format")
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Verify the token
		cfg := config.Load()
		authService := auth.NewAuthService(cfg)

		_, err := authService.VerifyToken(tokenString)
		if err != nil {
			logger.Error("Invalid token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// If valid, add user info to context and proceed
		// In a real implementation, you might want to store this in request context
		// For now, we just validate that the token is valid

		logger.Info("User authenticated successfully")
		next.ServeHTTP(w, r)
	})
}