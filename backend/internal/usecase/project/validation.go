package project

import (
	"server_nesting_optimizer/pkg/apperror"
	"unicode/utf8"
)

func (pc *CreateProjectInput) Validate() []apperror.FieldError {
	errors := pc.validateRequiredFields()

	errors = append(errors, validateNameLen(pc.Name)...)
	errors = append(errors, validateDescriptionLen(pc.Description)...)
	errors = append(errors, validateID(pc.UserID, "user_id")...)

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

func (pc *GetProjectByIDInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validateID(pc.UserID, "user_id")...)
	errors = append(errors, validateID(pc.ProjectID, "id")...)

	return errors
}
