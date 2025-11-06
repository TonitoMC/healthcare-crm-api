package medicalrecord

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/medicalrecord/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	medical := g.Group("/medical-records")

	medical.GET("/:patientId", h.GetByPatientID, middleware.RequirePermission("ver-pacientes"))
	medical.PUT("/:patientId", h.Update, middleware.RequirePermission("editar-pacientes"))
}

func (h *Handler) GetByPatientID(c echo.Context) error {
	patientID, err := strconv.Atoi(c.Param("patientId"))
	if err != nil {
		return appErr.Wrap("MedicalRecordHandler.GetByPatientID.ParseID", appErr.ErrInvalidInput, err)
	}

	// Ensure record exists
	if err := h.service.EnsureExists(patientID); err != nil {
		return err
	}

	record, err := h.service.GetByPatientID(patientID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, record)
}

func (h *Handler) Update(c echo.Context) error {
	patientID, err := strconv.Atoi(c.Param("patientId"))
	if err != nil {
		return appErr.Wrap("MedicalRecordHandler.Update.ParseID", appErr.ErrInvalidInput, err)
	}

	var req models.MedicalRecordUpdateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("MedicalRecordHandler.Update.Bind", appErr.ErrInvalidInput, err)
	}

	if err := h.service.Update(patientID, &req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Antecedentes actualizados correctamente"})
}
