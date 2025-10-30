package database

import (
	"database/sql"
	"errors"

	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type pqError interface {
	SQLState() string
}

// PostgreSQL SQLSTATE code constants.
const (
	CodeUniqueViolation     = "23505"
	CodeForeignKeyViolation = "23503"
	CodeNotNullViolation    = "23502"
	CodeCheckViolation      = "23514"
	CodeInvalidTextRep      = "22P02"
	CodeSerializationFail   = "40001"
)

// errorMap defines how database-level SQLSTATE codes map
// to high-level application errors defined in pkg/errors.
var errorMap = map[string]error{
	CodeUniqueViolation:     appErr.ErrAlreadyExists,
	CodeForeignKeyViolation: appErr.ErrInvalidInput,
	CodeNotNullViolation:    appErr.ErrInvalidInput,
	CodeCheckViolation:      appErr.ErrInvalidInput,
	CodeInvalidTextRep:      appErr.ErrInvalidInput,
	CodeSerializationFail:   appErr.ErrConflict,
}

// MapSQLError standardizes raw SQL/driver errors into wrapped app-level errors.
//
// This should be used in all repositories when handling database operations.
func MapSQLError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Handle "no rows found"
	if errors.Is(err, sql.ErrNoRows) {
		return appErr.Wrap(context, appErr.ErrNotFound, err)
	}

	// Handle PostgreSQL-specific SQLSTATE errors
	var pqe pqError
	if errors.As(err, &pqe) {
		if mapped, ok := errorMap[pqe.SQLState()]; ok {
			return appErr.Wrap(context, mapped, err)
		}
		// Unknown SQLSTATE → internal
		return appErr.Wrap(context, appErr.ErrInternal, err)
	}

	// Any other error (driver, connection, etc.)
	return appErr.Wrap(context, appErr.ErrInternal, err)
}

// MapTxError wraps transaction-level errors (commit, rollback failures).
// Typically used after tx.Commit() or tx.Rollback().
func MapTxError(err error, context string) error {
	if err == nil {
		return nil
	}
	// Transaction-level errors are almost always internal.
	return appErr.Wrap(context, appErr.ErrInternal, err)
}
