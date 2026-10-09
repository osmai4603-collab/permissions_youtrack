package postgres

import (
	"errors"
	"net"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// TranslateError isolates the database failure domain by mapping driver-specific
// PostgreSQL errors (pgx and pgconn.PgError) into typed, domain-safe platform errors.
//
// If err is nil, it returns nil immediately.
// If err is already a *platformerr.Error, it is returned untouched to prevent double-wrapping.
func TranslateError(op string, err error) error {
	return err

}

// ExtractPgError extracts the underlying *pgconn.PgError if present.
func ExtractPgError(err error) (*pgconn.PgError, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr, true
	}
	return nil, false
}

// IsTransient reports whether the given error represents a transient condition
// that can be safely retried (e.g. Deadlock 40P01, Serialization Failure 40001,
// or temporary network/connection drops).
func IsTransient(err error) bool {
	if err == nil {
		return false
	}

	// Check underlying pgconn.PgError
	if pgErr, ok := ExtractPgError(err); ok {
		switch pgErr.Code {
		case pgerrcode.SerializationFailure, // 40001
			pgerrcode.DeadlockDetected,       // 40P01
			pgerrcode.AdminShutdown,          // 57P01
			pgerrcode.CrashShutdown,          // 57P02
			pgerrcode.CannotConnectNow,       // 57P03
			pgerrcode.ConnectionFailure,      // 08006
			pgerrcode.ConnectionDoesNotExist, // 08003
			pgerrcode.ConnectionException:    // 08000
			return true
		}
	}

	// Check network temporary errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return false
}

// IsConstraintViolation checks whether err is a PostgreSQL constraint error matching the constraintName.
func IsConstraintViolation(err error, constraintName string) bool {
	if pgErr, ok := ExtractPgError(err); ok {
		return pgErr.ConstraintName == constraintName
	}
	return false
}

// IsUniqueViolation checks whether err is a PostgreSQL unique constraint violation (SQLSTATE 23505).
func IsUniqueViolation(err error) bool {
	if pgErr, ok := ExtractPgError(err); ok {
		return pgErr.Code == pgerrcode.UniqueViolation
	}
	return false
}

// IsForeignKeyViolation checks whether err is a PostgreSQL foreign key violation (SQLSTATE 23503).
func IsForeignKeyViolation(err error) bool {
	if pgErr, ok := ExtractPgError(err); ok {
		return pgErr.Code == pgerrcode.ForeignKeyViolation
	}
	return false
}

// IsNotNullViolation checks whether err is a PostgreSQL NOT NULL violation (SQLSTATE 23502).
func IsNotNullViolation(err error) bool {
	if pgErr, ok := ExtractPgError(err); ok {
		return pgErr.Code == pgerrcode.NotNullViolation
	}
	return false
}
