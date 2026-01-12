// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/shared/middleware"
	"real-time-leaderboard/internal/shared/response"
	"real-time-leaderboard/internal/shared/validator"

	"github.com/gin-gonic/gin"
)

// ScoreHandler handles HTTP requests for score submission
type ScoreHandler struct {
	scoreUseCase *application.ScoreUseCase
}

// NewScoreHandler creates a new score HTTP handler
func NewScoreHandler(scoreUseCase *application.ScoreUseCase) *ScoreHandler {
	return &ScoreHandler{
		scoreUseCase: scoreUseCase,
	}
}

// SubmitScore handles score update
func (h *ScoreHandler) SubmitScore(c *gin.Context) {
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

	if err := h.scoreUseCase.SubmitScore(c.Request.Context(), userID, req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"user_id": userID, "point": req.Point}, "Score updated successfully")
}
