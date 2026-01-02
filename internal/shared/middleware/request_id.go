package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID creates a middleware that generates a unique request ID for each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request already has a request ID from client
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a new UUID as request ID
			requestID = uuid.New().String()
		}

		// Store in context for use in other handlers
		c.Set("request_id", requestID)

		// Set in response header
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

// GetRequestID extracts the request ID from the context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
