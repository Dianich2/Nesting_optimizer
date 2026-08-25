package projectsurface

import (
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
)

func (input *CreateProjectSurfaceInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.SourceSurfaceID, "source_surface_id")...)

	errors = append(errors, validation.ValidateScale(input.Scale)...)

	return errors
}

func (input *GetProjectSurfaceByIDInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectSurfaceID, "project_surface_id")...)

	return errors
}

func (input *ListProjectSurfacesInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidatePagination(input.Page, input.PageSize)...)

	return errors
}

func (input *UpdateProjectSurfaceInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectSurfaceID, "project_surface_id")...)

	if input.Name == nil && input.Scale == nil {
		errors = append(errors, apperror.NewFieldError(
			"project_surface",
			apperror.FieldCodeRequired,
			"at least one project surface field must be provided"),
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

func (input *DeleteProjectSurfaceInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateID(input.UserID, "user_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectID, "project_id")...)
	errors = append(errors, validation.ValidateID(input.ProjectSurfaceID, "project_surface_id")...)

	return errors
}
