package middleware

import (
	"real-time-leaderboard/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

// RequestID creates a middleware that generates a unique request ID for each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in standard context for automatic propagation (OTel-like)
		c.Request = c.Request.WithContext(logger.WithRequestIDContext(c.Request.Context(), requestID))
		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Next()
	}
}

// GetRequestID extracts the request ID from the standard context
func GetRequestID(c *gin.Context) string {
	return logger.GetRequestIDFromContext(c.Request.Context())
}
