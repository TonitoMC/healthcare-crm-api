package appointment

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/appointment/models"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(e *echo.Group) {
	appointments := e.Group("/appointments")
	appointments.GET("/:id", h.GetByID, middleware.RequirePermission("ver-citas"))
	appointments.GET("/today", h.GetToday, middleware.RequirePermission("ver-citas"))
	appointments.GET("/date/:date", h.GetByDate, middleware.RequirePermission("ver-citas"))
	appointments.GET("/available-slots/:date", h.GetAvailableSlots, middleware.RequirePermission("ver-citas"))
	appointments.POST("", h.Create, middleware.RequirePermission("crear-citas"))
	appointments.POST("/with-new-patient", h.CreateWithNewPatient, middleware.RequirePermission("crear-citas"))
	appointments.PUT("/:id", h.Update, middleware.RequirePermission("editar-citas"))
	appointments.DELETE("/:id", h.Delete, middleware.RequirePermission("eliminar-citas"))
}

func (h *Handler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}
	appt, err := h.service.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, appt)
}

func (h *Handler) GetToday(c echo.Context) error {
	appts, err := h.service.GetToday()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, appts)
}

func (h *Handler) GetByDate(c echo.Context) error {
	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid date format, use YYYY-MM-DD"})
	}
	appts, err := h.service.GetByDate(date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, appts)
}

func (h *Handler) Create(c echo.Context) error {
	var req models.AppointmentCreateDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	id, err := h.service.Create(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, echo.Map{"id": id})
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}
	var req models.AppointmentUpdateDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	if err := h.service.Update(id, &req); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Appointment updated successfully"})
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}
	if err := h.service.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Appointment deleted successfully"})
}

func (h *Handler) CreateWithNewPatient(c echo.Context) error {
	var req models.AppointmentWithNewPatientDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}
	id, err := h.service.CreateWithNewPatient(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, echo.Map{"id": id})
}

func (h *Handler) GetAvailableSlots(c echo.Context) error {
	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid date format, use YYYY-MM-DD"})
	}

	slotDuration := int64(900) // 15 min default
	if dur := c.QueryParam("duration"); dur != "" {
		if parsed, err := strconv.ParseInt(dur, 10, 64); err == nil {
			slotDuration = parsed
		}
	}

	slots, err := h.service.GetAvailableSlots(date, slotDuration)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, slots)
}
