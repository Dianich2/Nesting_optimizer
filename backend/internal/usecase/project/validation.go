package project

import (
	"server_nesting_optimizer/pkg/apperror"
	"unicode/utf8"
)

func (pc *CreateProjectInput) Validate() []apperror.FieldError {
	errors := pc.validateRequiredFields()

	errors = append(errors, validateNameLen(pc.Name)...)
	errors = append(errors, validateDescriptionLen(pc.Description)...)
	errors = append(errors, validateUserID(pc.UserID)...)

	return errors
}

func (pc *CreateProjectInput) validateRequiredFields() []apperror.FieldError {
	var errors []apperror.FieldError

	if pc.Name == "" {
		errors = append(errors, apperror.NewFieldError(
			"name",
			apperror.FieldCodeRequired,
			"name must not be empty"),
		)
	}

	return errors
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

func validateDescriptionLen(
	description string,
) []apperror.FieldError {
	if utf8.RuneCountInString(description) > 2000 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"description",
				apperror.FieldCodeTooLong,
				"description is too long",
			),
		}
	}

	return nil
}

func validateUserID(
	userID int64,
) []apperror.FieldError {
	if userID <= 0 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"user_id",
				apperror.FieldCodeInvalid,
				"user_id must be greater than 0",
			),
		}
	}

	return nil
}
