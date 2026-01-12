// Package application provides use cases for the leaderboard module.
package application

import (
	"context"
	"time"

	"real-time-leaderboard/internal/module/leaderboard/domain"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/request"
	"real-time-leaderboard/internal/shared/response"

	"github.com/redis/go-redis/v9"
)

const (
	updateChannel = "leaderboard:updates"
)

// LeaderboardUseCase handles leaderboard use cases
type LeaderboardUseCase struct {
	leaderboardRepo domain.LeaderboardRepository
	backupRepo      domain.LeaderboardBackupRepository
	redisClient     *redis.Client
	logger          *logger.Logger
}

// NewLeaderboardUseCase creates a new leaderboard use case
func NewLeaderboardUseCase(
	leaderboardRepo domain.LeaderboardRepository,
	backupRepo domain.LeaderboardBackupRepository,
	redisClient *redis.Client,
	l *logger.Logger,
) *LeaderboardUseCase {
	return &LeaderboardUseCase{
		leaderboardRepo: leaderboardRepo,
		backupRepo:      backupRepo,
		redisClient:     redisClient,
		logger:          l,
	}
}

// SyncFromPostgres syncs all leaderboard entries from PostgreSQL to Redis
// This is called lazily when Redis is detected to be empty
func (uc *LeaderboardUseCase) SyncFromPostgres(ctx context.Context) error {
	total, err := uc.leaderboardRepo.GetTotalPlayers(ctx)
	if err != nil || total > 0 {
		return err
	}

	scores, err := uc.backupRepo.GetAllScores(ctx)
	if err != nil {
		return err
	}

	for _, score := range scores {
		if err := uc.leaderboardRepo.UpdateScore(ctx, score.UserID, score.Point); err != nil {
			uc.logger.Warnf(ctx, "Failed to sync score for user %s: %v", score.UserID, err)
		}
	}

	if len(scores) > 0 {
		uc.logger.Infof(ctx, "Synced %d leaderboard entries from PostgreSQL to Redis", len(scores))
	}
	return nil
}

// getLeaderboard retrieves the leaderboard
// This is a private helper method used internally by WatchLeaderboard
func (uc *LeaderboardUseCase) getLeaderboard(ctx context.Context, listReq *request.ListRequest) (*domain.Leaderboard, error) {
	limit := int64(listReq.GetLimit())
	offset := int64(listReq.GetOffset())

	entries, err := uc.leaderboardRepo.GetTopPlayers(ctx, limit, offset)
	if err != nil {
		uc.logger.Errorf(ctx, "Failed to get leaderboard: %v", err)
		return nil, response.NewInternalError("Failed to retrieve leaderboard", err)
	}

	total, err := uc.leaderboardRepo.GetTotalPlayers(ctx)
	if err != nil {
		uc.logger.Warnf(ctx, "Failed to get total players: %v", err)
		total = int64(len(entries))
	}

	return &domain.Leaderboard{
		Entries: entries,
		Total:   total,
	}, nil
}

// WatchLeaderboard returns a channel that streams leaderboard updates
// It handles Redis pub/sub subscription and sends leaderboard data on updates
func (uc *LeaderboardUseCase) WatchLeaderboard(ctx context.Context, listReq *request.ListRequest) <-chan *domain.Leaderboard {
	ch := make(chan *domain.Leaderboard, 1)

	go func() {
		defer close(ch)

		// Helper to send leaderboard to channel
		sendLeaderboard := func(ctx context.Context, listReq *request.ListRequest) bool {
			leaderboard, err := uc.getLeaderboard(ctx, listReq)
			if err != nil || leaderboard == nil {
				return false
			}
			select {
			case ch <- leaderboard:
				return true
			case <-ctx.Done():
				return false
			}
		}

		// Send initial leaderboard
		if !sendLeaderboard(ctx, listReq) {
			return
		}

		// Subscribe to Redis pub/sub
		pubsub := uc.redisClient.Subscribe(ctx, updateChannel)
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
				updateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				sendLeaderboard(updateCtx, listReq)
				cancel()
			}
		}
	}()

	return ch
}
