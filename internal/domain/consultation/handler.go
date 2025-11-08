package consultation

import (
	"net/http"
	"strconv"
	"strings"

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

// ===================== ROUTES =====================

func (h *Handler) RegisterRoutes(g *echo.Group) {
	consultations := g.Group("/consultations")

	// --- Consultations ---
	consultations.GET("", h.GetAll, middleware.RequirePermission("ver-consultas"))
	consultations.GET("/:id", h.GetByID, middleware.RequirePermission("ver-consultas"))
	consultations.GET("/patient/:patientId", h.GetByPatient, middleware.RequirePermission("ver-consultas"))
	consultations.GET("/:id/details", h.GetDetails, middleware.RequirePermission("ver-consultas"))
	consultations.POST("", h.Create, middleware.RequirePermission("manejar-consultas"))
	consultations.PUT("/:id", h.Update, middleware.RequirePermission("manejar-consultas"))
	consultations.DELETE("/:id", h.Delete, middleware.RequirePermission("manejar-consultas"))

	// --- Diagnostics ---
	consultations.GET("/:id/diagnostics", h.GetDiagnosticsByConsultation, middleware.RequirePermission("ver-consultas"))
	consultations.GET("/:id/diagnostics/:diagId", h.GetDiagnosticByID, middleware.RequirePermission("ver-consultas"))
	consultations.POST("/:id/diagnostics", h.CreateDiagnostic, middleware.RequirePermission("manejar-consultas"))
	consultations.PUT("/:id/diagnostics/:diagId", h.UpdateDiagnostic, middleware.RequirePermission("manejar-consultas"))
	consultations.DELETE("/:id/diagnostics/:diagId", h.DeleteDiagnostic, middleware.RequirePermission("manejar-consultas"))

	// --- Treatments ---
	consultations.GET("/:id/diagnostics/:diagId/treatments", h.GetTreatmentsByDiagnostic, middleware.RequirePermission("ver-consultas"))
	consultations.GET("/:id/diagnostics/:diagId/treatments/:treatmentId", h.GetTreatmentByID, middleware.RequirePermission("ver-consultas"))
	consultations.POST("/:id/diagnostics/:diagId/treatments", h.CreateTreatment, middleware.RequirePermission("manejar-consultas"))
	consultations.PUT("/:id/diagnostics/:diagId/treatments/:treatmentId", h.UpdateTreatment, middleware.RequirePermission("manejar-consultas"))
	consultations.DELETE("/:id/diagnostics/:diagId/treatments/:treatmentId", h.DeleteTreatment, middleware.RequirePermission("manejar-consultas"))

	// Answers

	// --- Answers ---
	consultations.GET("/:id/answers", h.GetAnswersByConsultation, middleware.RequirePermission("ver-consultas"))
	consultations.POST("/:id/answers", h.AddAnswers, middleware.RequirePermission("manejar-consultas"))
	consultations.PUT("/:id/answers", h.UpdateAnswers, middleware.RequirePermission("manejar-consultas"))
	consultations.DELETE("/:id/answers", h.DeleteAnswers, middleware.RequirePermission("manejar-consultas"))
}

// ===================== CONSULTATIONS =====================

func (h *Handler) GetAll(c echo.Context) error {
	consultations, err := h.service.GetAll()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, consultations)
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

// --- Details Aggregation ---

func parseIncludes(q string) (withDiagnostics, withTreatments, withAnswers bool) {
	for _, part := range strings.Split(q, ",") {
		switch strings.TrimSpace(strings.ToLower(part)) {
		case "diagnostics":
			withDiagnostics = true
		case "treatments":
			withTreatments = true
		case "answers":
			withAnswers = true
		}
	}
	return
}

func (h *Handler) GetDetails(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.GetDetails.ParseID", appErr.ErrInvalidInput, err)
	}

	withDiagnostics, withTreatments, withAnswers := parseIncludes(c.QueryParam("include"))

	consultation, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	resp := echo.Map{"consultation": consultation}

	if withDiagnostics {
		diagnostics, err := h.service.GetDiagnosticsByConsultation(id)
		if err != nil {
			return err
		}

		if withTreatments {
			type diagWithTreat struct {
				models.Diagnostic `json:"diagnostic"`
				Treatments        []models.Treatment `json:"treatments"`
			}
			var items []diagWithTreat
			for _, d := range diagnostics {
				trs, err := h.service.GetTreatmentsByDiagnostic(d.ID)
				if err != nil {
					return err
				}
				items = append(items, diagWithTreat{Diagnostic: d, Treatments: trs})
			}
			resp["diagnostics"] = items
		} else {
			resp["diagnostics"] = diagnostics
		}
	}

	if withAnswers {
		// TODO: When answers service exists
		resp["answers"] = []string{}
	}

	return c.JSON(http.StatusOK, resp)
}

// ===================== DIAGNOSTICS =====================

func (h *Handler) GetDiagnosticsByConsultation(c echo.Context) error {
	consultationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.GetDiagnosticsByConsultation.ParseID", appErr.ErrInvalidInput, err)
	}
	list, err := h.service.GetDiagnosticsByConsultation(consultationID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, list)
}

func (h *Handler) GetDiagnosticByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("diagId"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.GetDiagnosticByID.ParseID", appErr.ErrInvalidInput, err)
	}
	d, err := h.service.GetDiagnosticByID(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) CreateDiagnostic(c echo.Context) error {
	consultationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.CreateDiagnostic.ParseID", appErr.ErrInvalidInput, err)
	}
	var req models.DiagnosticCreateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("ConsultationHandler.CreateDiagnostic.Bind", appErr.ErrInvalidInput, err)
	}
	req.ConsultaID = consultationID
	id, err := h.service.CreateDiagnostic(&req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, echo.Map{"id": id, "message": "Diagnóstico creado correctamente"})
}

func (h *Handler) UpdateDiagnostic(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("diagId"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.UpdateDiagnostic.ParseID", appErr.ErrInvalidInput, err)
	}
	var req models.DiagnosticUpdateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("ConsultationHandler.UpdateDiagnostic.Bind", appErr.ErrInvalidInput, err)
	}
	if err := h.service.UpdateDiagnostic(id, &req); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Diagnóstico actualizado correctamente"})
}

func (h *Handler) DeleteDiagnostic(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("diagId"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.DeleteDiagnostic.ParseID", appErr.ErrInvalidInput, err)
	}
	if err := h.service.DeleteDiagnostic(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Diagnóstico eliminado correctamente"})
}

// ===================== TREATMENTS =====================

func (h *Handler) GetTreatmentsByDiagnostic(c echo.Context) error {
	diagID, err := strconv.Atoi(c.Param("diagId"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.GetTreatmentsByDiagnostic.ParseID", appErr.ErrInvalidInput, err)
	}
	list, err := h.service.GetTreatmentsByDiagnostic(diagID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, list)
}

func (h *Handler) GetTreatmentByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("treatmentId"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.GetTreatmentByID.ParseID", appErr.ErrInvalidInput, err)
	}
	t, err := h.service.GetTreatmentByID(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, t)
}

func (h *Handler) CreateTreatment(c echo.Context) error {
	diagID, err := strconv.Atoi(c.Param("diagId"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.CreateTreatment.ParseID", appErr.ErrInvalidInput, err)
	}
	var req models.TreatmentCreateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("ConsultationHandler.CreateTreatment.Bind", appErr.ErrInvalidInput, err)
	}
	req.DiagnosticoID = diagID
	id, err := h.service.CreateTreatment(&req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, echo.Map{"id": id, "message": "Tratamiento creado correctamente"})
}

func (h *Handler) UpdateTreatment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("treatmentId"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.UpdateTreatment.ParseID", appErr.ErrInvalidInput, err)
	}
	var req models.TreatmentUpdateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("ConsultationHandler.UpdateTreatment.Bind", appErr.ErrInvalidInput, err)
	}
	if err := h.service.UpdateTreatment(id, &req); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Tratamiento actualizado correctamente"})
}

func (h *Handler) DeleteTreatment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("treatmentId"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.DeleteTreatment.ParseID", appErr.ErrInvalidInput, err)
	}
	if err := h.service.DeleteTreatment(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Tratamiento eliminado correctamente"})
}

func (h *Handler) GetAnswersByConsultation(c echo.Context) error {
	consultationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.GetAnswersByConsultation.ParseID", appErr.ErrInvalidInput, err)
	}

	answers, err := h.service.GetAnswersByConsultation(consultationID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, answers)
}

func (h *Handler) AddAnswers(c echo.Context) error {
	consultationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.AddAnswers.ParseID", appErr.ErrInvalidInput, err)
	}

	var req models.AnswersCreateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("ConsultationHandler.AddAnswers.Bind", appErr.ErrInvalidInput, err)
	}

	id, err := h.service.AddAnswers(consultationID, &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"id":      id,
		"message": "Respuestas registradas correctamente",
	})
}

func (h *Handler) UpdateAnswers(c echo.Context) error {
	consultationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.UpdateAnswers.ParseID", appErr.ErrInvalidInput, err)
	}

	var req models.AnswersUpdateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("ConsultationHandler.UpdateAnswers.Bind", appErr.ErrInvalidInput, err)
	}

	if err := h.service.UpdateAnswers(consultationID, &req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Respuestas actualizadas correctamente",
	})
}

func (h *Handler) DeleteAnswers(c echo.Context) error {
	consultationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ConsultationHandler.DeleteAnswers.ParseID", appErr.ErrInvalidInput, err)
	}

	if err := h.service.DeleteAnswers(consultationID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Respuestas eliminadas correctamente",
	})
}
