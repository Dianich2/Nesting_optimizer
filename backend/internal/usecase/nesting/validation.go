package nesting

import (
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
)

const maxNestingInstances = 500
const maxNestingRotations = 36

func (input *RunNestingInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectSurfaceID, "project_surface_id")...)

	if input.Algorithm == "" {
		errors = append(errors, apperror.NewFieldError(
			"algorithm",
			apperror.FieldCodeRequired,
			"algorithm must not be empty",
		))
	}

	if input.Algorithm != "" && !input.Algorithm.IsValid() {
		errors = append(errors, apperror.NewFieldError(
			"algorithm",
			apperror.FieldCodeInvalid,
			"algorithm is invalid",
		))
	}

	if len(input.Patterns) == 0 {
		errors = append(errors, apperror.NewFieldError(
			"patterns",
			apperror.FieldCodeRequired,
			"patterns must not be empty",
		))
	}

	seenPatternIDs := make(map[int64]struct{})

	remaining := maxNestingInstances
	for _, pattern := range input.Patterns {
		if pattern.Quantity > remaining {
			errors = append(errors, apperror.NewFieldError("quantity", apperror.FieldCodeInvalid, "a nesting run supports at most 500 instances"))
			break
		}
		if pattern.Quantity > 0 {
			remaining -= pattern.Quantity
		}

		errors = append(errors, validation.ValidateID(pattern.ProjectPatternID, "project_pattern_id")...)
		errors = append(errors, validation.ValidateQuantity(pattern.Quantity)...)

		if pattern.ProjectPatternID <= 0 {
			continue
		}

		_, ok := seenPatternIDs[pattern.ProjectPatternID]
		if ok {
			errors = append(errors, apperror.NewFieldError(
				"patterns",
				apperror.FieldCodeInvalid,
				"patterns must not repeat",
			))

			continue
		}

		seenPatternIDs[pattern.ProjectPatternID] = struct{}{}
	}

	if len(input.AllowedRotations) == 0 {
		errors = append(errors, apperror.NewFieldError(
			"allowed_rotations",
			apperror.FieldCodeRequired,
			"allowed rotations must not be empty",
		))
	}

	if len(input.AllowedRotations) > maxNestingRotations {
		return append(errors, apperror.NewFieldError("allowed_rotations", apperror.FieldCodeInvalid, "a nesting run supports at most 36 rotations"))
	}
	for _, rotation := range input.AllowedRotations {
		errors = append(errors, validation.ValidateRotation(rotation)...)
	}

	return errors
}
