package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/lib/pq"
)

// mapDBError translates PostgreSQL driver errors into domain errors.
func mapDBError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505": // unique_violation
			return domain.ErrAlreadyExists
		case "23503": // foreign_key_violation
			return domain.ErrNotFound
		}
	}
	return err
}

// checkRowsAffected returns ErrNotFound when a write affected no rows.
func checkRowsAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
