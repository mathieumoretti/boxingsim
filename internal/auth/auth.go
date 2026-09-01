package auth

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	cfg *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		cfg: cfg,
	}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return "", err
	}
	return string(hashedPassword), nil
}

func (s *AuthService) CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *AuthService) GenerateTokenPair(user *model.User) (*TokenPair, error) {
	now := time.Now()
	atClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"iat":      now.Unix(),
		"exp":      now.Add(15 * time.Minute).Unix(),
		"username": user.Username,
	})
	accessToken, err := atClaims.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken: accessToken,
	}, nil
}

func (s *AuthService) VerifyToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// RequireAuth middleware extracts and validates JWT from Authorization header.
// If valid, injects user into context; otherwise returns 401 Unauthorized.
func (s *AuthService) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Bearer token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Parse "Bearer <token>" format
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			http.Error(w, `{"error": "Authorization header must use Bearer scheme"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(authHeader[len(bearerPrefix):])

		// Verify JWT using existing method
		claims, err := s.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Extract user identity from claims
		subFloat, ok := (*claims)["sub"]
		if !ok {
			http.Error(w, `{"error": "Invalid token: missing subject claim"}`, http.StatusUnauthorized)
			return
		}

		userID, ok := subFloat.(float64)
		if !ok {
			http.Error(w, `{"error": "Invalid token: malformed subject claim"}`, http.StatusUnauthorized)
			return
		}

		username, ok := (*claims)["username"].(string)
		if !ok {
			http.Error(w, `{"error": "Invalid token: missing username claim"}`, http.StatusUnauthorized)
			return
		}

		// Create user object from claims (password not needed for auth context)
		user := &model.User{
			ID:       int(userID),
			Username: username,
		}

		// Inject user into context
		ctx := WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
