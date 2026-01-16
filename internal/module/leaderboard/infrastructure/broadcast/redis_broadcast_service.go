// Package broadcast provides broadcast service implementations for the leaderboard module.
package broadcast

import (
	"context"
	"encoding/json"
	"fmt"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"

	"github.com/redis/go-redis/v9"
)

// RedisBroadcastService implements BroadcastService using Redis pub/sub
type RedisBroadcastService struct {
	client     *redis.Client
	logger     *logger.Logger
	viewerTopic string
}

// NewRedisBroadcastService creates a new Redis broadcast service
func NewRedisBroadcastService(
	client *redis.Client,
	logger *logger.Logger,
) application.BroadcastService {
	return &RedisBroadcastService{
		client:      client,
		logger:      logger,
		viewerTopic: domain.RedisViewerUpdateTopic,
	}
}

// BroadcastEntryUpdate broadcasts a leaderboard entry delta update to all subscribers
func (s *RedisBroadcastService) BroadcastEntryUpdate(ctx context.Context, entry *domain.LeaderboardEntry) error {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	return s.client.Publish(ctx, s.viewerTopic, jsonData).Err()
}

// SubscribeToEntryUpdates subscribes to leaderboard entry delta update broadcasts
func (s *RedisBroadcastService) SubscribeToEntryUpdates(ctx context.Context) (<-chan *domain.LeaderboardEntry, error) {
	pubsub := s.client.Subscribe(ctx, s.viewerTopic)
	ch := make(chan *domain.LeaderboardEntry, 1)

	go func() {
		defer close(ch)
		defer func() {
			_ = pubsub.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-pubsub.Channel():
				if msg == nil {
					return
				}

				var entry domain.LeaderboardEntry
				if err := json.Unmarshal([]byte(msg.Payload), &entry); err != nil {
					s.logger.Warnf(ctx, "Failed to unmarshal entry: %v", err)
					continue
				}

				select {
				case ch <- &entry:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}
