package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/auth"
)

func RegisterAuthRoutes(e *echo.Echo, authHandler *auth.AuthHandler) {
	authGroup := e.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
}
