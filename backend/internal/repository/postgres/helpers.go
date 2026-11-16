package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

func uniqueViolationConstraint(err error) (string, bool) {
	var pqErr *pq.Error

	if !errors.As(err, &pqErr) {
		return "", false
	}

	if string(pqErr.Code) != postgresUniqueViolationCode {
		return "", false
	}

	return pqErr.Constraint, true
}

func execAffected(
	ctx context.Context,
	db DBTX,
	query string,
	args ...any,
) (int64, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func nullInt64ToPtr(
	value sql.NullInt64,
) *int64 {
	var int64Val *int64
	if value.Valid {
		t := value.Int64
		int64Val = &t
	}

	return int64Val
}
