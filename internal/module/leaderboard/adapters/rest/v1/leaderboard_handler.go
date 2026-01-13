// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/middleware"
	"real-time-leaderboard/internal/shared/response"
	"real-time-leaderboard/internal/shared/validator"

	"github.com/gin-gonic/gin"
)

const (
	// Keep-alive interval for SSE connections
	keepAliveInterval = 15 * time.Second
	// Initial leaderboard limit for SSE
	defaultLeaderboardLimit = 100
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

// GetLeaderboard handles getting the leaderboard via SSE
// Handler only manages SSE connection lifecycle, business logic is in use case
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Parse limit from query params (offset is not supported for SSE, always 0)
	limit := defaultLeaderboardLimit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > defaultLeaderboardLimit {
				limit = defaultLeaderboardLimit
			}
		}
	}

	// Create context for the request
	ctx := c.Request.Context()

	// Sync from PostgreSQL to Redis if needed (lazy loading)
	_ = h.leaderboardUseCase.SyncFromPostgres(ctx)

	// Subscribe to leaderboard updates (gets channel for full leaderboard)
	updateCh, err := h.leaderboardUseCase.SubscribeToLeaderboardUpdates(ctx)
	if err != nil {
		// If subscription fails, create a closed channel
		closedCh := make(chan *domain.Leaderboard)
		close(closedCh)
		updateCh = closedCh
	}

	// Fetch initial leaderboard
	fullLeaderboard, err := h.leaderboardUseCase.GetFullLeaderboard(ctx)
	if err != nil {
		// If we can't get initial leaderboard, still try to stream updates
		fullLeaderboard = &domain.Leaderboard{Entries: []domain.LeaderboardEntry{}, Total: 0}
	}

	// Extract limit from full leaderboard
	filteredLeaderboard := extractLimit(fullLeaderboard, limit)

	// Send initial leaderboard via SSE using standard response format
	sendSSEResponse(c, filteredLeaderboard, "Leaderboard retrieved successfully")

	// Set up keep-alive ticker
	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	// Handle client disconnection
	notify := c.Writer.CloseNotify()

	// Keep connection, push updates from broadcaster
	for {
		select {
		case <-notify:
			// Client disconnected
			return

		case fullLeaderboard, ok := <-updateCh:
			if !ok {
				// Channel closed, connection ended
				return
			}

			// Extract limit from full leaderboard
			filtered := extractLimit(fullLeaderboard, limit)

			// Send leaderboard update to client using standard response format
			sendSSEResponse(c, filtered, "Leaderboard updated")

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

// extractLimit extracts the requested limit from full leaderboard
func extractLimit(leaderboard *domain.Leaderboard, limit int) *domain.Leaderboard {
	if leaderboard == nil {
		return &domain.Leaderboard{Entries: []domain.LeaderboardEntry{}, Total: 0}
	}

	if limit >= len(leaderboard.Entries) {
		return leaderboard
	}

	return &domain.Leaderboard{
		Entries: leaderboard.Entries[:limit],
		Total:   int64(limit),
	}
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
		leaderboard.GET("/stream", h.GetLeaderboard)
	}
}

// RegisterProtectedRoutes registers protected leaderboard routes (auth required)
func (h *LeaderboardHandler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	leaderboard := router.Group("/leaderboard")
	{
		leaderboard.PUT("/score", h.SubmitScore)
	}
}
