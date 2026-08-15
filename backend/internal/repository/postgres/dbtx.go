package postgres

import (
	"context"
	"database/sql"
)

type DBTX interface {
	GetContext(
		ctx context.Context,
		dest any,
		query string,
		args ...any,
	) error

	ExecContext(
		ctx context.Context,
		query string,
		args ...any,
	) (sql.Result, error)

	SelectContext(
		ctx context.Context,
		dest any,
		query string,
		args ...any,
	) error
}
