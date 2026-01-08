// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/shared/request"
	"real-time-leaderboard/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	// Keep-alive interval for SSE connections
	keepAliveInterval = 15 * time.Second
	// Initial leaderboard limit for SSE
	defaultLeaderboardLimit = 100
)

// Handler handles HTTP requests for leaderboards
type Handler struct {
	leaderboardUseCase *application.LeaderboardUseCase
	redisClient        *redis.Client
}

// NewHandler creates a new leaderboard HTTP handler
func NewHandler(leaderboardUseCase *application.LeaderboardUseCase, redisClient *redis.Client) *Handler {
	return &Handler{
		leaderboardUseCase: leaderboardUseCase,
		redisClient:        redisClient,
	}
}

// GetGlobalLeaderboard handles getting the global leaderboard via SSE
func (h *Handler) GetGlobalLeaderboard(c *gin.Context) {
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
	leaderboard, err := h.leaderboardUseCase.GetGlobalLeaderboard(ctx, listReq)
	cancel()

	if err == nil && leaderboard != nil {
		message, _ := json.Marshal(leaderboard)
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", message)
		c.Writer.Flush()
	}

	// Subscribe to Redis pub/sub for leaderboard updates
	pubsub := h.redisClient.Subscribe(c.Request.Context(), "leaderboard:updates:global")
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
			leaderboard, err := h.leaderboardUseCase.GetGlobalLeaderboard(ctx, listReq)
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

// GetGameLeaderboard handles getting a game-specific leaderboard via SSE
func (h *Handler) GetGameLeaderboard(c *gin.Context) {
	gameID := c.Param("game_id")

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
	leaderboard, err := h.leaderboardUseCase.GetGameLeaderboard(ctx, gameID, listReq)
	cancel()

	if err == nil && leaderboard != nil {
		message, _ := json.Marshal(leaderboard)
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", message)
		c.Writer.Flush()
	}

	// Subscribe to Redis pub/sub for leaderboard updates
	channel := fmt.Sprintf("leaderboard:updates:%s", gameID)
	pubsub := h.redisClient.Subscribe(c.Request.Context(), channel)
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
			leaderboard, err := h.leaderboardUseCase.GetGameLeaderboard(ctx, gameID, listReq)
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

// GetUserRank handles getting a user's rank
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
		leaderboard.GET("/global", h.GetGlobalLeaderboard)      // SSE stream for global leaderboard
		leaderboard.GET("/game/:game_id", h.GetGameLeaderboard) // SSE stream for game-specific leaderboard
		leaderboard.GET("/rank/:user_id", h.GetUserRank)        // Regular REST endpoint for user rank
	}
}
