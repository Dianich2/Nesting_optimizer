package handler

import (
	"context"
	"errors"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	nestingusecase "server_nesting_optimizer/internal/usecase/nesting"
	"server_nesting_optimizer/pkg/apperror"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const nestingDeadline = nestingusecase.ExecutionTimeout

type nestingExecution struct {
	preview                      bool
	userID, projectID, surfaceID int64
	key                          string
}

type activeNesting struct {
	context context.Context
	cancel  context.CancelFunc
	callers int
}

type nestingExecutions struct {
	mu     sync.Mutex
	active map[nestingExecution]*activeNesting
}

func (executions *nestingExecutions) begin(parent context.Context, key nestingExecution) (context.Context, func()) {
	if key.key == "" {
		return context.WithTimeout(context.WithoutCancel(parent), nestingDeadline)
	}
	executions.mu.Lock()
	defer executions.mu.Unlock()
	if executions.active == nil {
		executions.active = make(map[nestingExecution]*activeNesting)
	}
	job := executions.active[key]
	if job == nil {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), nestingDeadline)
		job = &activeNesting{context: ctx, cancel: cancel}
		executions.active[key] = job
	}
	job.callers++
	var once sync.Once
	return job.context, func() {
		once.Do(func() {
			executions.mu.Lock()
			defer executions.mu.Unlock()
			job.callers--
			if job.callers == 0 {
				job.cancel()
				delete(executions.active, key)
			}
		})
	}
}

func (executions *nestingExecutions) cancel(key nestingExecution) bool {
	executions.mu.Lock()
	defer executions.mu.Unlock()
	job := executions.active[key]
	if job == nil {
		return false
	}
	job.cancel()
	return true
}

func nestingResponseError(c fiber.Ctx, ctx context.Context, err error) error {
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return c.Status(fiber.StatusRequestTimeout).JSON(httperror.NewErrorResponse("nesting_timeout", "nesting execution timed out", nil))
	case errors.Is(ctx.Err(), context.Canceled):
		return c.Status(fiber.StatusConflict).JSON(httperror.NewErrorResponse("nesting_cancelled", "nesting was cancelled", nil))
	default:
		return err
	}
}

func parseExecutionKey(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	key, err := uuid.Parse(raw)
	if err != nil || key == uuid.Nil || key.String() != raw {
		return "", apperror.Validation("invalid idempotency key",
			apperror.NewFieldError("idempotency_key", apperror.FieldCodeInvalid, "expected a non-nil UUID in canonical form"))
	}
	return key.String(), nil
}

// @Summary Cancel an active nesting calculation
// @Param project_id path int true "Project ID"
// @Param project_surface_id path int true "Project surface ID"
// @Param kind query string false "Operation type: save (default) or preview"
// @Param operation_key path string true "Execution UUID"
// @Success 204
// @Failure 404
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{project_surface_id}/nesting/operations/{operation_key} [delete]
func (h *NestingHandler) Cancel(c fiber.Ctx) error {
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}
	surfaceID, err := getIDFromPath(c, "project_surface_id")
	if err != nil {
		return err
	}
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	key, err := parseExecutionKey(c.Params("operation_key"))
	if err != nil {
		return err
	}
	cancel := h.runNestingUseCase.CancelOperation
	if c.Query("kind") == "preview" {
		cancel = h.runNestingUseCase.CancelPreviewOperation
	} else if c.Query("kind") != "" && c.Query("kind") != "save" {
		return apperror.Validation("invalid operation kind")
	}
	cancelled, err := cancel(c.Context(), userID, projectID, surfaceID, key)
	if err != nil {
		return httperror.Handle(c, err)
	}
	if !cancelled {
		return c.Status(404).JSON(httperror.NewErrorResponse("nesting_not_running", "nesting is not running", nil))
	}

	return c.SendStatus(fiber.StatusNoContent)
}
