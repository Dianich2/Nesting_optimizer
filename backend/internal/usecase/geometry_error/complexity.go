package geometryerror

import (
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

// Complexity keeps the original cause for errors.Is while exposing only the
// established validation_error / geometry.too_long contract to HTTP clients.
func Complexity(cause error) *apperror.AppError {
	result := apperror.Validation("geometry complexity budget exceeded", apperror.NewFieldError("geometry", apperror.FieldCodeTooLong, "geometry complexity budget exceeded"))
	result.Err = cause
	return result
}

// Restore rehydrates the sentinel only for the canonical persisted complexity
// failure. No schema/new public code is required and unrelated validation
// failures must not acquire a complexity cause on idempotent replay.
func Restore(result *apperror.AppError) *apperror.AppError {
	if result.Err == nil && result.Code == apperror.AppCodeValidation && result.Message == "geometry complexity budget exceeded" && len(result.Details) == 1 && result.Details[0].Field == "geometry" && result.Details[0].Code == apperror.FieldCodeTooLong {
		result.Err = geometry.ErrComplexity
	}
	return result
}
