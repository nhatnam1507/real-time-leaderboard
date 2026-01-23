// Package v1 provides REST API v1 handlers for the leaderboard module.
package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/module/leaderboard/domain"
	lbmocks "real-time-leaderboard/internal/module/leaderboard/adapters/mocks"
	"real-time-leaderboard/internal/shared/logger"
	"real-time-leaderboard/internal/shared/response"
)

func TestLeaderboardHandler_GetLeaderboard_WhenValidQuery_ShouldReturn200WithEntriesAndMeta(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLB := lbmocks.NewMockLeaderboardUseCase(ctrl)
	mockScore := lbmocks.NewMockScoreUseCase(ctrl)

	mockLB.EXPECT().
		SyncFromPostgres(gomock.Any()).
		Return(nil).
		Times(1)
	mockLB.EXPECT().
		GetLeaderboard(gomock.Any(), int64(10), int64(0)).
		Return(
			[]domain.LeaderboardEntry{
				{UserID: "user-1", Username: "alice", Score: 1000, Rank: 1},
				{UserID: "user-2", Username: "bob", Score: 500, Rank: 2},
			},
			int64(2),
			nil,
		).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/leaderboard?limit=10&offset=0", nil)

	h := NewLeaderboardHandler(mockLB, mockScore, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.GetLeaderboard(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusOK, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.True(t, body.Success)
	require.Equal(t, "Leaderboard retrieved successfully", body.Message)
	require.NotNil(t, body.Data)
	require.NotNil(t, body.Meta)
}

func TestLeaderboardHandler_GetLeaderboard_WhenInvalidPagination_ShouldReturn400(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLB := lbmocks.NewMockLeaderboardUseCase(ctrl)
	mockScore := lbmocks.NewMockScoreUseCase(ctrl)
	mockLB.EXPECT().SyncFromPostgres(gomock.Any()).Times(0)
	mockLB.EXPECT().GetLeaderboard(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/leaderboard?limit=0&offset=0", nil)

	h := NewLeaderboardHandler(mockLB, mockScore, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.GetLeaderboard(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusBadRequest, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.NotNil(t, body.Error)
	require.Equal(t, string(response.CodeValidation), body.Error.Code)
}

func TestLeaderboardHandler_GetLeaderboard_WhenUseCaseReturnsError_ShouldReturn500(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLB := lbmocks.NewMockLeaderboardUseCase(ctrl)
	mockScore := lbmocks.NewMockScoreUseCase(ctrl)

	mockLB.EXPECT().
		SyncFromPostgres(gomock.Any()).
		Return(nil).
		Times(1)
	mockLB.EXPECT().
		GetLeaderboard(gomock.Any(), int64(10), int64(0)).
		Return(nil, int64(0), errUseCase).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/leaderboard?limit=10&offset=0", nil)

	h := NewLeaderboardHandler(mockLB, mockScore, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.GetLeaderboard(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusInternalServerError, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.Equal(t, string(response.CodeInternal), body.Error.Code)
}

func TestLeaderboardHandler_SubmitScore_WhenUserIDInContextAndValidBody_ShouldReturn200(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLB := lbmocks.NewMockLeaderboardUseCase(ctrl)
	mockScore := lbmocks.NewMockScoreUseCase(ctrl)
	req := application.SubmitScoreRequest{Score: 1500}
	mockScore.EXPECT().
		SubmitScore(gomock.Any(), "user-123", req).
		Return(nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/leaderboard/score", bytes.NewBufferString(`{"score":1500}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "user-123")

	h := NewLeaderboardHandler(mockLB, mockScore, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.SubmitScore(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusOK, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.True(t, body.Success)
	require.Equal(t, "Score updated successfully", body.Message)
}

func TestLeaderboardHandler_SubmitScore_WhenUserIDNotInContext_ShouldReturn401(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLB := lbmocks.NewMockLeaderboardUseCase(ctrl)
	mockScore := lbmocks.NewMockScoreUseCase(ctrl)
	mockScore.EXPECT().SubmitScore(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/leaderboard/score", bytes.NewBufferString(`{"score":1000}`))
	c.Request.Header.Set("Content-Type", "application/json")
	// do not set user_id

	h := NewLeaderboardHandler(mockLB, mockScore, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.SubmitScore(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusUnauthorized, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.Equal(t, string(response.CodeUnauthorized), body.Error.Code)
	require.Contains(t, body.Error.Message, "User ID not found")
}

func TestLeaderboardHandler_SubmitScore_WhenMissingRequiredFields_ShouldReturn400(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLB := lbmocks.NewMockLeaderboardUseCase(ctrl)
	mockScore := lbmocks.NewMockScoreUseCase(ctrl)
	mockScore.EXPECT().SubmitScore(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/leaderboard/score", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "user-123")

	h := NewLeaderboardHandler(mockLB, mockScore, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.SubmitScore(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusBadRequest, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.NotNil(t, body.Error)
	require.Equal(t, string(response.CodeValidation), body.Error.Code)
}

func TestLeaderboardHandler_SubmitScore_WhenUseCaseReturnsError_ShouldReturn500(t *testing.T) {
	// ── Arrange ────────────────────────────────────────────────────────
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLB := lbmocks.NewMockLeaderboardUseCase(ctrl)
	mockScore := lbmocks.NewMockScoreUseCase(ctrl)
	req := application.SubmitScoreRequest{Score: 1000}
	mockScore.EXPECT().
		SubmitScore(gomock.Any(), "user-123", req).
		Return(errUseCase).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/leaderboard/score", bytes.NewBufferString(`{"score":1000}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "user-123")

	h := NewLeaderboardHandler(mockLB, mockScore, logger.New("info", false))

	// ── Act ─────────────────────────────────────────────────────────────
	h.SubmitScore(c)

	// ── Assert ──────────────────────────────────────────────────────────
	require.Equal(t, http.StatusInternalServerError, w.Code)
	var body response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.Equal(t, string(response.CodeInternal), body.Error.Code)
}

// errUseCase is a sentinel for use case errors that get mapped to internal API error.
var errUseCase = errors.New("internal error")
