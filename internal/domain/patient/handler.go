package patient

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/consultation"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/exam"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/medicalrecord"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/patient/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Handler struct {
	service             Service
	examService         exam.Service
	consultationService consultation.Service
	recordService       medicalrecord.Service
}

func NewHandler(s Service, examService exam.Service, consultationService consultation.Service, recordService medicalrecord.Service) *Handler {
	return &Handler{service: s, examService: examService, consultationService: consultationService, recordService: recordService}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	patients := g.Group("/patients")

	patients.GET("", h.GetAll, middleware.RequirePermission("ver-pacientes"))
	patients.GET("/:id", h.GetByID, middleware.RequirePermission("ver-pacientes"))
	patients.GET("/:id/details", h.GetDetails, middleware.RequirePermission("ver-examenes"))
	patients.POST("", h.Create, middleware.RequirePermission("crear-pacientes"))
	patients.PUT("/:id", h.Update, middleware.RequirePermission("editar-pacientes"))
	patients.DELETE("/:id", h.Delete, middleware.RequirePermission("eliminar-pacientes"))
	patients.GET("/search", h.SearchByName, middleware.RequirePermission("ver-pacientes"))
}

func (h *Handler) GetAll(c echo.Context) error {
	patients, err := h.service.GetAll()
	if err != nil {
		return err
	}

	if len(patients) == 0 {
		return c.JSON(http.StatusOK, []models.Patient{})
	}

	return c.JSON(http.StatusOK, patients)
}

func (h *Handler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("PatientHandler.GetByID.ParseID", appErr.ErrInvalidInput, err)
	}

	patient, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, patient)
}

func (h *Handler) Create(c echo.Context) error {
	var req models.PatientCreateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("PatientHandler.Create.Bind", appErr.ErrInvalidInput, err)
	}

	id, err := h.service.Create(&req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{"id": id, "message": "Paciente creado correctamente"})
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("PatientHandler.Update.ParseID", appErr.ErrInvalidInput, err)
	}

	var req models.PatientUpdateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("PatientHandler.Update.Bind", appErr.ErrInvalidInput, err)
	}

	if err := h.service.Update(id, &req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Paciente actualizado correctamente"})
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("PatientHandler.Delete.ParseID", appErr.ErrInvalidInput, err)
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Paciente eliminado correctamente"})
}

func (h *Handler) SearchByName(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		return appErr.Wrap("PatientHandler.SearchByName", appErr.ErrInvalidInput, nil)
	}

	results, err := h.service.SearchByName(name)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, results)
}

func (h *Handler) GetDetails(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("PatientHandler.GetDetails.ParseID", appErr.ErrInvalidInput, err)
	}

	patient, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	// Parse query param
	include := strings.Split(c.QueryParam("include"), ",")
	includes := make(map[string]bool)
	for _, i := range include {
		includes[strings.TrimSpace(i)] = true
	}

	// Base response
	response := echo.Map{"patient": patient}

	// Add related data conditionally
	if includes["exams"] && h.examService != nil {
		if exams, err := h.examService.GetByPatient(id); err == nil {
			response["exams"] = exams
		}
	}

	if includes["consultations"] && h.consultationService != nil {
		if consultations, err := h.consultationService.GetByPatientWithDetails(id); err == nil {
			response["consultations"] = consultations
		}
	}

	if includes["record"] && h.recordService != nil {
		if record, err := h.recordService.GetByPatientID(id); err == nil {
			response["medical_record"] = record
		}
	}

	return c.JSON(http.StatusOK, response)
}
