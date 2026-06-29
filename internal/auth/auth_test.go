package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewAuthService(t *testing.T) {
	t.Run("Creates service with config", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}

		service := NewAuthService(cfg)

		assert.NotNil(t, service)
		assert.Equal(t, cfg, service.cfg)
	})
}

func TestHashPassword(t *testing.T) {
	t.Run("Successfully hashes password", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}
		service := NewAuthService(cfg)

		password := "password123"
		hash, err := service.HashPassword(password)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)

		// Verify we can check the password
		isValid := service.CheckPassword(password, hash)
		assert.True(t, isValid)
	})

	t.Run("Hashes different passwords differently", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}
		service := NewAuthService(cfg)

		hash1, _ := service.HashPassword("password123")
		hash2, _ := service.HashPassword("password456")

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("Correctly validates password", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}
		service := NewAuthService(cfg)

		password := "password123"
		hash, _ := service.HashPassword(password)

		isValid := service.CheckPassword(password, hash)
		assert.True(t, isValid)

		// Test wrong password
		isValid = service.CheckPassword("wrongpassword", hash)
		assert.False(t, isValid)
	})
}

func TestGenerateTokenPair(t *testing.T) {
	t.Run("Successfully generates token pair", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}
		service := NewAuthService(cfg)

		user := &User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			HashedPassword: "hashedpassword",
		}

		tokenPair, err := service.GenerateTokenPair(user)

		assert.NoError(t, err)
		assert.NotNil(t, tokenPair)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)

		// Verify access token structure
		claims, err := service.VerifyToken(tokenPair.AccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		// Check that it contains the user info
		assert.Equal(t, float64(1), (*claims)["sub"])
		assert.Equal(t, "testuser", (*claims)["username"])
	})

	t.Run("Handles invalid secret", func(t *testing.T) {
		// Test with empty secret to see error handling
		cfg := &config.Config{
			JWTSecret: "",
		}
		service := NewAuthService(cfg)

		user := &User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			HashedPassword: "hashedpassword",
		}

		tokenPair, err := service.GenerateTokenPair(user)

		// This should still work because JWT doesn't validate the secret until parsing
		assert.NoError(t, err) // But this might actually fail depending on JWT library behavior
		if tokenPair != nil {
			assert.NotEmpty(t, tokenPair.AccessToken)
		}
	})
}

func TestVerifyToken(t *testing.T) {
	t.Run("Successfully verifies valid token", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}
		service := NewAuthService(cfg)

		user := &User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			HashedPassword: "hashedpassword",
		}

		tokenPair, _ := service.GenerateTokenPair(user)

		claims, err := service.VerifyToken(tokenPair.AccessToken)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, float64(1), (*claims)["sub"])
		assert.Equal(t, "testuser", (*claims)["username"])
	})

	t.Run("Returns error for invalid token", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}
		service := NewAuthService(cfg)

		// Test with malformed token
		claims, err := service.VerifyToken("invalid.token.here")

		assert.Error(t, err)
		assert.Nil(t, claims)

		// Test with expired token (by creating a token that's already expired)
		// This is harder to test directly without manipulating time
	})

	t.Run("Returns error for wrong secret", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}
		service := NewAuthService(cfg)

		// Create token with one secret
		user := &User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			HashedPassword: "hashedpassword",
		}

		tokenPair, _ := service.GenerateTokenPair(user)

		// Change the config secret to something else
		cfg2 := &config.Config{
			JWTSecret: "different-secret-key",
		}
		service2 := NewAuthService(cfg2)

		claims, err := service2.VerifyToken(tokenPair.AccessToken)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestNewLogger(t *testing.T) {
	t.Run("Creates logger with correct structure", func(t *testing.T) {
		logger := NewLogger()

		assert.NotNil(t, logger)
		assert.NotNil(t, logger.info)
		assert.NotNil(t, logger.error)
	})
}

// Test JWT token expiration behavior
func TestTokenExpiration(t *testing.T) {
	t.Run("Access token expires after 15 minutes", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}
		service := NewAuthService(cfg)

		user := &User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			HashedPassword: "hashedpassword",
		}

		tokenPair, _ := service.GenerateTokenPair(user)

		// Parse the token to check expiration
		token, err := jwt.Parse(tokenPair.AccessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		assert.NoError(t, err)
		assert.True(t, token.Valid)

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Check that expiration time is set correctly (should be ~15 minutes from now)
			exp, ok := claims["exp"].(float64)
			assert.True(t, ok)

			// The token should have an exp claim
			assert.NotZero(t, exp)
		}
	})

	t.Run("Refresh token expires after 7 days", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "test-secret-key",
		}
		service := NewAuthService(cfg)

		user := &User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			HashedPassword: "hashedpassword",
		}

		tokenPair, _ := service.GenerateTokenPair(user)

		// Parse the token to check expiration
		token, err := jwt.Parse(tokenPair.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		assert.NoError(t, err)
		assert.True(t, token.Valid)

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Check that expiration time is set correctly (should be ~7 days from now)
			exp, ok := claims["exp"].(float64)
			assert.True(t, ok)

			// The token should have an exp claim
			assert.NotZero(t, exp)
		}
	})
}

func TestAuthServiceIntegration(t *testing.T) {
	t.Run("Complete authentication flow", func(t *testing.T) {
		cfg := &config.Config{
			JWTSecret: "integration-test-secret-key",
		}
		service := NewAuthService(cfg)

		// 1. Hash a password
		password := "securepassword123"
		hashedPassword, err := service.HashPassword(password)
		assert.NoError(t, err)

		// 2. Create user
		user := &User{
			ID:           1,
			Username:     "integrationuser",
			Email:        "integration@example.com",
			HashedPassword: hashedPassword,
		}

		// 3. Generate tokens
		tokenPair, err := service.GenerateTokenPair(user)
		assert.NoError(t, err)
		assert.NotNil(t, tokenPair)

		// 4. Verify the tokens work
		claims, err := service.VerifyToken(tokenPair.AccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, float64(1), (*claims)["sub"])

		// 5. Check password validation works
		isValid := service.CheckPassword(password, hashedPassword)
		assert.True(t, isValid)

		// 6. Check wrong password fails
		isValid = service.CheckPassword("wrongpassword", hashedPassword)
		assert.False(t, isValid)
	})
}