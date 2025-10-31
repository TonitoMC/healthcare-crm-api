// internal/domain/auth/middleware.go
package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// ErrorMiddleware returns an echo.MiddlewareFunc scoped to /auth routes.
func ErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			status, msg := mapError(err)
			c.Logger().Errorf("[Auth] %v", err)
			return c.JSON(status, echo.Map{"error": msg})
		}
	}
}

// local mapError helper (kept close)
func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, appErr.ErrInvalidCredentials):
		return http.StatusUnauthorized, "Credenciales inválidas."
	case errors.Is(err, appErr.ErrNotFound):
		return http.StatusNotFound, "Usuario no encontrado."
	case errors.Is(err, appErr.ErrAlreadyExists):
		return http.StatusConflict, "Usuario ya existente."
	case errors.Is(err, appErr.ErrInvalidInput):
		return http.StatusBadRequest, "Datos incompletos o incorrectos."
	default:
		return http.StatusInternalServerError, appErr.ErrInternal.Error()
	}
}
