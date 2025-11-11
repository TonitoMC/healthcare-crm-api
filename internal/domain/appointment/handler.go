package appointment

import (
"errors"
"net/http"
"strconv"
"time"

"github.com/labstack/echo/v4"
"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
"github.com/tonitomc/healthcare-crm-api/internal/domain/appointment/models"
appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Handler struct {
service Service
}

func NewHandler(service Service) *Handler {
return &Handler{service: service}
}

// handleError maps service errors to appropriate HTTP status codes and messages
func (h *Handler) handleError(c echo.Context, err error) error {
// Check for domain errors first
var domainErr *appErr.DomainError
if errors.As(err, &domainErr) {
switch {
case errors.Is(domainErr.Code, appErr.ErrConflict):
return c.JSON(http.StatusConflict, echo.Map{"error": domainErr.Message})
case errors.Is(domainErr.Code, appErr.ErrNotFound):
return c.JSON(http.StatusNotFound, echo.Map{"error": domainErr.Message})
case errors.Is(domainErr.Code, appErr.ErrInvalidInput):
return c.JSON(http.StatusBadRequest, echo.Map{"error": domainErr.Message})
}
}

// Check for sentinel errors
switch {
case errors.Is(err, appErr.ErrConflict):
return c.JSON(http.StatusConflict, echo.Map{"error": "Ya existe una cita en ese horario. Recuerda que debe haber 5 minutos entre citas."})
case errors.Is(err, appErr.ErrNotFound):
return c.JSON(http.StatusNotFound, echo.Map{"error": "Cita no encontrada"})
case errors.Is(err, appErr.ErrInvalidInput):
return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
case errors.Is(err, appErr.ErrInvalidRequest):
return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
default:
return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error interno del servidor"})
}
}

func (h *Handler) RegisterRoutes(e *echo.Group) {
appointments := e.Group("/appointments")
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
return h.handleError(c, err)
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
return h.handleError(c, err)
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
return h.handleError(c, err)
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
return h.handleError(c, err)
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
return h.handleError(c, err)
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
return h.handleError(c, err)
}
return c.JSON(http.StatusOK, echo.Map{"message": "Cita actualizada exitosamente"})
}

func (h *Handler) Delete(c echo.Context) error {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
return c.JSON(http.StatusBadRequest, echo.Map{"error": "ID inválido"})
}
if err := h.service.Delete(id); err != nil {
return h.handleError(c, err)
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
return h.handleError(c, err)
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
return h.handleError(c, err)
}
return c.JSON(http.StatusOK, slots)
}
