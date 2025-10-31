package errors

import (
	"errors"
	"fmt"
)

// Core sentinel errors (global, domain-agnostic)
// These errors are used across repositories and services, domain handlers
// & middleware map them into user-facing messages.
var (
	// Generic application errors
	ErrInvalidRequest = errors.New("solicitud inválida")
	ErrInvalidInput   = errors.New("entrada inválida")
	ErrIncompleteData = errors.New("datos incompletos o incorrectos")
	ErrNotFound       = errors.New("recurso no encontrado")
	ErrAlreadyExists  = errors.New("el recurso ya existe")
	ErrConflict       = errors.New("conflicto de datos")
	ErrInternal       = errors.New("error interno del servidor")

	// Authentication / authorization errors
	ErrUnauthorized       = errors.New("no autorizado")
	ErrForbidden          = errors.New("acceso denegado")
	ErrInvalidToken       = errors.New("token inválido o expirado")
	ErrInvalidCredentials = errors.New("credenciales inválidas")

	// Operational / rule violations
	ErrOperationNotAllowed = errors.New("operación no permitida")
)

// Wrap adds context, a human-readable sentinel, and an optional verbose internal error.
// Order: human-readable first → technical detail last.
func Wrap(context string, public, internal error) error {
	if internal == nil {
		return fmt.Errorf("%s: %w", context, public)
	}
	return fmt.Errorf("%s: %w: %v", context, public, internal)
}
