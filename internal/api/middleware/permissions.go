package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	authModels "github.com/tonitomc/healthcare-crm-api/internal/domain/auth/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// normalizePermission converts permission strings to a canonical comparable form
// e.g. "gestion_citas" => "gestion-citas"
func normalizePermission(p string) string {
	p = strings.TrimSpace(p)
	p = strings.ToLower(p)
	p = strings.ReplaceAll(p, "_", "-")
	return p
}

// hasPermission checks if the token permissions satisfy the required permission,
// supporting legacy aliases (e.g., gestion-citas, manejar-citas imply all citas actions).
func hasPermission(perms []string, required string) bool {
	req := normalizePermission(required)
	for _, raw := range perms {
		p := normalizePermission(raw)
		if p == req {
			return true
		}
		// Legacy/group aliases for appointments domain
		if strings.HasSuffix(req, "-citas") {
			if p == "gestion-citas" || p == "manejar-citas" {
				return true
			}
		}
	}
	return false
}

func RequirePermission(required string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := c.Get("user").(*jwt.Token)
			if !ok || token == nil {
				c.Logger().Errorf("[RequirePermission] Missing or invalid JWT token")
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Token no válido o ausente.",
				})
			}

			claims, ok := token.Claims.(*authModels.Claims)
			if !ok || claims == nil {
				c.Logger().Errorf("[RequirePermission] Invalid token claims type: %T", token.Claims)
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Estructura de token no válida.",
				})
			}

			if hasPermission(claims.Permissions, required) {
				return next(c)
			}

			c.Logger().Errorf("[RequirePermission] Permission '%s' denied. Token claims: %+v", required, claims)
			return c.JSON(http.StatusForbidden, echo.Map{
				"error": appErr.ErrForbidden.Error(),
			})
		}
	}
}
