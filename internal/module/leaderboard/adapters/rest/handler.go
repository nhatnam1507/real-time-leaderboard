// Package rest provides REST API handlers for the leaderboard module.
package rest

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
// @Summary Get global leaderboard
// @Description Get paginated global leaderboard showing top players across all games
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Number of top players" default(10) minimum(1) maximum(100) example(10)
// @Param offset query int false "Number of results to skip" default(0) minimum(0) example(0)
// @Success 200 {object} response.Response "Leaderboard retrieved successfully"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/leaderboard/global [get]
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
// @Summary Get game leaderboard
// @Description Get paginated leaderboard for a specific game showing top players
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param game_id path string true "Game ID" example(game1)
// @Param limit query int false "Number of top players" default(10) minimum(1) maximum(100) example(10)
// @Param offset query int false "Number of results to skip" default(0) minimum(0) example(0)
// @Success 200 {object} response.Response "Leaderboard retrieved successfully"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/leaderboard/game/{game_id} [get]
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
// @Summary Get user rank
// @Description Get a user's rank, score, and position in a leaderboard (global or game-specific)
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param user_id path string true "User ID" example(user123)
// @Param game_id query string false "Game ID (empty or 'global' for global leaderboard)" example(game1)
// @Success 200 {object} response.Response "User rank retrieved successfully"
// @Failure 404 {object} response.Response "User not found in leaderboard"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/leaderboard/rank/{user_id} [get]
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
