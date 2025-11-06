package dashboard

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(e *echo.Group) {
	dashboard := e.Group("/dashboard")
	dashboard.GET("/stats", h.GetStats, middleware.RequirePermission("ver-dashboard"))
	dashboard.GET("/activity/recent", h.GetRecentActivity, middleware.RequirePermission("ver-dashboard"))
	dashboard.GET("/exams/critical", h.GetCriticalExams, middleware.RequirePermission("ver-dashboard"))
}

func (h *Handler) GetStats(c echo.Context) error {
	stats, err := h.service.GetStats()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetRecentActivity(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 10
	}
	activities, err := h.service.GetRecentActivity(limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, activities)
}

func (h *Handler) GetCriticalExams(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 10
	}
	exams, err := h.service.GetCriticalExams(limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, exams)
}
