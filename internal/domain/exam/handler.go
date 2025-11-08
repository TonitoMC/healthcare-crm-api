package exam

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/exam/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(e *echo.Group) {
	exams := e.Group("/exams", ErrorMiddleware()) // attach error middleware

	exams.GET("/:id", h.GetByID, middleware.RequirePermission("ver-examenes"))
	exams.GET("/pending", h.GetPending, middleware.RequirePermission("ver-examenes"))

	exams.GET("/patient/:patientId", h.GetByPatientID, middleware.RequirePermission("ver-examenes"))
	exams.POST("", h.Create, middleware.RequirePermission("manejar-examenes"))
	exams.PATCH("/:id", h.Update, middleware.RequirePermission("manejar-examenes"))
	exams.DELETE("/:id", h.Delete, middleware.RequirePermission("manejar-examenes"))
}

// ============================================================================
// HANDLERS
// ============================================================================

func (h *Handler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ExamHandler.GetByID", appErr.ErrInvalidInput, err)
	}

	exam, err := h.service.GetByID(id)
	if err != nil {
		return err // bubble up to middleware
	}

	return c.JSON(http.StatusOK, exam)
}

func (h *Handler) Create(c echo.Context) error {
	var req models.ExamCreateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("ExamHandler.Create", appErr.ErrInvalidRequest, err)
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
		return appErr.Wrap("ExamHandler.Update", appErr.ErrInvalidInput, err)
	}

	var dto models.ExamDTO
	if err := c.Bind(&dto); err != nil {
		return appErr.Wrap("ExamHandler.Update", appErr.ErrInvalidRequest, err)
	}

	if err := h.service.Update(id, &dto); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Examen actualizado correctamente"})
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("ExamHandler.Delete", appErr.ErrInvalidInput, err)
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Examen eliminado correctamente"})
}

func (h *Handler) GetPending(c echo.Context) error {
	exams, err := h.service.GetPending()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, exams)
}

func (h *Handler) GetByPatientID(c echo.Context) error {
	patientID, err := strconv.Atoi(c.Param("patientId"))
	if err != nil {
		return appErr.Wrap("ExamHandler.GetByPatientID", appErr.ErrInvalidInput, err)
	}

	exams, err := h.service.GetByPatient(patientID)
	if err != nil {
		return err
	}

	if len(exams) == 0 {
		return c.JSON(http.StatusOK, []models.ExamDTO{})
	}

	return c.JSON(http.StatusOK, exams)
}
