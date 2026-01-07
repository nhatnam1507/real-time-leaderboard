// Package v1 provides REST API v1 handlers and routing.
package v1

import (
	authV1 "real-time-leaderboard/internal/module/auth/adapters/rest/v1"
	"real-time-leaderboard/internal/module/auth/application"
	leaderboardV1 "real-time-leaderboard/internal/module/leaderboard/adapters/rest/v1"
	reportV1 "real-time-leaderboard/internal/module/report/adapters/rest/v1"
	scoreV1 "real-time-leaderboard/internal/module/score/adapters/rest/v1"
	"real-time-leaderboard/internal/shared/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up all v1 API routes
func SetupRoutes(
	router *gin.RouterGroup,
	authUseCase *application.AuthUseCase,
	authHandler *authV1.Handler,
	scoreHandler *scoreV1.Handler,
	leaderboardHandler *leaderboardV1.Handler,
	reportHandler *reportV1.Handler,
) {
	// Auth routes (no auth required)
	authHandler.RegisterRoutes(router)

	// Protected routes
	authMiddleware := middleware.NewAuthMiddleware(authUseCase.ValidateToken)
	protected := router.Group("")
	protected.Use(authMiddleware.RequireAuth())
	{
		scoreHandler.RegisterRoutes(protected)
	}

	// Public routes
	leaderboardHandler.RegisterRoutes(router)
	reportHandler.RegisterRoutes(router)
}
