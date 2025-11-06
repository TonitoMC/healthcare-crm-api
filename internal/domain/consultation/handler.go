package consultation

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/consultation/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	consultations := g.Group("/consultations")

	consultations.GET("/:id", h.GetByID, middleware.RequirePermission("ver-pacientes"))
	consultations.GET("/patient/:patientId", h.GetByPatient, middleware.RequirePermission("ver-pacientes"))
	consultations.POST("", h.Create, middleware.RequirePermission("crear-pacientes"))
	consultations.PUT("/:id", h.Update, middleware.RequirePermission("editar-pacientes"))
	consultations.DELETE("/:id", h.Delete, middleware.RequirePermission("eliminar-pacientes"))
}

func (h *Handler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.GetByID.ParseID", appErr.ErrInvalidInput, err)
	}

	consultation, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, consultation)
}

func (h *Handler) GetByPatient(c echo.Context) error {
	patientID, err := strconv.Atoi(c.Param("patientId"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.GetByPatient.ParseID", appErr.ErrInvalidInput, err)
	}

	consultations, err := h.service.GetByPatient(patientID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, consultations)
}

func (h *Handler) Create(c echo.Context) error {
	var req models.ConsultationCreateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("ConsultationHandler.Create.Bind", appErr.ErrInvalidInput, err)
	}

	id, err := h.service.Create(&req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{"id": id, "message": "Consulta creada correctamente"})
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.Update.ParseID", appErr.ErrInvalidInput, err)
	}

	var req models.ConsultationUpdateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("ConsultationHandler.Update.Bind", appErr.ErrInvalidInput, err)
	}

	if err := h.service.Update(id, &req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Consulta actualizada correctamente"})
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.Delete.ParseID", appErr.ErrInvalidInput, err)
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Consulta eliminada correctamente"})
}
