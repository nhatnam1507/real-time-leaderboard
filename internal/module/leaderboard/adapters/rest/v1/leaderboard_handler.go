// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/shared/request"

	"github.com/gin-gonic/gin"
)

const (
	// Keep-alive interval for SSE connections
	keepAliveInterval = 15 * time.Second
	// Initial leaderboard limit for SSE
	defaultLeaderboardLimit = 100
)

// LeaderboardHandler handles HTTP requests for leaderboards
type LeaderboardHandler struct {
	leaderboardUseCase *application.LeaderboardUseCase
}

// NewLeaderboardHandler creates a new leaderboard HTTP handler
func NewLeaderboardHandler(leaderboardUseCase *application.LeaderboardUseCase) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardUseCase: leaderboardUseCase,
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

	listReq := request.ListRequest{
		PaginationRequest: request.PaginationRequest{
			Limit:  limit,
			Offset: 0, // Always 0 for SSE
		},
	}

	// Create context for the request
	ctx := c.Request.Context()

	// Sync from PostgreSQL if needed (lazy loading)
	_ = h.leaderboardUseCase.SyncFromPostgres(ctx)

	// Get update channel from use case (handles Redis pub/sub)
	updateCh := h.leaderboardUseCase.WatchLeaderboard(ctx, &listReq)

	// Set up keep-alive ticker
	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	// Handle client disconnection
	notify := c.Writer.CloseNotify()

	// Stream events to client
	for {
		select {
		case <-notify:
			// Client disconnected
			return

		case leaderboard, ok := <-updateCh:
			if !ok {
				// Channel closed, connection ended
				return
			}

			// Send leaderboard update to client
			message, _ := json.Marshal(leaderboard)
			_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", message)
			c.Writer.Flush()

		case <-ticker.C:
			// Send keep-alive comment
			_, _ = fmt.Fprintf(c.Writer, ": keep-alive\n\n")
			c.Writer.Flush()
		}
	}
}

// RegisterRoutes registers leaderboard and score routes
func RegisterRoutes(router *gin.RouterGroup, leaderboardHandler *LeaderboardHandler, scoreHandler *ScoreHandler) {
	if leaderboardHandler != nil {
		router.GET("/leaderboard", leaderboardHandler.GetLeaderboard)
	}
	if scoreHandler != nil {
		router.PUT("/score", scoreHandler.SubmitScore)
	}
}
