// Package adapters provides adapters for the leaderboard module.
package adapters

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

// LeaderboardBroadcast handles broadcasting leaderboard updates to all connected clients
type LeaderboardBroadcast struct {
	leaderboardRepo application.LeaderboardRepository
	backupRepo      application.LeaderboardBackupRepository
	userRepo        application.UserRepository
	redisClient     *redis.Client
	logger          *logger.Logger

	// Topics (use domain constants)
	scoreTopic  string // domain.RedisScoreUpdateTopic (listens here)
	viewerTopic string // domain.RedisViewerUpdateTopic (publishes here)

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// NewLeaderboardBroadcast creates and starts broadcast service
func NewLeaderboardBroadcast(
	leaderboardRepo application.LeaderboardRepository,
	backupRepo application.LeaderboardBackupRepository,
	userRepo application.UserRepository,
	redisClient *redis.Client,
	logger *logger.Logger,
) *LeaderboardBroadcast {
	ctx, cancel := context.WithCancel(context.Background())

	broadcast := &LeaderboardBroadcast{
		leaderboardRepo: leaderboardRepo,
		backupRepo:      backupRepo,
		userRepo:        userRepo,
		redisClient:     redisClient,
		logger:          logger,
		scoreTopic:      domain.RedisScoreUpdateTopic,
		viewerTopic:     domain.RedisViewerUpdateTopic,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start in background
	go func() {
		if err := broadcast.Start(ctx); err != nil {
			broadcast.logger.Errorf(ctx, "Broadcast service error: %v", err)
		}
	}()

	return broadcast
}

// Start listens to score updates and broadcasts
func (b *LeaderboardBroadcast) Start(ctx context.Context) error {
	pubsub := b.redisClient.Subscribe(ctx, b.scoreTopic)
	defer func() {
		_ = pubsub.Close()
	}()

	// Generate unique instance ID for lock
	instanceID := b.getInstanceID()
	lockKey := domain.RedisBroadcastLockKey
	lockTTL := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-pubsub.Channel():
			if msg == nil {
				return nil
			}

			// Try to acquire distributed lock (only one instance processes)
			acquired, err := b.redisClient.SetNX(ctx, lockKey, instanceID, lockTTL).Result()
			if err != nil || !acquired {
				// Another instance is processing, skip
				continue
			}

			// We acquired the lock, process the update
			leaderboard := b.fetchFullLeaderboard(ctx)
			if leaderboard != nil {
				jsonData, err := json.Marshal(leaderboard)
				if err != nil {
					b.logger.Errorf(ctx, "Failed to marshal leaderboard: %v", err)
				} else {
					b.redisClient.Publish(ctx, b.viewerTopic, jsonData)
				}
			}

			// Release lock (optional, will expire anyway)
			b.redisClient.Del(ctx, lockKey)
		}
	}
}

// getInstanceID generates unique instance identifier
func (b *LeaderboardBroadcast) getInstanceID() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s:%d", hostname, os.Getpid())
}

// RegisterClient subscribes client to viewer topic
func (b *LeaderboardBroadcast) RegisterClient(ctx context.Context) <-chan *domain.Leaderboard {
	ch := make(chan *domain.Leaderboard, 1)

	// Subscribe to viewer topic
	pubsub := b.redisClient.Subscribe(ctx, b.viewerTopic)

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

				// Unmarshal full leaderboard
				var leaderboard domain.Leaderboard
				if err := json.Unmarshal([]byte(msg.Payload), &leaderboard); err != nil {
					b.logger.Warnf(ctx, "Failed to unmarshal leaderboard: %v", err)
					continue
				}

				// Send full leaderboard to channel
				select {
				case ch <- &leaderboard:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch
}

// fetchFullLeaderboard fetches full leaderboard (helper method)
func (b *LeaderboardBroadcast) fetchFullLeaderboard(ctx context.Context) *domain.Leaderboard {
	// Fetch with large limit (e.g., 1000) to get full leaderboard
	entries, err := b.leaderboardRepo.GetTopPlayers(ctx, 1000, 0)
	if err != nil {
		b.logger.Errorf(ctx, "Failed to fetch full leaderboard: %v", err)
		return nil
	}

	total, err := b.leaderboardRepo.GetTotalPlayers(ctx)
	if err != nil {
		b.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(entries))
	}

	// Enrich entries with usernames
	if err := b.enrichEntriesWithUsernames(ctx, entries); err != nil {
		b.logger.Warnf(ctx, "Failed to enrich entries with usernames: %v", err)
		// Continue even if username enrichment fails
	}

	return &domain.Leaderboard{
		Entries: entries,
		Total:   total,
	}
}

// enrichEntriesWithUsernames enriches leaderboard entries with usernames
func (b *LeaderboardBroadcast) enrichEntriesWithUsernames(ctx context.Context, entries []domain.LeaderboardEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Extract user IDs
	userIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		userIDs = append(userIDs, entry.UserID)
	}

	// Fetch usernames in batch
	usernames, err := b.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return err
	}

	// Enrich entries with usernames
	for i := range entries {
		if username, ok := usernames[entries[i].UserID]; ok {
			entries[i].Username = username
		} else {
			// User not found - log warning and use empty string
			b.logger.Warnf(ctx, "Username not found for user ID: %s", entries[i].UserID)
			entries[i].Username = ""
		}
	}

	return nil
}

// Stop stops the broadcast service
func (b *LeaderboardBroadcast) Stop() {
	b.cancel()
}
