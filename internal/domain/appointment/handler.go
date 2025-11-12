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
	appointments := e.Group("/appointments", ErrorMiddleware())
	appointments.GET("", h.GetBetween, middleware.RequirePermission("ver-citas"))
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
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "ID inválido"})
	}
	appt, err := h.service.GetByID(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, appt)
}

func (h *Handler) GetToday(c echo.Context) error {
	// Localizar al timezone de la clínica para evitar desalineaciones con TIMESTAMPTZ
	clinicLoc, _ := time.LoadLocation("America/Guatemala")
	// time.Now() podría venir en otro TZ según el servidor; normalizamos
	today := time.Now().In(clinicLoc)
	appts, err := h.service.GetByDate(today)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, appts)
}

func (h *Handler) GetByDate(c echo.Context) error {
	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Formato de fecha inválido, use AAAA-MM-DD"})
	}
	clinicLoc, _ := time.LoadLocation("America/Guatemala")
	localized := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, clinicLoc)
	appts, err := h.service.GetByDate(localized)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, appts)
}

func (h *Handler) GetBetween(c echo.Context) error {
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	if startStr == "" || endStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Se requieren los parámetros 'start' y 'end' en formato AAAA-MM-DD"})
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Formato de fecha inicial inválido, use AAAA-MM-DD"})
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Formato de fecha final inválido, use AAAA-MM-DD"})
	}

	clinicLoc, _ := time.LoadLocation("America/Guatemala")
	localizedStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, clinicLoc)
	localizedEnd := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, clinicLoc)

	appts, err := h.service.GetBetween(localizedStart, localizedEnd)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, appts)
}

func (h *Handler) Create(c echo.Context) error {
	var req models.AppointmentCreateDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Cuerpo de solicitud inválido"})
	}
	id, err := h.service.Create(&req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, echo.Map{"id": id})
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "ID inválido"})
	}
	var req models.AppointmentUpdateDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Cuerpo de solicitud inválido"})
	}
	if err := h.service.Update(id, &req); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Cita actualizada exitosamente"})
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "ID inválido"})
	}
	if err := h.service.Delete(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Cita eliminada exitosamente"})
}

func (h *Handler) CreateWithNewPatient(c echo.Context) error {
	var req models.AppointmentWithNewPatientDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Cuerpo de solicitud inválido"})
	}
	id, err := h.service.CreateWithNewPatient(&req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, echo.Map{"id": id})
}

func (h *Handler) GetAvailableSlots(c echo.Context) error {
	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Formato de fecha inválido, use AAAA-MM-DD"})
	}
	clinicLoc, _ := time.LoadLocation("America/Guatemala")
	localized := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, clinicLoc)

	slotDuration := int64(900) // 15 min default
	if dur := c.QueryParam("duration"); dur != "" {
		if parsed, err := strconv.ParseInt(dur, 10, 64); err == nil {
			slotDuration = parsed
		}
	}

	slots, err := h.service.GetAvailableSlots(localized, slotDuration)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, slots)
}
