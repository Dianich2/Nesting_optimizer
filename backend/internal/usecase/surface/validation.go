package surface

import (
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
)

func (input *CreateSurfaceInput) Validate() []apperror.FieldError {
	errors := input.validateRequiredFields()

	errors = append(errors, validation.ValidateMaxLen("name", input.Name, config.MaxNameLen)...)
	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)

	return errors
}

func (input *CreateSurfaceInput) validateRequiredFields() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateFieldForEmptiness("name", input.Name)...)

	return errors
}

func (input *GetSurfaceByIDInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.SurfaceID, "id")...)
	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)

	return errors
}

func (input *ListSurfacesInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidatePagination(input.Page, input.PageSize)...)

	return errors
}

func (input *UpdateSurfaceInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.SurfaceID, "id")...)
	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)

	if input.Name == nil && input.Geometry == nil {
		errors = append(errors, apperror.NewFieldError(
			"surface",
			apperror.FieldCodeRequired,
			"at least one surface field must be provided"),
		)
	}

	if input.Name != nil {
		errors = append(errors, validation.ValidateFieldForEmptiness("name", *input.Name)...)
		errors = append(errors, validation.ValidateMaxLen("name", *input.Name, config.MaxNameLen)...)
	}

	return errors
}

func (input *DeleteSurfaceInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.SurfaceID, "id")...)
	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)

	return errors
}
