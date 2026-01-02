package middleware

import (
	"net/http"

	"real-time-leaderboard/internal/shared/errors"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Recovery creates a recovery middleware
func Recovery(l *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Get request ID if it exists (from RequestID middleware)
		requestID := GetRequestID(c)
		log := l
		if requestID != "" {
			log = l.WithRequestID(requestID)
		}

		log.Errorf("Panic recovered: %v", recovered)

		response.Error(c, errors.NewInternalError("Internal server error", nil))
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
