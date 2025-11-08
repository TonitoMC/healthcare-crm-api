package questionnaire

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/api/middleware"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/questionnaire/models"
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
	q := g.Group("/questionnaires")

	q.GET("", h.GetAll, middleware.RequirePermission("ver-cuestionarios"))
	q.GET("/:id", h.GetByID, middleware.RequirePermission("ver-cuestionarios"))
	q.GET("/names", h.GetNames, middleware.RequirePermission("ver-cuestionarios"))
	q.GET("/active/:name", h.GetActiveByName, middleware.RequirePermission("ver-cuestionarios"))

	q.POST("", h.Create, middleware.RequirePermission("manejar-cuestionarios"))
	q.PUT("/:id", h.Update, middleware.RequirePermission("manejar-cuestionarios"))
	q.DELETE("/:id", h.Delete, middleware.RequirePermission("manejar-cuestionarios"))

	q.PUT("/:id/activate", h.SetActive, middleware.RequirePermission("manejar-cuestionarios"))
	q.PUT("/:id/deactivate", h.SetInactive, middleware.RequirePermission("manejar-cuestionarios"))

	// optional: validate answers externally (for testing)
	q.POST("/:id/validate", h.ValidateAnswers, middleware.RequirePermission("ver-cuestionarios"))
}

// ===================== HANDLERS =====================

func (h *Handler) GetAll(c echo.Context) error {
	list, err := h.service.GetAll()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, list)
}

func (h *Handler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("QuestionnaireHandler.GetByID.ParseID", appErr.ErrInvalidInput, err)
	}
	q, err := h.service.GetByID(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, q)
}

func (h *Handler) GetNames(c echo.Context) error {
	names, err := h.service.GetQuestionnaireNames()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, names)
}

func (h *Handler) GetActiveByName(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "Debe especificar el nombre del cuestionario.")
	}
	q, err := h.service.GetActiveByName(name)
	if err != nil {
		return err
	}
	if q == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "No se encontró un cuestionario activo con ese nombre"})
	}
	return c.JSON(http.StatusOK, q)
}

func (h *Handler) Create(c echo.Context) error {
	var req models.QuestionnaireCreateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("QuestionnaireHandler.Create.Bind", appErr.ErrInvalidInput, err)
	}

	id, err := h.service.Create(&req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, echo.Map{"id": id, "message": "Cuestionario creado correctamente"})
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("QuestionnaireHandler.Update.ParseID", appErr.ErrInvalidInput, err)
	}
	var req models.QuestionnaireUpdateDTO
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("QuestionnaireHandler.Update.Bind", appErr.ErrInvalidInput, err)
	}
	if err := h.service.Update(id, &req); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Cuestionario actualizado correctamente"})
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("QuestionnaireHandler.Delete.ParseID", appErr.ErrInvalidInput, err)
	}
	if err := h.service.Delete(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Cuestionario eliminado correctamente"})
}

func (h *Handler) SetActive(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("QuestionnaireHandler.SetActive.ParseID", appErr.ErrInvalidInput, err)
	}
	if err := h.service.SetActive(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Cuestionario activado correctamente"})
}

func (h *Handler) SetInactive(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("QuestionnaireHandler.SetInactive.ParseID", appErr.ErrInvalidInput, err)
	}
	if err := h.service.SetInactive(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Cuestionario desactivado correctamente"})
}

// --- Validation Endpoint (Optional) ---
// This lets you test the questionnaire.Validate() logic directly via API.
func (h *Handler) ValidateAnswers(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return appErr.Wrap("QuestionnaireHandler.ValidateAnswers.ParseID", appErr.ErrInvalidInput, err)
	}

	var body map[string]any
	if err := c.Bind(&body); err != nil {
		return appErr.Wrap("QuestionnaireHandler.ValidateAnswers.Bind", appErr.ErrInvalidInput, err)
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return appErr.Wrap("QuestionnaireHandler.ValidateAnswers.Marshal", appErr.ErrInternal, err)
	}

	if err := h.service.Validate(id, raw); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Respuestas válidas"})
}
