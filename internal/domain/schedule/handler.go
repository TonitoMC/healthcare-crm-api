package schedule

import (
	"fmt"
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

	// Validation operations
	scheduleGroup.GET("/validate/date/:date", h.ValidateDateOpen)
	scheduleGroup.POST("/validate/range", h.ValidateTimeRange)

	// Write operations
	scheduleGroup.PUT("/working-hours/:id", h.UpdateWorkDay)
	scheduleGroup.POST("/special-hours", h.AddSpecialDay)
	scheduleGroup.PUT("/special-hours/:id", h.UpdateSpecialDay)
	scheduleGroup.DELETE("/special-hours/:id", h.DeleteSpecialDay)
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

// GET /schedule/validate/date/:date
func (h *Handler) ValidateDateOpen(c echo.Context) error {
	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return appErr.Wrap("Schedule.ValidateDateOpen.Parse", appErr.ErrInvalidInput, err)
	}

	open, err := h.service.IsDateOpen(date)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"date": date.Format("2006-01-02"),
		"open": open,
	})
}

// POST /schedule/validate/range
func (h *Handler) ValidateTimeRange(c echo.Context) error {
	var body struct {
		Date  string `json:"date"`
		Start string `json:"start"`
		End   string `json:"end"`
	}

	if err := c.Bind(&body); err != nil {
		return appErr.Wrap("Schedule.ValidateTimeRange.Bind", appErr.ErrInvalidInput, err)
	}

	date, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		return appErr.Wrap("Schedule.ValidateTimeRange.ParseDate", appErr.ErrInvalidInput, err)
	}

	start, err := time.Parse("15:04", body.Start)
	if err != nil {
		return appErr.Wrap("Schedule.ValidateTimeRange.ParseStart", appErr.ErrInvalidInput, err)
	}

	end, err := time.Parse("15:04", body.End)
	if err != nil {
		return appErr.Wrap("Schedule.ValidateTimeRange.ParseEnd", appErr.ErrInvalidInput, err)
	}

	valid, err := h.service.IsTimeRangeWithinWorkingHours(date, start, end)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"valid": valid,
	})
}

// PUT /schedule/working-hours/:id
func (h *Handler) UpdateWorkDay(c echo.Context) error {
	var req models.CreateWorkDayRequest
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("Schedule.UpdateWorkDay.Bind", appErr.ErrInvalidInput, err)
	}

	workDay := models.WorkDay{
		DayOfWeek: req.DayOfWeek,
		Active:    true,
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

	day := models.SpecialDay{Date: date, Active: true}
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

// PUT /schedule/special-hours/:id
func (h *Handler) UpdateSpecialDay(c echo.Context) error {
	var req models.CreateSpecialDayRequest
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("Schedule.UpdateSpecialDay.Bind", appErr.ErrInvalidInput, err)
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return appErr.Wrap("Schedule.UpdateSpecialDay.ParseDate", appErr.ErrInvalidInput, err)
	}

	day := models.SpecialDay{Date: date, Active: true}
	for _, r := range req.Ranges {
		start, err1 := time.Parse("15:04", r.Start)
		end, err2 := time.Parse("15:04", r.End)
		if err1 != nil || err2 != nil {
			return appErr.Wrap("Schedule.UpdateSpecialDay.ParseTime", appErr.ErrInvalidInput, nil)
		}
		day.Ranges = append(day.Ranges, models.TimeRange{Start: start, End: end})
	}

	if err := h.service.UpdateSpecialDay(day); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Horario especial actualizado correctamente"})
}

// DELETE /schedule/special-hours/:id
func (h *Handler) DeleteSpecialDay(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return appErr.Wrap("Schedule.DeleteSpecialDay", appErr.ErrInvalidInput, nil)
	}

	var dayID int
	_, err := fmt.Sscan(id, &dayID)
	if err != nil {
		return appErr.Wrap("Schedule.DeleteSpecialDay.ParseID", appErr.ErrInvalidInput, err)
	}

	if err := h.service.DeleteSpecialDay(dayID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Horario especial eliminado correctamente"})
}
