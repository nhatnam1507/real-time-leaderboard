package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"real-time-leaderboard/internal/module/leaderboard/infrastructure/mocks"
	"real-time-leaderboard/internal/shared/logger"
)

func TestScoreUseCase_SubmitScore_WhenValidRequest_ShouldUpdateScoreAndBroadcastEntry(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockPersistenceRepo.EXPECT().
		UpsertScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		UpdateScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)
	mockCacheRepo.EXPECT().
		GetUserRank(ctx, "user-123").
		Return(int64(1), nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByIDs(ctx, []string{"user-123"}).
		Return(map[string]string{"user-123": "alice"}, nil).
		Times(1)

	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)
	mockBroadcastService.EXPECT().
		BroadcastEntryUpdate(ctx, gomock.Any()).
		Return(nil).
		Times(1)

	logger := logger.New("info", false)
	uc := NewScoreUseCase(mockPersistenceRepo, mockCacheRepo, mockUserRepo, mockBroadcastService, logger)

	req := SubmitScoreRequest{Score: 1000}

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SubmitScore(ctx, "user-123", req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
}

func TestScoreUseCase_SubmitScore_WhenPersistenceFails_ShouldReturnInternalError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockPersistenceRepo.EXPECT().
		UpsertScore(ctx, "user-123", int64(1000)).
		Return(errors.New("database error")).
		Times(1)

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewScoreUseCase(mockPersistenceRepo, mockCacheRepo, mockUserRepo, mockBroadcastService, logger)

	req := SubmitScoreRequest{Score: 1000}

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SubmitScore(ctx, "user-123", req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to update score")
	require.Contains(t, err.Error(), "database error")
}

func TestScoreUseCase_SubmitScore_WhenCacheUpdateFails_ShouldReturnNilError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockPersistenceRepo.EXPECT().
		UpsertScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		UpdateScore(ctx, "user-123", int64(1000)).
		Return(errors.New("redis error")).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewScoreUseCase(mockPersistenceRepo, mockCacheRepo, mockUserRepo, mockBroadcastService, logger)

	req := SubmitScoreRequest{Score: 1000}

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SubmitScore(ctx, "user-123", req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err) // Cache failure is non-critical, returns nil
}

func TestScoreUseCase_SubmitScore_WhenBroadcastFails_ShouldReturnNilError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockPersistenceRepo.EXPECT().
		UpsertScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		UpdateScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)
	mockCacheRepo.EXPECT().
		GetUserRank(ctx, "user-123").
		Return(int64(1), nil).
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		GetByIDs(ctx, []string{"user-123"}).
		Return(map[string]string{"user-123": "alice"}, nil).
		Times(1)

	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)
	mockBroadcastService.EXPECT().
		BroadcastEntryUpdate(ctx, gomock.Any()).
		Return(errors.New("broadcast error")).
		Times(1)

	logger := logger.New("info", false)
	uc := NewScoreUseCase(mockPersistenceRepo, mockCacheRepo, mockUserRepo, mockBroadcastService, logger)

	req := SubmitScoreRequest{Score: 1000}

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SubmitScore(ctx, "user-123", req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err) // Broadcast failure is non-critical, returns nil
}

func TestScoreUseCase_SubmitScore_WhenRankOutsideBroadcastRange_ShouldNotBroadcast(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPersistenceRepo := mocks.NewMockLeaderboardPersistenceRepository(ctrl)
	mockPersistenceRepo.EXPECT().
		UpsertScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)

	mockCacheRepo := mocks.NewMockLeaderboardCacheRepository(ctrl)
	mockCacheRepo.EXPECT().
		UpdateScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)
	mockCacheRepo.EXPECT().
		GetUserRank(ctx, "user-123").
		Return(int64(1500), nil). // Rank outside MaxBroadcastRank (1000)
		Times(1)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)
	// Should NOT be called since rank is outside broadcast range

	logger := logger.New("info", false)
	uc := NewScoreUseCase(mockPersistenceRepo, mockCacheRepo, mockUserRepo, mockBroadcastService, logger)

	req := SubmitScoreRequest{Score: 1000}

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SubmitScore(ctx, "user-123", req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
	// Broadcast should not be called for ranks outside MaxBroadcastRank
}
