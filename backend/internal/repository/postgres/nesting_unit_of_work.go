package postgres

import (
	"context"
	"errors"
	"fmt"
	"server_nesting_optimizer/internal/geometry"
	nestingusecase "server_nesting_optimizer/internal/usecase/nesting"

	"github.com/jmoiron/sqlx"
)

type NestingUnitOfWork struct {
	db            *sqlx.DB
	geometryCodec geometry.Codec
}

func NewNestingUnitOfWork(
	db *sqlx.DB,
	geometryCodec geometry.Codec,
) *NestingUnitOfWork {
	return &NestingUnitOfWork{
		db:            db,
		geometryCodec: geometryCodec,
	}
}

var _ nestingusecase.UnitOfWork = (*NestingUnitOfWork)(nil)

func (nuow *NestingUnitOfWork) WithinTransaction(
	ctx context.Context,
	fn func(repositories nestingusecase.TransactionRepositories) error,
) error {
	tx, err := nuow.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if panicValue := recover(); panicValue != nil {
			_ = tx.Rollback()
			panic(panicValue)
		}
	}()

	repositories := nestingusecase.TransactionRepositories{
		Placements: NewPlacementRepository(tx, nuow.geometryCodec),
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
