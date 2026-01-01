package http

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"real-time-leaderboard/internal/module/report/application"
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
// @Description Get a report of top players, optionally filtered by game and date range
// @Tags reports
// @Accept json
// @Produce json
// @Param game_id query string false "Game ID (empty for global)"
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param limit query int false "Number of top players" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/reports/top-players [get]
func (h *Handler) GetTopPlayersReport(c *gin.Context) {
	gameID := c.DefaultQuery("game_id", "global")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

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
		GameID:    gameID,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     limit,
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
