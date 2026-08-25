package projectpattern

import (
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
)

func (input *CreateProjectPatternInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.SourcePatternID, "source_pattern_id")...)

	errors = append(errors, validation.ValidateScale(input.Scale)...)

	return errors
}

func (input *GetProjectPatternByIDInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectPatternID, "project_pattern_id")...)

	return errors
}

func (input *ListProjectPatternsInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidatePagination(input.Page, input.PageSize)...)

	return errors
}

func (input *UpdateProjectPatternInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectPatternID, "project_pattern_id")...)

	if input.Name == nil && input.Scale == nil {
		errors = append(errors, apperror.NewFieldError(
			"project_pattern",
			apperror.FieldCodeRequired,
			"at least one project pattern field must be provided"),
		)
	}

	if input.Name != nil {
		errors = append(errors, validation.ValidateFieldForEmptiness("name", *input.Name)...)
		errors = append(errors, validation.ValidateMaxLen("name", *input.Name, config.MaxNameLen)...)
	}

	if input.Scale != nil {
		errors = append(errors, validation.ValidateScale(*input.Scale)...)
	}

	return errors
}

func (input *DeleteProjectPatternInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectPatternID, "project_pattern_id")...)

	return errors
}
