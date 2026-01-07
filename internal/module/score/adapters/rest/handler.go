// Package rest provides REST API handlers for the score module.
package rest

import (
	"real-time-leaderboard/internal/module/score/application"
	"real-time-leaderboard/internal/shared/middleware"
	"real-time-leaderboard/internal/shared/request"
	"real-time-leaderboard/internal/shared/response"
	"real-time-leaderboard/internal/shared/validator"

	"github.com/gin-gonic/gin"
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
// @Description Submit a score for a game. Requires authentication. The score will be saved and the leaderboard will be updated.
// @Tags scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body application.SubmitScoreRequest true "Score submission request" example({"game_id":"game1","score":1000,"metadata":{"level":5,"time":120}})
// @Success 201 {object} response.Response "Score submitted successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Router /api/v1/scores [post]
func (h *Handler) SubmitScore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, response.NewUnauthorizedError("User ID not found in context"))
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
// @Description Get paginated scores for the authenticated user. Optionally filter by game ID.
// @Tags scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param game_id query string false "Filter by game ID" example(game1)
// @Param limit query int false "Number of results per page" default(10) minimum(1) maximum(100) example(10)
// @Param offset query int false "Number of results to skip" default(0) minimum(0) example(0)
// @Success 200 {object} response.Response "Scores retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Router /api/v1/scores/me [get]
func (h *Handler) GetUserScores(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, response.NewUnauthorizedError("User ID not found in context"))
		return
	}

	var listReq request.ListRequest
	if err := listReq.FromGinContext(c); err != nil {
		response.Error(c, err)
		return
	}

	if err := listReq.Validate(); err != nil {
		response.Error(c, err)
		return
	}

	gameID := c.Query("game_id")

	scores, err := h.scoreUseCase.GetUserScores(c.Request.Context(), userID, gameID, &listReq)
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
