// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/shared/request"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
	redisClient        *redis.Client
}

// NewLeaderboardHandler creates a new leaderboard HTTP handler
func NewLeaderboardHandler(leaderboardUseCase *application.LeaderboardUseCase, redisClient *redis.Client) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardUseCase: leaderboardUseCase,
		redisClient:        redisClient,
	}
}

// GetLeaderboard handles getting the leaderboard via SSE
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Send initial leaderboard
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	listReq := &request.ListRequest{
		PaginationRequest: request.PaginationRequest{
			Limit:  defaultLeaderboardLimit,
			Offset: 0,
		},
	}
	leaderboard, err := h.leaderboardUseCase.GetLeaderboard(ctx, listReq)
	cancel()

	if err == nil && leaderboard != nil {
		message, _ := json.Marshal(leaderboard)
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", message)
		c.Writer.Flush()
	}

	// Subscribe to Redis pub/sub for leaderboard updates
	pubsub := h.redisClient.Subscribe(c.Request.Context(), "leaderboard:updates")
	defer func() {
		_ = pubsub.Close()
	}()

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

		case <-pubsub.Channel():
			// Redis pub/sub message received, fetch fresh leaderboard
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			listReq := &request.ListRequest{
				PaginationRequest: request.PaginationRequest{
					Limit:  defaultLeaderboardLimit,
					Offset: 0,
				},
			}
			leaderboard, err := h.leaderboardUseCase.GetLeaderboard(ctx, listReq)
			cancel()

			if err == nil && leaderboard != nil {
				message, _ := json.Marshal(leaderboard)
				_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", message)
				c.Writer.Flush()
			}

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
		router.POST("/score", scoreHandler.SubmitScore)
	}
}
