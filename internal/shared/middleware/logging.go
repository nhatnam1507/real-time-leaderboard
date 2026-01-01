package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"real-time-leaderboard/internal/shared/logger"
)

// RequestLogger creates a logging middleware
func RequestLogger(l *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		requestID := c.GetString("request_id")
		log := l
		if requestID != "" {
			log = l.WithRequestID(requestID)
		}

		if errorMessage != "" {
			log.Errorf("[%s] %s | %d | %v | %s | %s | %s",
				method,
				path,
				statusCode,
				latency,
				clientIP,
				raw,
				errorMessage,
			)
		} else {
			log.Infof("[%s] %s | %d | %v | %s | %s",
				method,
				path,
				statusCode,
				latency,
				clientIP,
				raw,
			)
		}
	}
}

// RequestID adds a request ID to the context
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	seed := time.Now().UnixNano()
	for i := range b {
		b[i] = charset[seed%int64(len(charset))]
		seed = seed*1103515245 + 12345 // Simple LCG
	}
	return string(b)
}
