package project

import (
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
)

func (pc *CreateProjectInput) Validate() []apperror.FieldError {
	errors := pc.validateRequiredFields()

	errors = append(errors, validation.ValidateMaxLen("name", pc.Name, config.MaxNameLen)...)
	errors = append(errors, validation.ValidateMaxLen("description", pc.Description, config.MaxDescriptionLen)...)
	errors = append(errors, validation.ValidateID(pc.UserID, "user_id")...)

	return errors
}

func (pc *CreateProjectInput) validateRequiredFields() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateFieldForEmptiness("name", pc.Name)...)

	return errors
}

func (pc *GetProjectByIDInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(pc.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(pc.ProjectID, "id")...)

	return errors
}

func (pc *ListProjectsInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(pc.UserID, "user_id")...)
	errors = append(errors, validation.ValidatePagination(pc.Page, pc.PageSize)...)

	return errors
}

func (pc *UpdateProjectInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	if pc.Name == nil && pc.Description == nil {
		errors = append(errors, apperror.NewFieldError(
			"project",
			apperror.FieldCodeRequired,
			"at least one project field must be provided"),
		)
	}

	if pc.Name != nil {
		errors = append(errors, validation.ValidateFieldForEmptiness("name", *pc.Name)...)
		errors = append(errors, validation.ValidateMaxLen("name", *pc.Name, config.MaxNameLen)...)
	}

	if pc.Description != nil {
		errors = append(errors, validation.ValidateMaxLen("description", *pc.Description, config.MaxDescriptionLen)...)
	}

	errors = append(errors, validation.ValidateID(pc.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(pc.ProjectID, "id")...)

	return errors
}

func (pc *DeleteProjectInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(pc.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(pc.ProjectID, "id")...)

	return errors
}
