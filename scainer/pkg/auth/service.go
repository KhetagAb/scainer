package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const ContextUsernameKey = "username"

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Service struct {
	username, password, secret []byte
	ttl                        time.Duration
}

func New(username, password, secret string, ttl time.Duration) Service {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return Service{
		username: []byte(username),
		password: []byte(password),
		secret:   []byte(secret),
		ttl:      ttl,
	}
}

func (s Service) ValidateCredentials(username, password string) bool {
	if subtle.ConstantTimeCompare([]byte(username), s.username) != 1 {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(password), s.password) != 1 {
		return false
	}
	return true
}

func (s Service) IssueToken(username string) (token string, expiresIn int, err error) {
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign token: %w", err)
	}
	return signed, int(s.ttl.Seconds()), nil
}

func (s Service) ParseToken(bearerHeader string) (*Claims, error) {
	tokenString := strings.TrimSpace(strings.TrimPrefix(bearerHeader, "Bearer "))
	if tokenString == "" {
		return nil, ErrInvalidToken
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Username == "" {
		claims.Username = claims.Subject
	}
	return claims, nil
}
