package nesting

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	optimizer "server_nesting_optimizer/internal/nesting"
	projecttransaction "server_nesting_optimizer/internal/usecase/project_transaction"
	"server_nesting_optimizer/pkg/apperror"
	"time"
)

type PreviewData struct {
	Problem      optimizer.Problem
	Input        RunNestingInput
	SnapshotHash [32]byte
	Result       optimizer.Result
	Duration     time.Duration
}
type StoredPreview struct {
	OperationKey string
	ID           string
	Data         PreviewData
	ExpiresAt    time.Time
	Saved        *RunNestingOutput
}
type PreviewRepository interface {
	FindByKey(context.Context, RunNestingInput, string) (StoredPreview, bool, error)
	Store(context.Context, StoredPreview) error
	Get(context.Context, RunNestingInput) (StoredPreview, error)
	MarkSaved(context.Context, string, RunNestingOutput) error
}

func (uc *RunNestingUseCase) retainPreview(ctx context.Context, input RunNestingInput, computed calculatedLayout, output *RunNestingOutput, operationKey string, finish func(context.Context, RunNestingOutput) error) error {
	encoded, err := json.Marshal(output)
	if err != nil {
		return apperror.Internal("encode preview result", err)
	}
	if len(encoded) > 8*1024*1024 {
		return apperror.Validation("preview result exceeds 8 MiB storage budget")
	}
	if uc.Previews == nil {
		return apperror.Internal("preview storage unavailable", nil)
	}
	record := StoredPreview{OperationKey: operationKey, ID: uuid.NewString(), ExpiresAt: time.Now().Add(15 * time.Minute), Data: PreviewData{
		Problem: computed.snapshot.Problem, Input: input, SnapshotHash: computed.snapshot.Hash, Result: computed.result, Duration: computed.duration,
	}}
	_, err = projecttransaction.Run(ctx, uc.Transactions, input.ProjectID, input.UserID, func(ctx context.Context) (bool, error) {
		if err := uc.Previews.Store(ctx, record); err != nil {
			return false, err
		}
		output.PreviewID = record.ID
		output.PreviewExpiresAt = record.ExpiresAt
		if finish != nil {
			if err := finish(ctx, *output); err != nil {
				return false, err
			}
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	output.PreviewID = record.ID
	output.PreviewExpiresAt = record.ExpiresAt
	return nil
}

func (uc *RunNestingUseCase) loadPreview(ctx context.Context, input RunNestingInput) (StoredPreview, error) {
	if _, err := parseOperationKey(input.PreviewID); err != nil {
		return StoredPreview{}, err
	}
	if uc.Previews == nil {
		return StoredPreview{}, apperror.Internal("preview storage unavailable", nil)
	}
	preview, err := uc.Previews.Get(ctx, input)
	if err != nil {
		return StoredPreview{}, err
	}
	original := input
	original.PreviewID = ""
	expected, err := Fingerprint(preview.Data.Input)
	if err != nil {
		return StoredPreview{}, err
	}
	actual, err := Fingerprint(original)
	if err != nil {
		return StoredPreview{}, err
	}
	if expected != actual {
		return StoredPreview{}, apperror.Conflict("preview parameters changed")
	}
	return preview, nil
}

func (uc *RunNestingUseCase) restorePreview(ctx context.Context, input RunNestingInput, result RunNestingOutput) (RunNestingOutput, error) {
	input.PreviewID = result.PreviewID
	record, err := uc.Previews.Get(ctx, input)
	if err != nil {
		if app, ok := apperror.As(err); ok && app.Code == apperror.AppCodeNotFound {
			err = &apperror.AppError{Code: apperror.Code("preview_expired"), Message: "preview expired or evicted; start a new operation"}
		}
		return RunNestingOutput{}, err
	}
	placements, err := validateResult(ctx, uc.geometryEngine, input.ProjectSurfaceID, record.Data.Problem, record.Data.Result)
	if err != nil {
		return RunNestingOutput{}, err
	}
	output := previewOutput(calculatedLayout{result: record.Data.Result, placements: placements})
	output.PreviewID = record.ID
	output.PreviewExpiresAt = record.ExpiresAt
	return output, nil
}
func (uc *RunNestingUseCase) GetPreviewOperation(ctx context.Context, user, project, surface int64, key string) (OperationStatus, error) {
	coordinated := *uc
	coordinated.Operations = uc.PreviewOperations
	status, err := coordinated.GetOperation(ctx, user, project, surface, key)
	if err == nil && status.Result != nil {
		if output, e := uc.restorePreview(ctx, RunNestingInput{UserID: user, ProjectID: project, ProjectSurfaceID: surface}, *status.Result); e != nil {
			if app, ok := apperror.As(e); ok && app.Code == apperror.Code("preview_expired") {
				return OperationStatus{Status: "failed", Failure: &OperationFailure{Code: app.Code, Message: app.Message}}, nil
			}
			return OperationStatus{}, e
		} else {
			status.Result = &output
		}
	}
	return status, err
}
func (uc *RunNestingUseCase) CancelPreviewOperation(ctx context.Context, user, project, surface int64, key string) (bool, error) {
	coordinated := *uc
	coordinated.Operations = uc.PreviewOperations
	return coordinated.CancelOperation(ctx, user, project, surface, key)
}
