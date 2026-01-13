package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"real-time-leaderboard/internal/module/leaderboard/mocks"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

func TestScoreUseCase_SubmitScore_WhenValidRequest_ShouldUpdateScoreAndPublishNotification(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackupRepo := mocks.NewMockLeaderboardBackupRepository(ctrl)
	mockBackupRepo.EXPECT().
		UpsertScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)

	mockLeaderboardRepo := mocks.NewMockLeaderboardRepository(ctrl)
	mockLeaderboardRepo.EXPECT().
		UpdateScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)

	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)
	mockBroadcastService.EXPECT().
		PublishScoreUpdate(ctx).
		Return(nil).
		Times(1)

	logger := logger.New("info", false)
	uc := NewScoreUseCase(mockBackupRepo, mockLeaderboardRepo, mockBroadcastService, logger)

	req := SubmitScoreRequest{Score: 1000}

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SubmitScore(ctx, "user-123", req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err)
}

func TestScoreUseCase_SubmitScore_WhenBackupFails_ShouldReturnInternalError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackupRepo := mocks.NewMockLeaderboardBackupRepository(ctrl)
	mockBackupRepo.EXPECT().
		UpsertScore(ctx, "user-123", int64(1000)).
		Return(errors.New("database error")).
		Times(1)

	mockLeaderboardRepo := mocks.NewMockLeaderboardRepository(ctrl)
	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewScoreUseCase(mockBackupRepo, mockLeaderboardRepo, mockBroadcastService, logger)

	req := SubmitScoreRequest{Score: 1000}

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SubmitScore(ctx, "user-123", req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Error(t, err)

	var appErr *response.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, response.CodeInternal, appErr.Code)
	require.Contains(t, appErr.Message, "Failed to update score")
}

func TestScoreUseCase_SubmitScore_WhenRedisUpdateFails_ShouldReturnNilError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackupRepo := mocks.NewMockLeaderboardBackupRepository(ctrl)
	mockBackupRepo.EXPECT().
		UpsertScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)

	mockLeaderboardRepo := mocks.NewMockLeaderboardRepository(ctrl)
	mockLeaderboardRepo.EXPECT().
		UpdateScore(ctx, "user-123", int64(1000)).
		Return(errors.New("redis error")).
		Times(1)

	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)

	logger := logger.New("info", false)
	uc := NewScoreUseCase(mockBackupRepo, mockLeaderboardRepo, mockBroadcastService, logger)

	req := SubmitScoreRequest{Score: 1000}

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SubmitScore(ctx, "user-123", req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err) // Redis failure is non-critical, returns nil
}

func TestScoreUseCase_SubmitScore_WhenBroadcastFails_ShouldReturnNilError(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackupRepo := mocks.NewMockLeaderboardBackupRepository(ctrl)
	mockBackupRepo.EXPECT().
		UpsertScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)

	mockLeaderboardRepo := mocks.NewMockLeaderboardRepository(ctrl)
	mockLeaderboardRepo.EXPECT().
		UpdateScore(ctx, "user-123", int64(1000)).
		Return(nil).
		Times(1)

	mockBroadcastService := mocks.NewMockBroadcastService(ctrl)
	mockBroadcastService.EXPECT().
		PublishScoreUpdate(ctx).
		Return(errors.New("broadcast error")).
		Times(1)

	logger := logger.New("info", false)
	uc := NewScoreUseCase(mockBackupRepo, mockLeaderboardRepo, mockBroadcastService, logger)

	req := SubmitScoreRequest{Score: 1000}

	// ── Act ─────────────────────────────────────────────────────────────
	err := uc.SubmitScore(ctx, "user-123", req)

	// ── Assert ──────────────────────────────────────────────────────────
	require.NoError(t, err) // Broadcast failure is non-critical, returns nil
}
