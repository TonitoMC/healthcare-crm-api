package schedule

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/schedule/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Handler exposes HTTP endpoints for schedule operations.
type Handler struct {
	service Service
}

// NewHandler constructs a new ScheduleHandler.
func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes mounts /schedule routes under the provided Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group) {
	scheduleGroup := g.Group("/schedule", ErrorMiddleware())

	// Read operations
	scheduleGroup.GET("/working-hours", h.GetWorkingHours)
	scheduleGroup.GET("/special-hours", h.GetSpecialHoursBetween)
	scheduleGroup.GET("/effective/day/:date", h.GetEffectiveDay)
	scheduleGroup.GET("/effective/range", h.GetEffectiveRange)

	// Write operations
	scheduleGroup.POST("/working-hours", h.UpdateWorkDay)
	scheduleGroup.POST("/special-hours", h.AddSpecialDay)
	scheduleGroup.DELETE("/special-hours/:date", h.DeleteSpecialDay)
}

// GET /schedule/working-hours
func (h *Handler) GetWorkingHours(c echo.Context) error {
	data, err := h.service.GetWorkingHours()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, data)
}

// GET /schedule/special-hours?start=YYYY-MM-DD&end=YYYY-MM-DD
func (h *Handler) GetSpecialHoursBetween(c echo.Context) error {
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return appErr.Wrap("Schedule.GetSpecialHoursBetween.ParseStart", appErr.ErrInvalidInput, err)
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return appErr.Wrap("Schedule.GetSpecialHoursBetween.ParseEnd", appErr.ErrInvalidInput, err)
	}

	data, err := h.service.GetSpecialHoursBetween(start, end)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, data)
}

// GET /schedule/effective/day/:date
func (h *Handler) GetEffectiveDay(c echo.Context) error {
	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return appErr.Wrap("Schedule.GetEffectiveDay.Parse", appErr.ErrInvalidInput, err)
	}

	eff, err := h.service.GetEffectiveDay(date)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, eff)
}

// GET /schedule/effective/range?start=YYYY-MM-DD&end=YYYY-MM-DD
func (h *Handler) GetEffectiveRange(c echo.Context) error {
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return appErr.Wrap("Schedule.GetEffectiveRange.ParseStart", appErr.ErrInvalidInput, err)
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return appErr.Wrap("Schedule.GetEffectiveRange.ParseEnd", appErr.ErrInvalidInput, err)
	}

	data, err := h.service.GetEffectiveRange(start, end)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, data)
}

// POST /schedule/working-hours
func (h *Handler) UpdateWorkDay(c echo.Context) error {
	var req models.CreateWorkDayRequest
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("Schedule.UpdateWorkDay.Bind", appErr.ErrInvalidInput, err)
	}

	active := len(req.Ranges) > 0

	workDay := models.WorkDay{
		DayOfWeek: req.DayOfWeek,
		Active:    active,
	}

	for _, r := range req.Ranges {
		start, err1 := time.Parse("15:04", r.Start)
		end, err2 := time.Parse("15:04", r.End)
		if err1 != nil || err2 != nil {
			return appErr.Wrap("Schedule.UpdateWorkDay.ParseTime", appErr.ErrInvalidInput, nil)
		}
		workDay.Ranges = append(workDay.Ranges, models.TimeRange{Start: start, End: end})
	}

	if err := h.service.UpdateWorkDay(workDay); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Horario laboral actualizado correctamente"})
}

// POST /schedule/special-hours
func (h *Handler) AddSpecialDay(c echo.Context) error {
	var req models.CreateSpecialDayRequest
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("Schedule.AddSpecialDay.Bind", appErr.ErrInvalidInput, err)
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return appErr.Wrap("Schedule.AddSpecialDay.ParseDate", appErr.ErrInvalidInput, err)
	}

	active := len(req.Ranges) > 0

	day := models.SpecialDay{Date: date, Active: active}
	for _, r := range req.Ranges {
		start, err1 := time.Parse("15:04", r.Start)
		end, err2 := time.Parse("15:04", r.End)
		if err1 != nil || err2 != nil {
			return appErr.Wrap("Schedule.AddSpecialDay.ParseTime", appErr.ErrInvalidInput, nil)
		}
		day.Ranges = append(day.Ranges, models.TimeRange{Start: start, End: end})
	}

	if err := h.service.AddSpecialDay(day); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "Horario especial agregado correctamente"})
}

// DELETE /schedule/special-hours/:date
func (h *Handler) DeleteSpecialDay(c echo.Context) error {
	dateStr := c.Param("date")
	if dateStr == "" {
		return appErr.Wrap("Schedule.DeleteSpecialDay", appErr.ErrInvalidInput, nil)
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return appErr.Wrap("Schedule.DeleteSpecialDay.ParseDate", appErr.ErrInvalidInput, err)
	}

	if err := h.service.DeleteSpecialDay(date); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Horario especial eliminado correctamente",
	})
}
