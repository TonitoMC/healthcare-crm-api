package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	authModels "github.com/tonitomc/healthcare-crm-api/internal/domain/auth/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// ─────────────────────────────────────────────────────────────
// PermissionProvider Interface (decouples from userDomain)
// ─────────────────────────────────────────────────────────────

type PermissionProvider interface {
	GetRolesAndPermissions(userID int) ([]any, []PermissionLike, error)
}

type PermissionLike interface {
	GetName() string
}

var permissionProvider PermissionProvider

func InjectPermissionProvider(provider PermissionProvider) {
	permissionProvider = provider
}

// ─────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────

func normalizePermission(p string) string {
	p = strings.TrimSpace(strings.ToLower(p))
	return strings.ReplaceAll(p, "_", "-")
}

func hasPermission(perms []string, required string) bool {
	req := normalizePermission(required)
	for _, raw := range perms {
		if normalizePermission(raw) == req {
			return true
		}
	}
	return false
}

// ─────────────────────────────────────────────────────────────
// Middleware
// ─────────────────────────────────────────────────────────────

func RequirePermission(required string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := c.Get("user").(*jwt.Token)
			if !ok || token == nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Token no válido o ausente.",
				})
			}

			claims, ok := token.Claims.(*authModels.Claims)
			if !ok || claims == nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Estructura de token no válida.",
				})
			}

			userID := int(claims.UserID)
			if userID <= 0 {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Token sin ID de usuario válido.",
				})
			}

			if permissionProvider == nil {
				c.Logger().Error("[RequirePermission] No permission provider injected")
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "No se pudo validar permisos — configuración incompleta.",
				})
			}

			_, dbPerms, err := permissionProvider.GetRolesAndPermissions(userID)
			if err != nil {
				c.Logger().Errorf("[RequirePermission] DB lookup failed: %v", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "No se pudieron verificar los permisos del usuario.",
				})
			}

			var perms []string
			for _, p := range dbPerms {
				perms = append(perms, p.GetName())
			}

			if hasPermission(perms, required) || hasPermission(claims.Permissions, required) {
				return next(c)
			}

			c.Logger().Warnf("[RequirePermission] Permission '%s' denied for user %d", required, userID)
			return c.JSON(http.StatusForbidden, echo.Map{
				"error": appErr.ErrForbidden.Error(),
			})
		}
	}
}
