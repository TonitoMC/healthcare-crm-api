package routes

import "github.com/labstack/echo/v4"

// RouteRegistrar defines the contract for domain handlers
// that know how to register their own routes under /api.
type RouteRegistrar interface {
	RegisterRoutes(g *echo.Group)
}

// RegisterRoutes mounts all domain route groups under /api.
func RegisterRoutes(e *echo.Echo, registrars ...RouteRegistrar) {
	api := e.Group("/api")

	for _, r := range registrars {
		r.RegisterRoutes(api)
	}
}
