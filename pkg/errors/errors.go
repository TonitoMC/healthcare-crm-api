package errors

import "errors"

// Centralized application-level errors.
// These are returned by repositories and services, then unwrapped by handlers.
// All messages are in Spanish for consistency with frontend and logs.

var (
	// Generic errors
	ErrNotFound      = errors.New("recurso no encontrado")
	ErrAlreadyExists = errors.New("el recurso ya existe")
	ErrInvalidInput  = errors.New("entrada inválida")
	ErrConflict      = errors.New("conflicto de datos")
	ErrInternal      = errors.New("error interno del servidor")

	// Authentication / authorization errors
	ErrUnauthorized       = errors.New("no autorizado")
	ErrForbidden          = errors.New("acceso denegado")
	ErrInvalidToken       = errors.New("token inválido o expirado")
	ErrInvalidCredentials = errors.New("credenciales inválidas")

	// Domain-specific user errors
	ErrUserNotFound        = errors.New("usuario no encontrado")
	ErrUserAlreadyExists   = errors.New("usuario ya registrado")
	ErrIncompleteData      = errors.New("datos incompletos o incorrectos")
	ErrOperationNotAllowed = errors.New("operación no permitida")
)
