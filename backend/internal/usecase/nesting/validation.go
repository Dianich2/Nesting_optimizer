package nesting

import (
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
)

func (input *RunNestingInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectSurfaceID, "project_surface_id")...)

	if len(input.Patterns) == 0 {
		errors = append(errors, apperror.NewFieldError(
			"patterns",
			apperror.FieldCodeRequired,
			"patterns must not be empty",
		))
	}

	seenPatternIDs := make(map[int64]struct{})

	for _, pattern := range input.Patterns {
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

	for _, rotation := range input.AllowedRotations {
		errors = append(errors, validation.ValidateRotation(rotation)...)
	}

	return errors
}
