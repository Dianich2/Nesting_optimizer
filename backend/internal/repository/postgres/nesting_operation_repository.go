package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	nestingcase "server_nesting_optimizer/internal/usecase/nesting"
)

type NestingOperationRepository struct{ db DBTX }

func NewNestingOperationRepository(db DBTX) *NestingOperationRepository {
	return &NestingOperationRepository{db: db}
}

const operationLockID = `hashtextextended($1::text || ':' || $2::text, 2738504)`

func (r *NestingOperationRepository) Lock(ctx context.Context, surfaceID int64, key uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `SELECT pg_advisory_xact_lock(`+operationLockID+`)`, surfaceID, key.String())
	return err
}

func (r *NestingOperationRepository) Running(ctx context.Context, surfaceID int64, key uuid.UUID) (bool, error) {
	var acquired bool
	err := r.db.GetContext(ctx, &acquired, `SELECT pg_try_advisory_xact_lock(`+operationLockID+`)`, surfaceID, key.String())
	return !acquired, err
}

// Attempt is called under the project transaction. A failed attempt rolls back
// layout writes before the use case can record a terminal rejection.
func (r *NestingOperationRepository) Attempt(ctx context.Context, run func() error) error {
	if _, err := r.db.ExecContext(ctx, "SAVEPOINT nesting_attempt"); err != nil {
		return err
	}
	if err := run(); err != nil {
		if _, rollbackErr := r.db.ExecContext(ctx, "ROLLBACK TO SAVEPOINT nesting_attempt"); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	_, err := r.db.ExecContext(ctx, "RELEASE SAVEPOINT nesting_attempt")
	return err
}

func (r *NestingOperationRepository) Find(ctx context.Context, surfaceID int64, key uuid.UUID) ([]byte, *nestingcase.OperationStatus, error) {
	var row struct {
		Hash     []byte `db:"request_hash"`
		Response []byte `db:"response"`
	}
	err := r.db.GetContext(ctx, &row, `SELECT request_hash, response FROM nesting_operations WHERE project_surface_id=$1 AND operation_key=$2`, surfaceID, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var output nestingcase.OperationStatus
	if err = json.Unmarshal(row.Response, &output); err != nil {
		return nil, nil, err
	}
	return row.Hash, &output, nil
}

func (r *NestingOperationRepository) Save(ctx context.Context, surfaceID int64, key uuid.UUID, hash []byte, result nestingcase.OperationStatus) error {
	response, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO nesting_operations (project_surface_id,operation_key,request_hash,response) VALUES ($1,$2,$3,$4::jsonb)`, surfaceID, key, hash, string(response))
	return err
}
