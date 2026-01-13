// Package broadcast provides broadcast service implementations for the leaderboard module.
package broadcast

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"

	"github.com/redis/go-redis/v9"
)

// RedisBroadcastService implements BroadcastService using Redis pub/sub
type RedisBroadcastService struct {
	client     *redis.Client
	logger     *logger.Logger
	scoreTopic string
	viewerTopic string
	lockKey    string
	instanceID string
}

// NewRedisBroadcastService creates a new Redis broadcast service
func NewRedisBroadcastService(
	client *redis.Client,
	logger *logger.Logger,
) application.BroadcastService {
	hostname, _ := os.Hostname()
	return &RedisBroadcastService{
		client:      client,
		logger:      logger,
		scoreTopic:  domain.RedisScoreUpdateTopic,
		viewerTopic: domain.RedisViewerUpdateTopic,
		lockKey:     domain.RedisBroadcastLockKey,
		instanceID:  fmt.Sprintf("%s:%d", hostname, os.Getpid()),
	}
}

// PublishScoreUpdate publishes a score update notification
func (s *RedisBroadcastService) PublishScoreUpdate(ctx context.Context) error {
	return s.client.Publish(ctx, s.scoreTopic, "updated").Err()
}

// SubscribeToScoreUpdates subscribes to score update notifications with distributed locking
func (s *RedisBroadcastService) SubscribeToScoreUpdates(ctx context.Context) (<-chan struct{}, error) {
	pubsub := s.client.Subscribe(ctx, s.scoreTopic)
	ch := make(chan struct{}, 1)

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

				acquired, err := s.client.SetNX(ctx, s.lockKey, s.instanceID, 5*time.Second).Result()
				if err != nil || !acquired {
					continue
				}

				select {
				case ch <- struct{}{}:
				case <-ctx.Done():
					return
				}

				s.client.Del(ctx, s.lockKey)
			}
		}
	}()

	return ch, nil
}

// BroadcastLeaderboard broadcasts the enriched leaderboard to all subscribers
func (s *RedisBroadcastService) BroadcastLeaderboard(ctx context.Context, leaderboard *domain.Leaderboard) error {
	jsonData, err := json.Marshal(leaderboard)
	if err != nil {
		return fmt.Errorf("failed to marshal leaderboard: %w", err)
	}

	return s.client.Publish(ctx, s.viewerTopic, jsonData).Err()
}

// SubscribeToLeaderboardUpdates subscribes to leaderboard update broadcasts
func (s *RedisBroadcastService) SubscribeToLeaderboardUpdates(ctx context.Context) (<-chan *domain.Leaderboard, error) {
	pubsub := s.client.Subscribe(ctx, s.viewerTopic)
	ch := make(chan *domain.Leaderboard, 1)

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

				var leaderboard domain.Leaderboard
				if err := json.Unmarshal([]byte(msg.Payload), &leaderboard); err != nil {
					s.logger.Warnf(ctx, "Failed to unmarshal leaderboard: %v", err)
					continue
				}

				select {
				case ch <- &leaderboard:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}
