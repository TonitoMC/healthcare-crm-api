package routes

import (
	"github.com/labstack/echo/v4"

	authDomain "github.com/tonitomc/healthcare-crm-api/internal/domain/auth"
)

// RegisterRoutes wires all domain route groups under /api/v1.
func RegisterRoutes(e *echo.Echo, authHandler *authDomain.Handler) {
	api := e.Group("/api")

	authHandler.RegisterRoutes(api)
}
