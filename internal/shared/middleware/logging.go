package middleware

import (
	"time"

	"real-time-leaderboard/internal/shared/logger"

	"github.com/gin-gonic/gin"
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

		// Extract request ID for logging
		requestID := GetRequestID(c)
		log := l
		if requestID != "" {
			log = l.WithRequestID(requestID)
		}

		// Log to application logger (zerolog)
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
