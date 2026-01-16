// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"encoding/json"
	"fmt"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/module/leaderboard/domain"
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
	leaderboardUseCase *application.LeaderboardUseCase
	scoreUseCase       *application.ScoreUseCase
}

// NewLeaderboardHandler creates a new leaderboard HTTP handler
func NewLeaderboardHandler(
	leaderboardUseCase *application.LeaderboardUseCase,
	scoreUseCase *application.ScoreUseCase,
) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardUseCase: leaderboardUseCase,
		scoreUseCase:       scoreUseCase,
	}
}

// GetLeaderboardPaginated handles GET /leaderboard with pagination
func (h *LeaderboardHandler) GetLeaderboardPaginated(c *gin.Context) {
	var pagination request.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.Error(c, validator.Validate(pagination))
		return
	}

	if err := validator.Validate(pagination); err != nil {
		response.Error(c, err)
		return
	}

	ctx := c.Request.Context()

	// Sync from PostgreSQL to Redis if needed (lazy loading)
	_ = h.leaderboardUseCase.SyncFromPostgres(ctx)

	// Normalize pagination
	normalized := pagination.Normalize()
	limit := int64(normalized.GetLimit())
	offset := int64(normalized.GetOffset())

	entries, total, err := h.leaderboardUseCase.GetLeaderboardPaginated(ctx, limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Create pagination metadata
	meta := response.NewPagination(normalized.GetOffset(), normalized.GetLimit(), total)
	response.SuccessWithMeta(c, entries, "Leaderboard retrieved successfully", meta)
}

// GetLeaderboardStream handles GET /leaderboard/stream via SSE for real-time delta updates
func (h *LeaderboardHandler) GetLeaderboardStream(c *gin.Context) {
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
			sendSSEResponse(c, entry, "Leaderboard entry updated")

		case <-ticker.C:
			// Send keep-alive comment
			_, _ = fmt.Fprintf(c.Writer, ": keep-alive\n\n")
			c.Writer.Flush()
		}
	}
}

// sendSSEResponse sends a standard API response via SSE
func sendSSEResponse(c *gin.Context, data interface{}, message string) {
	resp := response.Response{
		Success: true,
		Data:    data,
		Message: message,
	}
	messageBytes, _ := json.Marshal(resp)
	_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", messageBytes)
	c.Writer.Flush()
}

// SubmitScore handles score update
func (h *LeaderboardHandler) SubmitScore(c *gin.Context) {
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

	response.Success(c, gin.H{"user_id": userID, "score": req.Score}, "Score updated successfully")
}

// RegisterPublicRoutes registers public leaderboard routes (no auth required)
func (h *LeaderboardHandler) RegisterPublicRoutes(router *gin.RouterGroup) {
	leaderboard := router.Group("/leaderboard")
	{
		leaderboard.GET("", h.GetLeaderboardPaginated)
		leaderboard.GET("/stream", h.GetLeaderboardStream)
	}
}

// RegisterProtectedRoutes registers protected leaderboard routes (auth required)
func (h *LeaderboardHandler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	leaderboard := router.Group("/leaderboard")
	{
		leaderboard.PUT("/score", h.SubmitScore)
	}
}
