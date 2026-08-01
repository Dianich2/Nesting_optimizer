package postgres

import (
	"context"
	"errors"
	"fmt"

	userusecase "server_nesting_optimizer/internal/usecase/user"

	"github.com/jmoiron/sqlx"
)

type UnitOfWork struct {
	db *sqlx.DB
}

func NewUnitOfWork(
	db *sqlx.DB,
) *UnitOfWork {
	return &UnitOfWork{
		db: db,
	}
}

var _ userusecase.UnitOfWork = (*UnitOfWork)(nil)
var _ userusecase.UserRepository = (*UserRepository)(nil)
var _ userusecase.SessionRepository = (*SessionRepository)(nil)

func (uow *UnitOfWork) WithinTransaction(
	ctx context.Context,
	fn func(repositories userusecase.TransactionRepositories) error,
) error {
	tx, err := uow.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if panicValue := recover(); panicValue != nil {
			_ = tx.Rollback()
			panic(panicValue)
		}
	}()

	repositories := userusecase.TransactionRepositories{
		Users:    NewUserRepository(tx),
		Sessions: NewSessionRepository(tx),
	}

	if callbackErr := fn(repositories); callbackErr != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(
				callbackErr,
				fmt.Errorf("rollback transaction: %w", rollbackErr),
			)
		}

		return callbackErr
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
