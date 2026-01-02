// Package middleware provides HTTP middleware functions for authentication and request handling.
package middleware

import (
	"context"
	"strings"

	"real-time-leaderboard/internal/shared/response"

	"github.com/gin-gonic/gin"
)

const (
	authHeaderPrefix = "Bearer "
	userIDKey        = "user_id"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	validateToken func(ctx context.Context, token string) (string, error)
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(validateToken func(ctx context.Context, token string) (string, error)) *AuthMiddleware {
	return &AuthMiddleware{
		validateToken: validateToken,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, response.NewUnauthorizedError("Authorization header is required"))
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, authHeaderPrefix) {
			response.Error(c, response.NewUnauthorizedError("Invalid authorization header format"))
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, authHeaderPrefix)
		userID, err := m.validateToken(c.Request.Context(), token)
		if err != nil {
			response.Error(c, response.NewUnauthorizedError("Invalid or expired token"))
			c.Abort()
			return
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(userIDKey)
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}
