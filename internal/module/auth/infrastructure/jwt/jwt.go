package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"real-time-leaderboard/internal/module/auth/domain"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair generates access and refresh tokens
func (m *JWTManager) GenerateTokenPair(userID string) (*domain.TokenPair, error) {
	accessToken, accessExpiresIn, err := m.generateToken(userID, m.accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := m.generateToken(userID, m.refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessExpiresIn.Seconds()),
	}, nil
}

// generateToken generates a JWT token
func (m *JWTManager) generateToken(userID string, expiry time.Duration) (string, time.Duration, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiry, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (m *JWTManager) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", errors.New("invalid token")
}
