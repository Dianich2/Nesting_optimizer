package placement

import (
	"math"
	"server_nesting_optimizer/pkg/apperror"
)

func validateID(
	id int64,
	fieldName string,
) []apperror.FieldError {
	if id <= 0 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				fieldName,
				apperror.FieldCodeInvalid,
				"id must be greater than 0",
			),
		}
	}

	return nil
}

func validateCoordinate(
	field string,
	value float64,
) []apperror.FieldError {
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return []apperror.FieldError{
			apperror.NewFieldError(
				field,
				apperror.FieldCodeInvalid,
				"coordinate must not be NaN or Inf",
			),
		}
	}

	return nil
}

func validateRotation(
	rotation float64,
) []apperror.FieldError {
	if math.IsInf(rotation, 0) || math.IsNaN(rotation) {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"rotation",
				apperror.FieldCodeInvalid,
				"rotation must not be NaN or Inf",
			),
		}
	}

	return nil
}

func (input *CreatePlacementInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validateID(input.UserID, "user_id")...)
	errors = append(errors, validateID(input.ProjectID, "project_id")...)
	errors = append(errors, validateID(input.ProjectSurfaceID, "project_surface_id")...)
	errors = append(errors, validateID(input.ProjectPatternID, "project_pattern_id")...)
	errors = append(errors, validateCoordinate("x", input.X)...)
	errors = append(errors, validateCoordinate("y", input.Y)...)
	errors = append(errors, validateRotation(input.Rotation)...)

	return errors
}
