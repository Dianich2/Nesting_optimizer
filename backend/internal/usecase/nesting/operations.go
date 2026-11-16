package nesting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"

	"github.com/google/uuid"
	"server_nesting_optimizer/pkg/apperror"
)

type OperationRepository interface {
	Find(context.Context, int64, uuid.UUID) ([]byte, *OperationStatus, error)
	Save(context.Context, int64, uuid.UUID, []byte, OperationStatus) error
	Attempt(context.Context, func() error) error
	Lock(context.Context, int64, uuid.UUID) error
	Running(context.Context, int64, uuid.UUID) (bool, error)
}

func parseOperationKey(raw string) (uuid.UUID, error) {
	key, err := uuid.Parse(raw)
	if err != nil || key == uuid.Nil || key.String() != raw {
		return uuid.Nil, apperror.Validation("invalid idempotency key",
			apperror.NewFieldError("idempotency_key", apperror.FieldCodeInvalid, "expected a non-nil UUID in canonical form"))
	}
	return key, nil
}

func (uc *RunNestingUseCase) runOperation(ctx context.Context, input RunNestingInput, key uuid.UUID, hash []byte) (OperationStatus, error) {
	if _, err := uc.projectSurfaceRepository.GetByID(ctx, input.UserID, input.ProjectID, input.ProjectSurfaceID); err != nil {
		return OperationStatus{}, operationSurfaceError(err)
	}
	if err := uc.Operations.Lock(ctx, input.ProjectSurfaceID, key); err != nil {
		return OperationStatus{}, apperror.Internal("lock nesting operation", err)
	}
	savedHash, saved, err := uc.Operations.Find(ctx, input.ProjectSurfaceID, key)
	if err != nil {
		return OperationStatus{}, apperror.Internal("read nesting operation", err)
	}
	if saved != nil {
		if !bytes.Equal(hash, savedHash) {
			return OperationStatus{}, apperror.Conflict("idempotency key already used for different nesting parameters")
		}
		return *saved, nil
	}
	status := OperationStatus{Status: "completed"}
	err = uc.Operations.Attempt(ctx, func() error {
		result, err := uc.execute(ctx, input)
		if err == nil {
			status.Result = &result
		}
		return err
	})
	if err != nil {
		failure, ok := apperror.As(err)
		if !ok || failure.Code == apperror.AppCodeInternal {
			return OperationStatus{}, err
		}
		status = OperationStatus{Status: "failed", Failure: &OperationFailure{
			Code: failure.Code, Message: failure.Message, Details: failure.Details,
		}}
	}
	if err = uc.Operations.Save(ctx, input.ProjectSurfaceID, key, hash, status); err != nil {
		return OperationStatus{}, apperror.Internal("save nesting operation", err)
	}
	return status, nil
}

func fingerprint(input RunNestingInput) ([32]byte, error) {
	input.AllowedRotations = NormalizeAllowedRotations(input.AllowedRotations)
	encoded, err := json.Marshal(input)
	return sha256.Sum256(encoded), err
}

type OperationStatus struct {
	Status  string
	Result  *RunNestingOutput
	Failure *OperationFailure
}

type OperationFailure struct {
	Code    apperror.Code
	Message string
	Details []apperror.FieldError
}

func (uc *RunNestingUseCase) GetOperation(ctx context.Context, userID, projectID, surfaceID int64, rawKey string) (OperationStatus, error) {
	key, err := parseOperationKey(rawKey)
	if err != nil {
		return OperationStatus{}, err
	}
	if _, err = uc.projectSurfaceRepository.GetByID(ctx, userID, projectID, surfaceID); err != nil {
		return OperationStatus{}, operationSurfaceError(err)
	}
	running, err := uc.Operations.Running(ctx, surfaceID, key)
	if err != nil {
		return OperationStatus{}, apperror.Internal("read nesting operation lock", err)
	}
	_, saved, err := uc.Operations.Find(ctx, surfaceID, key)
	if err != nil {
		return OperationStatus{}, apperror.Internal("read nesting operation", err)
	}
	if saved != nil {
		return *saved, nil
	}
	if running {
		return OperationStatus{Status: "running"}, nil
	}
	return OperationStatus{Status: "unknown"}, nil
}

func operationSurfaceError(err error) error {
	if errors.Is(err, domainprojectsurface.ErrNotFound) {
		return apperror.NotFound("project surface not found")
	}
	return apperror.Internal("read project surface", err)
}
