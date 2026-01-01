package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"real-time-leaderboard/internal/shared/errors"
	"real-time-leaderboard/internal/shared/response"
)

// AuthMiddleware creates authentication middleware
// This will be implemented to use the auth module's application layer
type AuthMiddleware struct {
	validateToken func(ctx context.Context, token string) (string, error) // Returns userID and error
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
			response.Error(c, errors.NewUnauthorizedError("Authorization header is required"))
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errors.NewUnauthorizedError("Invalid authorization header format"))
			c.Abort()
			return
		}

		token := parts[1]
		userID, err := m.validateToken(c.Request.Context(), token)
		if err != nil {
			response.Error(c, errors.NewUnauthorizedError("Invalid or expired token"))
			c.Abort()
			return
		}

		// Store user ID in context
		c.Set("user_id", userID)
		c.Next()
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}
