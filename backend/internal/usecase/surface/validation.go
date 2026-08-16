package surface

import (
	"server_nesting_optimizer/pkg/apperror"
	"unicode/utf8"
)

func (input *CreateSurfaceInput) Validate() []apperror.FieldError {
	errors := input.validateRequiredFields()

	errors = append(errors, validateNameLen(input.Name)...)
	errors = append(errors, validateID(input.UserID, "user_id")...)

	return errors
}

func (input *CreateSurfaceInput) validateRequiredFields() []apperror.FieldError {
	var errs []apperror.FieldError

	if input.Name == "" {
		errs = append(
			errs,
			apperror.NewFieldError(
				"name",
				apperror.FieldCodeRequired,
				"name must not be empty",
			),
		)
	}

	return errs
}

func validateNameLen(
	name string,
) []apperror.FieldError {
	if utf8.RuneCountInString(name) > 150 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"name",
				apperror.FieldCodeTooLong,
				"name is too long",
			),
		}
	}

	return nil
}

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
