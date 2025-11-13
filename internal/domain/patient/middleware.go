package patient

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// ErrorMiddleware returns an echo.MiddlewareFunc scoped to /exam routes.
func ErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			status, msg := mapError(err)
			c.Logger().Errorf("[Patient] %v", err)
			return c.JSON(status, echo.Map{"error": msg})
		}
	}
}

// mapError maps internal errors to user-facing HTTP responses.
func mapError(err error) (int, string) {
	switch {
	case appErr.IsDomainError(err):
		return http.StatusConflict, err.Error()

	case errors.Is(err, appErr.ErrInvalidInput):
		return http.StatusBadRequest, "Datos inválidos o incompletos."

	case errors.Is(err, appErr.ErrNotFound):
		return http.StatusNotFound, "Paciente no encontrado."

	case errors.Is(err, appErr.ErrConflict):
		return http.StatusConflict, "Conflicto de datos."

	default:
		return http.StatusInternalServerError, appErr.ErrInternal.Error()
	}
}
