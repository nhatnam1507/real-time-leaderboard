// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/shared/request"
	"real-time-leaderboard/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for leaderboards
type Handler struct {
	leaderboardUseCase *application.LeaderboardUseCase
}

// NewHandler creates a new leaderboard HTTP handler
func NewHandler(leaderboardUseCase *application.LeaderboardUseCase) *Handler {
	return &Handler{
		leaderboardUseCase: leaderboardUseCase,
	}
}

// GetGlobalLeaderboard handles getting the global leaderboard
func (h *Handler) GetGlobalLeaderboard(c *gin.Context) {
	var listReq request.ListRequest
	if err := listReq.FromGinContext(c); err != nil {
		response.Error(c, err)
		return
	}

	if err := listReq.Validate(); err != nil {
		response.Error(c, err)
		return
	}

	leaderboard, err := h.leaderboardUseCase.GetGlobalLeaderboard(c.Request.Context(), &listReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, leaderboard, "Leaderboard retrieved successfully")
}

// GetGameLeaderboard handles getting a game-specific leaderboard
func (h *Handler) GetGameLeaderboard(c *gin.Context) {
	gameID := c.Param("game_id")

	var listReq request.ListRequest
	if err := listReq.FromGinContext(c); err != nil {
		response.Error(c, err)
		return
	}

	if err := listReq.Validate(); err != nil {
		response.Error(c, err)
		return
	}

	leaderboard, err := h.leaderboardUseCase.GetGameLeaderboard(c.Request.Context(), gameID, &listReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, leaderboard, "Leaderboard retrieved successfully")
}

// GetUserRank handles getting a user's rank
func (h *Handler) GetUserRank(c *gin.Context) {
	userID := c.Param("user_id")
	gameID := c.DefaultQuery("game_id", "global")

	entry, err := h.leaderboardUseCase.GetUserRank(c.Request.Context(), gameID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, entry, "User rank retrieved successfully")
}

// RegisterRoutes registers leaderboard routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	leaderboard := router.Group("/leaderboard")
	{
		leaderboard.GET("/global", h.GetGlobalLeaderboard)
		leaderboard.GET("/game/:game_id", h.GetGameLeaderboard)
		leaderboard.GET("/rank/:user_id", h.GetUserRank)
	}
}

