package medicalrecord

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	recordModels "github.com/tonitomc/healthcare-crm-api/internal/domain/medicalrecord/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// ============================================================================
//
//	ROUTES
//
// ============================================================================
func (h *Handler) RegisterRoutes(g *echo.Group) {
	mr := g.Group("/medical-records", ErrorMiddleware())

	// You can change permission name to whatever you decide later.
	mr.GET("/:patient_id",
		h.GetByPatientID,
		middleware.RequirePermission("ver-pacientes"))

	mr.PUT("/:patient_id",
		h.Update,
		middleware.RequirePermission("ver-pacientes"))
}

// ============================================================================
//
//	GET /medical-records/:patient_id
//	Retorna el expediente médico completo
//
// ============================================================================
func (h *Handler) GetByPatientID(c echo.Context) error {
	patientID, err := strconv.Atoi(c.Param("patient_id"))
	if err != nil || patientID <= 0 {
		return appErr.Wrap("MedicalRecordHandler.GetByPatientID.ParseID",
			appErr.ErrInvalidInput, err)
	}

	record, svcErr := h.service.GetByPatientID(patientID)
	if svcErr != nil {
		return svcErr // service already returns domain errors
	}

	if record == nil {
		return c.JSON(http.StatusOK, echo.Map{
			"message": "No hay antecedentes registrados",
			"data":    recordModels.MedicalRecord{},
		})
	}

	return c.JSON(http.StatusOK, record)
}

// ============================================================================
//
//	PUT /medical-records/:patient_id
//	Actualiza todos los campos (todos requeridos)
//
// ============================================================================
func (h *Handler) Update(c echo.Context) error {
	patientID, err := strconv.Atoi(c.Param("patient_id"))
	if err != nil || patientID <= 0 {
		return appErr.Wrap("MedicalRecordHandler.Update.ParseID",
			appErr.ErrInvalidInput, err)
	}

	var dto recordModels.MedicalRecordUpdateDTO
	if err := c.Bind(&dto); err != nil {
		return appErr.Wrap("MedicalRecordHandler.Update.Bind",
			appErr.ErrInvalidInput, err)
	}

	// Validate ALL fields must be present
	if dto.Medicos == nil ||
		dto.Familiares == nil ||
		dto.Oculares == nil ||
		dto.Alergicos == nil ||
		dto.Otros == nil {
		return appErr.Wrap("Error", appErr.ErrInvalidInput, err)
	}

	if svcErr := h.service.Update(patientID, &dto); svcErr != nil {
		return svcErr
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Expediente médico actualizado correctamente",
	})
}
