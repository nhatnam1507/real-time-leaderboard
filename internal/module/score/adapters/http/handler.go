package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"real-time-leaderboard/internal/module/score/application"
	"real-time-leaderboard/internal/shared/errors"
	"real-time-leaderboard/internal/shared/middleware"
	"real-time-leaderboard/internal/shared/response"
	"real-time-leaderboard/internal/shared/validator"
)

// Handler handles HTTP requests for scores
type Handler struct {
	scoreUseCase *application.ScoreUseCase
}

// NewHandler creates a new score HTTP handler
func NewHandler(scoreUseCase *application.ScoreUseCase) *Handler {
	return &Handler{
		scoreUseCase: scoreUseCase,
	}
}

// SubmitScore handles score submission
// @Summary Submit a score
// @Description Submit a score for a game
// @Tags scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body application.SubmitScoreRequest true "Score submission request"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/scores [post]
func (h *Handler) SubmitScore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errors.NewUnauthorizedError("User ID not found in context"))
		return
	}

	var req application.SubmitScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, validator.Validate(req))
		return
	}

	if err := validator.Validate(req); err != nil {
		response.Error(c, err)
		return
	}

	score, err := h.scoreUseCase.SubmitScore(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, score, "Score submitted successfully")
}

// GetUserScores handles retrieving user scores
// @Summary Get user scores
// @Description Get scores for the authenticated user
// @Tags scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param game_id query string false "Filter by game ID"
// @Param limit query int false "Limit results" default(10)
// @Param offset query int false "Offset results" default(0)
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/scores/me [get]
func (h *Handler) GetUserScores(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errors.NewUnauthorizedError("User ID not found in context"))
		return
	}

	gameID := c.Query("game_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	scores, err := h.scoreUseCase.GetUserScores(c.Request.Context(), userID, gameID, limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, scores, "Scores retrieved successfully")
}

// RegisterRoutes registers score routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	scores := router.Group("/scores")
	{
		scores.POST("", h.SubmitScore)
		scores.GET("/me", h.GetUserScores)
	}
}
