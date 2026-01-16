// Package application provides use cases for the leaderboard module.
package application

//go:generate mockgen -source=broadcast_service.go -destination=../mocks/broadcast_service_mock.go -package=mocks BroadcastService

import (
	"context"

	"real-time-leaderboard/internal/module/leaderboard/domain"
)

// BroadcastService defines the interface for broadcasting leaderboard entry delta updates
type BroadcastService interface {
	BroadcastEntryUpdate(ctx context.Context, entry *domain.LeaderboardEntry) error
	SubscribeToEntryUpdates(ctx context.Context) (<-chan *domain.LeaderboardEntry, error)
}
