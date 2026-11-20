package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	nestingcase "server_nesting_optimizer/internal/usecase/nesting"
	"server_nesting_optimizer/pkg/apperror"
	"strings"
	"time"
)

type NestingOperationRepository struct {
	db      DBTX
	preview bool
}

func NewNestingOperationRepository(db DBTX) *NestingOperationRepository {
	return &NestingOperationRepository{db: db}
}

func NewPreviewOperationRepository(db DBTX) *NestingOperationRepository {
	return &NestingOperationRepository{db: db, preview: true}
}
func (r *NestingOperationRepository) query(q string) string {
	if r.preview {
		return strings.ReplaceAll(q, "nesting_operations", "nesting_preview_operations")
	}
	return q
}

func (r *NestingOperationRepository) Reserve(ctx context.Context, surfaceID int64, key uuid.UUID, hash []byte, token uuid.UUID, expires time.Time) (bool, error) {
	result, err := r.db.ExecContext(ctx, r.query(`INSERT INTO nesting_operations(project_surface_id,operation_key,request_hash,response,state,lease_token,lease_expires_at)
        VALUES($1,$2,$3,'{"Status":"running"}','running',$4,clock_timestamp()+($5*interval '1 millisecond')) ON CONFLICT DO NOTHING`), surfaceID, key, hash, token, max(int64(0), time.Until(expires).Milliseconds()))
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count == 1, err
}

func (r *NestingOperationRepository) Find(ctx context.Context, surfaceID int64, key uuid.UUID) ([]byte, *nestingcase.OperationStatus, error) {
	_, err := r.db.ExecContext(ctx, r.query(`UPDATE nesting_operations SET state='failed',response='{"Status":"failed","Failure":{"Code":"nesting_timeout","Message":"nesting worker deadline expired; start a new operation"}}'
		WHERE project_surface_id=$1 AND operation_key=$2 AND state='running' AND lease_expires_at <= clock_timestamp()`), surfaceID, key)
	if err != nil {
		return nil, nil, err
	}
	var row struct {
		Hash     []byte `db:"request_hash"`
		Response []byte `db:"response"`
	}
	err = r.db.GetContext(ctx, &row, r.query(`SELECT request_hash,response FROM nesting_operations WHERE project_surface_id=$1 AND operation_key=$2`), surfaceID, key)
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

func (r *NestingOperationRepository) Finish(ctx context.Context, surfaceID int64, key, token uuid.UUID, status nestingcase.OperationStatus) error {
	if r.preview && status.Result != nil {
		result := *status.Result
		result.Placements = nil
		result.Unplaced = nil
		status.Result = &result
	}
	response, err := json.Marshal(status)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, r.query(`UPDATE nesting_operations SET state=$4,response=$5::jsonb
        WHERE project_surface_id=$1 AND operation_key=$2 AND lease_token=$3 AND state='running'
        AND ($4 <> 'completed' OR (NOT cancel_requested AND lease_expires_at > clock_timestamp()))`), surfaceID, key, token, status.Status, string(response))
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		var row struct {
			Cancelled bool `db:"cancel_requested"`
			Expired   bool `db:"expired"`
		}
		lookup := r.db.GetContext(ctx, &row, r.query(`SELECT cancel_requested,lease_expires_at<=clock_timestamp() AS expired FROM nesting_operations WHERE project_surface_id=$1 AND operation_key=$2 AND lease_token=$3`), surfaceID, key, token)
		if errors.Is(lookup, sql.ErrNoRows) {
			return apperror.NotFound("nesting operation not found")
		}
		if lookup != nil {
			return lookup
		}
		if row.Cancelled {
			return &apperror.AppError{Code: apperror.AppCodeCancelled, Message: "nesting was cancelled"}
		}
		if row.Expired {
			return &apperror.AppError{Code: apperror.AppCodeTimeout, Message: "nesting execution timed out"}
		}
		return &apperror.AppError{Code: apperror.AppCodeInterrupted, Message: "nesting operation is no longer active"}
	}
	return nil
}

func (r *NestingOperationRepository) Cancel(ctx context.Context, surfaceID int64, key uuid.UUID) (bool, error) {
	result, err := r.db.ExecContext(ctx, r.query(`UPDATE nesting_operations SET cancel_requested=TRUE WHERE project_surface_id=$1 AND operation_key=$2 AND state='running' AND lease_expires_at>clock_timestamp()`), surfaceID, key)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count == 1, err
}
func (r *NestingOperationRepository) CancellationRequested(ctx context.Context, surfaceID int64, key, token uuid.UUID) (bool, error) {
	var active bool
	err := r.db.GetContext(ctx, &active, r.query(`SELECT state='running' AND NOT cancel_requested AND lease_expires_at>clock_timestamp() FROM nesting_operations WHERE project_surface_id=$1 AND operation_key=$2 AND lease_token=$3`), surfaceID, key, token)
	if errors.Is(err, sql.ErrNoRows) {
		return false, apperror.NotFound("nesting operation not found")
	}
	return !active, err
}
