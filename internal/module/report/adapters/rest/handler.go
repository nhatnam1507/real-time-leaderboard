// Package rest provides REST API handlers for the report module.
package rest

import (
	"time"

	"github.com/gin-gonic/gin"
	"real-time-leaderboard/internal/module/report/application"
	"real-time-leaderboard/internal/shared/request"
	"real-time-leaderboard/internal/shared/response"
)

// Handler handles HTTP requests for reports
type Handler struct {
	reportUseCase *application.ReportUseCase
}

// NewHandler creates a new report HTTP handler
func NewHandler(reportUseCase *application.ReportUseCase) *Handler {
	return &Handler{
		reportUseCase: reportUseCase,
	}
}

// GetTopPlayersReport handles getting top players report
// @Summary Get top players report
// @Description Get a paginated report of top players, optionally filtered by game and date range. Uses Redis for current data or PostgreSQL for historical data when date range is provided.
// @Tags reports
// @Accept json
// @Produce json
// @Param game_id query string false "Game ID (empty or 'global' for global leaderboard)" example(game1)
// @Param start_date query string false "Start date in RFC3339 format" example(2024-01-01T00:00:00Z)
// @Param end_date query string false "End date in RFC3339 format" example(2024-01-31T23:59:59Z)
// @Param limit query int false "Number of top players" default(10) minimum(1) maximum(100) example(10)
// @Param offset query int false "Number of results to skip" default(0) minimum(0) example(0)
// @Success 200 {object} response.Response "Report generated successfully"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/reports/top-players [get]
func (h *Handler) GetTopPlayersReport(c *gin.Context) {
	var listReq request.ListRequest
	if err := listReq.FromGinContext(c); err != nil {
		response.Error(c, err)
		return
	}

	if err := listReq.Validate(); err != nil {
		response.Error(c, err)
		return
	}

	gameID := c.DefaultQuery("game_id", "global")

	var startDate, endDate *time.Time

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	req := application.GetTopPlayersReportRequest{
		ListRequest: listReq,
		GameID:      gameID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	report, err := h.reportUseCase.GetTopPlayersReport(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, report, "Report generated successfully")
}

// RegisterRoutes registers report routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	reports := router.Group("/reports")
	{
		reports.GET("/top-players", h.GetTopPlayersReport)
	}
}
