// Package middleware provides HTTP middleware functions for authentication and request handling.
package middleware

import (
	"context"
	"strings"

	"real-time-leaderboard/internal/shared/logger"
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
	logger        *logger.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(validateToken func(ctx context.Context, token string) (string, error), l *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		validateToken: validateToken,
		logger:        l,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apiErr := response.NewUnauthorizedError("Authorization header is required")
			m.logger.Error(c.Request.Context(), apiErr.Error())
			response.Error(c, apiErr)
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, authHeaderPrefix) {
			apiErr := response.NewUnauthorizedError("Invalid authorization header format")
			m.logger.Error(c.Request.Context(), apiErr.Error())
			response.Error(c, apiErr)
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, authHeaderPrefix)
		userID, err := m.validateToken(c.Request.Context(), token)
		if err != nil {
			apiErr := response.AsAPIError(err)
			m.logger.Err(c.Request.Context(), err).Msg("Request error")
			response.Error(c, apiErr)
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
