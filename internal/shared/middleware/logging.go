package middleware

import (
	"fmt"
	"time"

	"real-time-leaderboard/internal/shared/logger"

	"github.com/gin-gonic/gin"
)

// RequestLogger creates a logging middleware
func RequestLogger(l *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		msg := fmt.Sprintf("[%s] %s | %d | %v | %s | %s",
			c.Request.Method, c.Request.URL.Path, statusCode, latency, c.ClientIP(), c.Request.URL.RawQuery)

		if errorMessage != "" {
			l.Errorf(c.Request.Context(), "%s | %s", msg, errorMessage)
		} else {
			l.Info(c.Request.Context(), msg)
		}
	}
}
