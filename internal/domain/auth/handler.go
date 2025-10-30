package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// AuthHandler holds a reference to the AuthService
type AuthHandler struct {
	service *Service
}

// NewAuthHandler creates a new handler instance
func NewAuthHandler(service *Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// RegisterRequest represents the expected JSON body for /register
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the expected JSON body for /login
type LoginRequest struct {
	Identifier string `json:"identifier"` // username or email
	Password   string `json:"password"`
}

// POST /register
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	if err := h.service.Register(req.Username, req.Email, req.Password); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "User registered successfully"})
}

// POST /login
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	token, err := h.service.Login(req.Identifier, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}
