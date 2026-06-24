package auth

import (
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mormm/boxing/internal/platform/config"
)

type User struct {
	ID           int
	Username     string
	Email        string
	HashedPassword string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	cfg    *config.Config
	logger *Logger
}

type Logger struct {
	info  *log.Logger
	error *log.Logger
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		cfg:    cfg,
		logger: NewLogger(),
	}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *AuthService) GenerateTokenPair(user *User) (*TokenPair, error) {
	now := time.Now()
	atClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"iat":      now.Unix(),
		"exp":      now.Add(15 * time.Minute).Unix(),
		"username": user.Username,
	})

	at, err := atClaims.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	rtClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"iat": now.Unix(),
		"exp": now.Add(7 * 24 * time.Hour).Unix(),
	})

	rt, err := rtClaims.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  at,
		RefreshToken: rt,
	}, nil
}

func (s *AuthService) VerifyToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "[AUTH] ", log.LstdFlags|log.Lshortfile),
		error: log.New(os.Stdout, "[AUTH-ERROR] ", log.LstdFlags|log.Lshortfile),
	}
}