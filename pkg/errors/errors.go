package errors

import (
	"errors"
	"fmt"
)

// Core sentinel errors (global, domain-agnostic)
// These errors are used across repositories and services, domain handlers
// & middleware map them into user-facing messages.

// Errors in this api usually have 3 "levels"
// These sentinel ones, which are the least descriptive
// Sentinels ones translated by Domain, eg. resource not found -> user not found
// Custom domain errors (seen below) where the error type is kept & the message completely replaced
// eg. "Appointment is outside business hours"
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

// DomainError represents a domain-specific error that carries both a sentinel code
// and a contextual, human-readable message (useful for API responses).
type DomainError struct {
	Code    error  // the sentinel (e.g., ErrConflict, ErrNotFound)
	Message string // domain-specific or user-friendly message
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	if e.Message == "" {
		return e.Code.Error()
	}
	return fmt.Sprintf("%s: %s", e.Code.Error(), e.Message)
}

// NewDomainError is a helper to construct a DomainError cleanly.
func NewDomainError(code error, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

// IsDomainError checks whether an error is a domain-level business-rule violation.
// This allows middleware and services to distinguish user-facing domain errors
// from lower-level or unexpected failures.
func IsDomainError(err error) bool {
	var d *DomainError
	return errors.As(err, &d)
}
