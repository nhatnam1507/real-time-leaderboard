// Package v1 provides REST API v1 handlers for the score module.
package v1

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
