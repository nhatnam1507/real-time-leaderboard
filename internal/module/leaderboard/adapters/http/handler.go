package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/shared/response"
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
// @Description Get the top players across all games
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Number of top players" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/leaderboard/global [get]
func (h *Handler) GetGlobalLeaderboard(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	leaderboard, err := h.leaderboardUseCase.GetGlobalLeaderboard(c.Request.Context(), limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, leaderboard, "Leaderboard retrieved successfully")
}

// GetGameLeaderboard handles getting a game-specific leaderboard
// @Summary Get game leaderboard
// @Description Get the top players for a specific game
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param game_id path string true "Game ID"
// @Param limit query int false "Number of top players" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/leaderboard/game/{game_id} [get]
func (h *Handler) GetGameLeaderboard(c *gin.Context) {
	gameID := c.Param("game_id")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	leaderboard, err := h.leaderboardUseCase.GetGameLeaderboard(c.Request.Context(), gameID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, leaderboard, "Leaderboard retrieved successfully")
}

// GetUserRank handles getting a user's rank
// @Summary Get user rank
// @Description Get a user's rank in a leaderboard
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param game_id query string false "Game ID (empty for global)"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
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

