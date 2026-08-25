package pattern

import (
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
)

func (input *CreatePatternInput) Validate() []apperror.FieldError {
	errors := input.validateRequiredFields()

	errors = append(errors, validation.ValidateMaxLen("name", input.Name, config.MaxNameLen)...)
	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)

	return errors
}

func (input *CreatePatternInput) validateRequiredFields() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateFieldForEmptiness("name", input.Name)...)

	return errors
}

func (input *GetPatternByIDInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.PatternID, "id")...)
	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)

	return errors
}

func (input *ListPatternsInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidatePagination(input.Page, input.PageSize)...)

	return errors
}

func (input *UpdatePatternInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.PatternID, "id")...)
	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)

	if input.Name == nil && input.Geometry == nil {
		errors = append(errors, apperror.NewFieldError(
			"pattern",
			apperror.FieldCodeRequired,
			"at least one pattern field must be provided"),
		)
	}

	if input.Name != nil {
		errors = append(errors, validation.ValidateFieldForEmptiness("name", *input.Name)...)
		errors = append(errors, validation.ValidateMaxLen("name", *input.Name, config.MaxNameLen)...)
	}

	return errors
}

func (input *DeletePatternInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.PatternID, "id")...)
	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)

	return errors
}
