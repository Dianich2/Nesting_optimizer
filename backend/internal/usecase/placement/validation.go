package placement

import (
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
)

func (input *CreatePlacementInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectSurfaceID, "project_surface_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectPatternID, "project_pattern_id")...)
	errors = append(errors, validation.ValidateCoordinate("x", input.X)...)
	errors = append(errors, validation.ValidateCoordinate("y", input.Y)...)
	errors = append(errors, validation.ValidateRotation(input.Rotation)...)

	return errors
}

func (input *GetPlacementByIDInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.PlacementID, "placement_id")...)

	return errors
}

func (input *ListPlacementsInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectSurfaceID, "project_surface_id")...)

	return errors
}

func (input *UpdatePlacementInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.PlacementID, "placement_id")...)

	if input.X == nil &&
		input.Y == nil &&
		input.Rotation == nil {
		errors = append(
			errors,
			apperror.NewFieldError(
				"placement",
				apperror.FieldCodeRequired,
				"at least one placement field must be provided",
			),
		)
	}

	if input.X != nil {
		errors = append(errors, validation.ValidateCoordinate("x", *input.X)...)
	}

	if input.Y != nil {
		errors = append(errors, validation.ValidateCoordinate("y", *input.Y)...)
	}

	if input.Rotation != nil {
		errors = append(errors, validation.ValidateRotation(*input.Rotation)...)
	}

	return errors
}

func (input *DeletePlacementInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.PlacementID, "placement_id")...)

	return errors
}
