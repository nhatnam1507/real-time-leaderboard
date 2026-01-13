// Package application provides use cases for the leaderboard module.
package application

import (
	"context"

	"real-time-leaderboard/internal/module/leaderboard/domain"
)

// BroadcastService defines the interface for broadcasting leaderboard updates
type BroadcastService interface {
	PublishScoreUpdate(ctx context.Context) error
	SubscribeToScoreUpdates(ctx context.Context) (<-chan struct{}, error)
	BroadcastLeaderboard(ctx context.Context, leaderboard *domain.Leaderboard) error
	SubscribeToLeaderboardUpdates(ctx context.Context) (<-chan *domain.Leaderboard, error)
}
