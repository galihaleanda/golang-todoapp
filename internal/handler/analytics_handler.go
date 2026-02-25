package handler

import (
	"fmt"
	"time"

	"github.com/galihaleanda/todo-app/internal/middleware"
	"github.com/galihaleanda/todo-app/internal/service"
	"github.com/galihaleanda/todo-app/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AnalyticsHandler exposes analytics endpoints.
type AnalyticsHandler struct {
	analyticsSvc *service.AnalyticsService
}

// NewAnalyticsHandler creates an AnalyticsHandler.
func NewAnalyticsHandler(analyticsSvc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsSvc: analyticsSvc}
}

// Dashboard godoc
// @Summary Get productivity dashboard
// @Tags analytics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Envelope{data=domain.AnalyticsDashboard}
// @Router /analytics/dashboard [get]
func (h *AnalyticsHandler) Dashboard(c *gin.Context) {
	dash, err := h.analyticsSvc.GetDashboard(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, dash)
}

// DailyStats godoc
// @Summary Get daily productivity stats for a custom date range
// @Tags analytics
// @Security BearerAuth
// @Produce json
// @Param from query string true "Start date (YYYY-MM-DD)"
// @Param to query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.Envelope{data=[]domain.DailyStats}
// @Router /analytics/daily [get]
func (h *AnalyticsHandler) DailyStats(c *gin.Context) {
	from, err := parseDate(c.Query("from"))
	if err != nil {
		response.BadRequest(c, "INVALID_DATE", "from must be YYYY-MM-DD", nil)
		return
	}

	to, err := parseDate(c.Query("to"))
	if err != nil {
		response.BadRequest(c, "INVALID_DATE", "to must be YYYY-MM-DD", nil)
		return
	}

	stats, err := h.analyticsSvc.GetDailyStats(c.Request.Context(), middleware.CurrentUserID(c), from, to)
	if err != nil {
		response.BadRequest(c, "INVALID_RANGE", err.Error(), nil)
		return
	}

	response.OK(c, stats)
}

// --- shared helpers ---

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	return uuid.Parse(c.Param(param))
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	return time.Parse("2006-01-02", s)
}
