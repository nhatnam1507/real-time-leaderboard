package middleware

import (
	"net/http"

	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Recovery creates a recovery middleware
func Recovery(l *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		l.Errorf(c.Request.Context(), "Panic recovered: %v", recovered)
		apiErr := response.NewInternalError("Internal server error")
		response.Error(c, apiErr)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
