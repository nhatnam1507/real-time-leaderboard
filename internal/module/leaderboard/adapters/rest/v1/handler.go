// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"encoding/json"
	"fmt"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/middleware"
	"real-time-leaderboard/internal/shared/request"
	"real-time-leaderboard/internal/shared/response"
	"real-time-leaderboard/internal/shared/validator"

	"github.com/gin-gonic/gin"
)

const (
	// Keep-alive interval for SSE connections
	keepAliveInterval = 15 * time.Second
)

// LeaderboardHandler handles HTTP requests for leaderboards and scores
type LeaderboardHandler struct {
	leaderboardUseCase application.LeaderboardUseCase
	scoreUseCase       application.ScoreUseCase
	logger             *logger.Logger
}

// NewLeaderboardHandler creates a new leaderboard HTTP handler
func NewLeaderboardHandler(
	leaderboardUseCase application.LeaderboardUseCase,
	scoreUseCase application.ScoreUseCase,
	l *logger.Logger,
) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardUseCase: leaderboardUseCase,
		scoreUseCase:       scoreUseCase,
		logger:             l,
	}
}

// GetLeaderboard handles GET /leaderboard with pagination
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	var pagination request.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		valErr := validator.Validate(pagination)
		apiErr := toAPIError(valErr)
		h.logger.Err(c.Request.Context(), valErr).Msg("Request error")
		response.Error(c, apiErr)
		return
	}

	if err := validator.Validate(pagination); err != nil {
		apiErr := toAPIError(err)
		h.logger.Err(c.Request.Context(), err).Msg("Request error")
		response.Error(c, apiErr)
		return
	}

	ctx := c.Request.Context()
	_ = h.leaderboardUseCase.SyncFromPostgres(ctx)

	normalized := pagination.Normalize()
	entries, total, err := h.leaderboardUseCase.GetLeaderboard(ctx, normalized.GetLimit(), normalized.GetOffset())
	if err != nil {
		apiErr := toAPIError(err)
		h.logger.Err(c.Request.Context(), err).Msg("Request error")
		response.Error(c, apiErr)
		return
	}

	meta := response.NewPagination(normalized.GetOffset(), normalized.GetLimit(), total)
	response.SuccessWithMeta(c, entries, "Leaderboard retrieved successfully", meta)
}

// GetLeaderboardUpdate handles GET /leaderboard/stream via SSE for real-time delta updates
func (h *LeaderboardHandler) GetLeaderboardUpdate(c *gin.Context) {
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create context for the request
	ctx := c.Request.Context()

	// Sync from PostgreSQL to Redis if needed (lazy loading)
	_ = h.leaderboardUseCase.SyncFromPostgres(ctx)

	// Subscribe to entry delta updates
	updateCh, err := h.leaderboardUseCase.SubscribeToEntryUpdates(ctx)
	if err != nil {
		// If subscription fails, create a closed channel
		closedCh := make(chan *domain.LeaderboardEntry)
		close(closedCh)
		updateCh = closedCh
	}

	// Set up keep-alive ticker
	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	// Handle client disconnection
	notify := c.Writer.CloseNotify()

	// Keep connection, push delta updates from broadcaster
	for {
		select {
		case <-notify:
			// Client disconnected
			return

		case entry, ok := <-updateCh:
			if !ok {
				// Channel closed, connection ended
				return
			}

			// Send entry delta update to client using standard response format
			resp := response.Response{
				Success: true,
				Data:    entry,
				Message: "Leaderboard entry updated",
			}
			messageBytes, _ := json.Marshal(resp)
			_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", messageBytes)
			c.Writer.Flush()

		case <-ticker.C:
			// Send keep-alive comment
			_, _ = fmt.Fprintf(c.Writer, ": keep-alive\n\n")
			c.Writer.Flush()
		}
	}
}

// SubmitScore handles score update
func (h *LeaderboardHandler) SubmitScore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apiErr := response.NewUnauthorizedError("User ID not found in context")
		h.logger.Error(c.Request.Context(), apiErr.Error())
		response.Error(c, apiErr)
		return
	}

	var req application.SubmitScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		valErr := validator.Validate(req)
		apiErr := toAPIError(valErr)
		h.logger.Err(c.Request.Context(), valErr).Msg("Request error")
		response.Error(c, apiErr)
		return
	}

	if err := validator.Validate(req); err != nil {
		apiErr := toAPIError(err)
		h.logger.Err(c.Request.Context(), err).Msg("Request error")
		response.Error(c, apiErr)
		return
	}

	if err := h.scoreUseCase.SubmitScore(c.Request.Context(), userID, req); err != nil {
		apiErr := toAPIError(err)
		h.logger.Err(c.Request.Context(), err).Msg("Request error")
		response.Error(c, apiErr)
		return
	}

	response.Success(c, gin.H{"user_id": userID, "score": req.Score}, "Score updated successfully")
}

// RegisterPublicRoutes registers public leaderboard routes (no auth required)
func (h *LeaderboardHandler) RegisterPublicRoutes(router *gin.RouterGroup) {
	leaderboard := router.Group("/leaderboard")
	{
		leaderboard.GET("", h.GetLeaderboard)
		leaderboard.GET("/stream", h.GetLeaderboardUpdate)
	}
}

// RegisterProtectedRoutes registers protected leaderboard routes (auth required)
func (h *LeaderboardHandler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	leaderboard := router.Group("/leaderboard")
	{
		leaderboard.PUT("/score", h.SubmitScore)
	}
}
