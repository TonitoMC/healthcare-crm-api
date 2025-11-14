package reminder

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	models "github.com/tonitomc/healthcare-crm-api/internal/domain/reminder/models"
	"github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	r := g.Group("/reminders", ErrorMiddleware(), middleware.RequireAuth())

	r.GET("", h.GetMyReminders)
	r.POST("", h.CreateReminder)
	r.PUT("/:id/done", h.MarkDone)
	r.PUT("/:id/undone", h.MarkUndone)
	r.DELETE("/:id", h.DeleteReminder)
}

// ----------------------------------------------------------------------

func (h *Handler) GetMyReminders(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return errors.Wrap("Reminder.GetMyReminders", errors.ErrUnauthorized, nil)
	}

	data, err := h.service.GetForUser(claims.UserID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, data)
}

func (h *Handler) CreateReminder(c echo.Context) error {
	var req models.CreateReminderRequest
	if err := c.Bind(&req); err != nil {
		return errors.Wrap("Reminder.Create.Bind", errors.ErrInvalidInput, err)
	}

	claims := middleware.GetClaims(c)
	if claims == nil {
		return errors.Wrap("Reminder.Create.GetClaims", errors.ErrUnauthorized, nil)
	}

	rem, err := h.service.Create(claims.UserID, req.Description, req.Global)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, rem)
}

func (h *Handler) MarkDone(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.SetDone(id); err != nil {
		return err // middleware will handle
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":          true,
		"fecha_completado": time.Now(),
	})
}

func (h *Handler) MarkUndone(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.SetUndone(id); err != nil {
		return err // middleware will handle
	}

	return c.JSON(http.StatusOK, map[string]bool{
		"success": true,
	})
}

func (h *Handler) DeleteReminder(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.Delete(id); err != nil {
		return err // middleware will handle
	}

	return c.JSON(http.StatusOK, map[string]bool{
		"success": true,
	})
}
