package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	authModels "github.com/tonitomc/healthcare-crm-api/internal/domain/auth/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Handler exposes HTTP endpoints for authentication.
type Handler struct {
	service Service
}

// NewHandler constructs a new AuthHandler.
func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes mounts /auth routes under the provided Echo group.
// The route group will have error-handling middleware attached externally (via routes.go).
func (h *Handler) RegisterRoutes(g *echo.Group) {
	authGroup := g.Group("/auth", ErrorMiddleware())
	authGroup.POST("/register", h.Register)
	authGroup.POST("/login", h.Login)
}

// -----------------------------------------------------------------------------
// POST /auth/register
// -----------------------------------------------------------------------------
func (h *Handler) Register(c echo.Context) error {
	var req authModels.RegisterRequest

	// Bind JSON input safely
	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("Auth.Register.Bind", appErr.ErrInvalidRequest, err)
	}

	// Delegate to service
	if err := h.service.Register(req.Username, req.Email, req.Password); err != nil {
		return err // handled by global middleware
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "Usuario registrado correctamente",
	})
}

// -----------------------------------------------------------------------------
// POST /auth/login
// -----------------------------------------------------------------------------
func (h *Handler) Login(c echo.Context) error {
	var req authModels.LoginRequest

	if err := c.Bind(&req); err != nil {
		return appErr.Wrap("Auth.Login.Bind", appErr.ErrInvalidRequest, err)
	}

	token, err := h.service.Login(req.Identifier, req.Password)
	if err != nil {
		return err // handled by middleware
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
	})
}
