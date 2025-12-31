package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"real-time-leaderboard/internal/shared/errors"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

// Recovery creates a recovery middleware
func Recovery(l *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := c.GetString("request_id")
		log := l
		if requestID != "" {
			log = l.WithRequestID(requestID)
		}

		log.Errorf("Panic recovered: %v", recovered)
		
		response.Error(c, errors.NewInternalError("Internal server error", nil))
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

