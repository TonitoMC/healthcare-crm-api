package database

import (
	"database/sql"
	"errors"

	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// pqError defines the minimal interface implemented by PostgreSQL driver errors.
type pqError interface {
	SQLState() string
}

// -----------------------------------------------------------------------------
// PostgreSQL SQLSTATE code constants
// -----------------------------------------------------------------------------
const (
	CodeUniqueViolation     = "23505"
	CodeForeignKeyViolation = "23503"
	CodeNotNullViolation    = "23502"
	CodeCheckViolation      = "23514"
	CodeInvalidTextRep      = "22P02"
	CodeSerializationFail   = "40001"
)

// -----------------------------------------------------------------------------
// Error mapping: SQLSTATE → application-level error
// -----------------------------------------------------------------------------
var errorMap = map[string]error{
	CodeUniqueViolation:     appErr.ErrAlreadyExists,
	CodeForeignKeyViolation: appErr.ErrInvalidInput,   // invalid FK reference
	CodeNotNullViolation:    appErr.ErrIncompleteData, // missing required value
	CodeCheckViolation:      appErr.ErrInvalidInput,   // constraint validation failed
	CodeInvalidTextRep:      appErr.ErrInvalidRequest, // malformed literal or bad type
	CodeSerializationFail:   appErr.ErrConflict,       // concurrent write conflict
}

// -----------------------------------------------------------------------------
// MapSQLError
// -----------------------------------------------------------------------------

// MapSQLError standardizes raw SQL/driver errors into wrapped app-level errors.
//
// Always use this in repositories when handling database queries or commands.
// Example:
//
//	if err := r.db.Query(...); err != nil {
//	    return database.MapSQLError(err, "UserRepository.Create")
//	}
func MapSQLError(err error, context string) error {
	if err == nil {
		return nil
	}

	// "No rows found" case
	if errors.Is(err, sql.ErrNoRows) {
		return appErr.Wrap(context, appErr.ErrNotFound, err)
	}

	// PostgreSQL SQLSTATE error
	var pqe pqError
	if errors.As(err, &pqe) {
		if mapped, ok := errorMap[pqe.SQLState()]; ok {
			return appErr.Wrap(context, mapped, err)
		}
		// Unknown SQLSTATE → internal server error
		return appErr.Wrap(context, appErr.ErrInternal, err)
	}

	// Other (driver/connection) errors
	return appErr.Wrap(context, appErr.ErrInternal, err)
}

// -----------------------------------------------------------------------------
// MapTxError
// -----------------------------------------------------------------------------

// MapTxError wraps transaction-level errors (commit, rollback failures).
// Should be used after tx.Commit() or tx.Rollback() calls.
func MapTxError(err error, context string) error {
	if err == nil {
		return nil
	}
	return appErr.Wrap(context, appErr.ErrInternal, err)
}
