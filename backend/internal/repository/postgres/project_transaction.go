package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"server_nesting_optimizer/pkg/apperror"
)

type projectTransactionKey struct{}

// ScopedDB routes repository calls through the transaction carried by the use case.
type ScopedDB struct{ DB *sqlx.DB }

func (s *ScopedDB) connection(ctx context.Context) DBTX {
	if tx, ok := ctx.Value(projectTransactionKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return s.DB
}
func (s *ScopedDB) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return s.connection(ctx).GetContext(ctx, dest, query, args...)
}
func (s *ScopedDB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return s.connection(ctx).SelectContext(ctx, dest, query, args...)
}
func (s *ScopedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.connection(ctx).ExecContext(ctx, query, args...)
}
func (s *ScopedDB) WithinProject(ctx context.Context, projectID, userID int64, fn func(context.Context) error) error {
	tx, err := s.DB.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return apperror.Internal("begin project transaction", err)
	}
	defer func() { _ = tx.Rollback() }()
	var id int64
	err = tx.GetContext(ctx, &id, "SELECT id FROM users WHERE id = $1 FOR KEY SHARE", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.NotFound("project not found")
	}
	if err != nil {
		return apperror.Internal("lock project owner", err)
	}
	err = tx.GetContext(ctx, &id, "SELECT id FROM projects WHERE id = $1 AND user_id = $2 FOR UPDATE", projectID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.NotFound("project not found")
	}
	if err != nil {
		return apperror.Internal("lock project", err)
	}
	if err = fn(context.WithValue(ctx, projectTransactionKey{}, tx)); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return apperror.Internal("commit project transaction", err)
	}
	return nil
}
