// Package v1 provides REST API v1 handlers for the report module.
package v1

import (
	"time"

	"real-time-leaderboard/internal/module/report/application"
	"real-time-leaderboard/internal/shared/request"
	"real-time-leaderboard/internal/shared/response"

	"github.com/gin-gonic/gin"
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
