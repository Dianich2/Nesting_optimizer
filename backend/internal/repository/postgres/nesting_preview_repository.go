package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	nestingcase "server_nesting_optimizer/internal/usecase/nesting"
	"server_nesting_optimizer/pkg/apperror"
	"time"
)

type NestingPreviewRepository struct{ db DBTX }

func NewNestingPreviewRepository(db DBTX) *NestingPreviewRepository {
	return &NestingPreviewRepository{db: db}
}

func (r *NestingPreviewRepository) Store(ctx context.Context, record nestingcase.StoredPreview) error {
	input := record.Data.Input
	payload, err := json.Marshal(record.Data)
	if err != nil {
		return err
	}
	if len(payload) > 8*1024*1024 {
		return apperror.Validation("preview exceeds 8 MiB storage budget")
	}
	// The caller holds a short project transaction. This per-user lock also
	// serializes preview quotas for different projects of the same owner.
	if _, err = r.db.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", -input.UserID); err != nil {
		return err
	}
	if _, err = r.db.ExecContext(ctx, "DELETE FROM nesting_previews WHERE user_id=$1 AND expires_at<=clock_timestamp()", input.UserID); err != nil {
		return err
	}
	var count int
	if err = r.db.GetContext(ctx, &count, "SELECT count(*) FROM nesting_previews WHERE user_id=$1", input.UserID); err != nil {
		return err
	}
	if count >= 10 {
		// A bounded cache evicts the oldest preview. Its subsequent use returns
		// a missing/expired error, while committed operations remain replayable.
		if _, err = r.db.ExecContext(ctx, `DELETE FROM nesting_previews WHERE id=(SELECT id FROM nesting_previews WHERE user_id=$1 ORDER BY created_at,id LIMIT 1)`, input.UserID); err != nil {
			return err
		}
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO nesting_previews(id,user_id,project_id,project_surface_id,payload,expires_at,operation_key) VALUES($1,$2,$3,$4,$5::jsonb,$6,NULLIF($7,'')::uuid)`, record.ID, input.UserID, input.ProjectID, input.ProjectSurfaceID, string(payload), record.ExpiresAt, record.OperationKey)
	var databaseError *pq.Error
	if errors.As(err, &databaseError) && databaseError.Code == "23514" {
		return apperror.Validation("preview exceeds 8 MiB storage budget")
	}
	return err
}
func (r *NestingPreviewRepository) Get(ctx context.Context, input nestingcase.RunNestingInput) (nestingcase.StoredPreview, error) {
	var row struct {
		Payload []byte    `db:"payload"`
		Saved   []byte    `db:"saved_result"`
		Expires time.Time `db:"expires_at"`
	}
	err := r.db.GetContext(ctx, &row, `SELECT payload,saved_result,expires_at FROM nesting_previews WHERE id=$1 AND user_id=$2 AND project_id=$3 AND project_surface_id=$4 AND expires_at>clock_timestamp() FOR UPDATE`, input.PreviewID, input.UserID, input.ProjectID, input.ProjectSurfaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nestingcase.StoredPreview{}, apperror.NotFound("preview expired or not found")
	}
	if err != nil {
		return nestingcase.StoredPreview{}, err
	}
	output := nestingcase.StoredPreview{ID: input.PreviewID, ExpiresAt: row.Expires}
	if err = json.Unmarshal(row.Payload, &output.Data); err != nil {
		return output, err
	}
	if len(row.Saved) > 0 {
		output.Saved = &nestingcase.RunNestingOutput{}
		if err = json.Unmarshal(row.Saved, output.Saved); err != nil {
			return output, err
		}
	}
	return output, nil
}
func (r *NestingPreviewRepository) MarkSaved(ctx context.Context, id string, result nestingcase.RunNestingOutput) error {
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	written, err := r.db.ExecContext(ctx, "UPDATE nesting_previews SET saved_result=$2::jsonb WHERE id=$1", id, string(encoded))
	if err != nil {
		return err
	}
	count, err := written.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return apperror.NotFound("preview expired or not found")
	}
	return nil
}

func (r *NestingPreviewRepository) PurgeExpired(ctx context.Context, limit int) (int64, error) {
	if limit < 1 || limit > 1000 {
		limit = 1000
	}
	result, err := r.db.ExecContext(ctx, `WITH expired AS (
        SELECT id FROM nesting_previews WHERE expires_at<=clock_timestamp()
        ORDER BY expires_at,id LIMIT $1 FOR UPDATE SKIP LOCKED
    ) DELETE FROM nesting_previews p USING expired e WHERE p.id=e.id`, limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *NestingPreviewRepository) FindByKey(ctx context.Context, input nestingcase.RunNestingInput, key string) (nestingcase.StoredPreview, bool, error) {
	var id string
	err := r.db.GetContext(ctx, &id, `SELECT id::text FROM nesting_previews WHERE user_id=$1 AND project_id=$2 AND project_surface_id=$3 AND operation_key=$4 AND expires_at>clock_timestamp()`, input.UserID, input.ProjectID, input.ProjectSurfaceID, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nestingcase.StoredPreview{}, false, nil
	}
	if err != nil {
		return nestingcase.StoredPreview{}, false, err
	}
	input.PreviewID = id
	record, err := r.Get(ctx, input)
	if app, ok := apperror.As(err); ok && app.Code == apperror.AppCodeNotFound {
		return nestingcase.StoredPreview{}, false, nil
	}
	return record, err == nil, err
}
