package projectsurface

import (
	"server_nesting_optimizer/pkg/apperror"
	"unicode/utf8"
)

func (input *CreateProjectSurfaceInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validateID(input.UserID, "user_id")...)
	errors = append(errors, validateID(input.ProjectID, "project_id")...)
	errors = append(errors, validateID(input.SourceSurfaceID, "source_surface_id")...)

	errors = append(errors, validateScale(input.Scale)...)

	return errors
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

func validateScale(
	scale float64,
) []apperror.FieldError {
	if scale <= 0 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"scale",
				apperror.FieldCodeInvalid,
				"scale must be greater than 0",
			),
		}
	}

	return nil
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

func (input *GetProjectSurfaceByIDInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validateID(input.UserID, "user_id")...)
	errors = append(errors, validateID(input.ProjectID, "project_id")...)
	errors = append(errors, validateID(input.ProjectSurfaceID, "project_surface_id")...)

	return errors
}

func (input *ListProjectSurfacesInput) Validate() []apperror.FieldError {
	errors := []apperror.FieldError{}

	errors = append(errors, validateID(input.UserID, "user_id")...)
	errors = append(errors, validateID(input.ProjectID, "project_id")...)

	if input.Page < 1 {
		errors = append(errors,
			apperror.NewFieldError(
				"page",
				apperror.FieldCodeInvalid,
				"page must be greater than 0",
			),
		)
	}

	if input.PageSize < 1 || input.PageSize > 100 {
		errors = append(errors,
			apperror.NewFieldError(
				"page_size",
				apperror.FieldCodeInvalid,
				"page_size must be between 1 and 100",
			),
		)
	}

	return errors
}

func (input *UpdateProjectSurfaceInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validateID(input.UserID, "user_id")...)
	errors = append(errors, validateID(input.ProjectID, "project_id")...)
	errors = append(errors, validateID(input.ProjectSurfaceID, "project_surface_id")...)

	if input.Name == nil && input.Scale == nil {
		errors = append(errors, apperror.NewFieldError(
			"project_surface",
			apperror.FieldCodeRequired,
			"at least one project surface field must be provided"),
		)
	}

	if input.Name != nil {
		if *input.Name == "" {
			errors = append(
				errors,
				apperror.NewFieldError(
					"name",
					apperror.FieldCodeRequired,
					"name must not be empty",
				),
			)
		}

		errors = append(errors, validateNameLen(*input.Name)...)
	}

	if input.Scale != nil {
		errors = append(errors, validateScale(*input.Scale)...)
	}

	return errors
}

func (input *DeleteProjectSurfaceInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validateID(input.UserID, "user_id")...)
	errors = append(errors, validateID(input.ProjectID, "project_id")...)
	errors = append(errors, validateID(input.ProjectSurfaceID, "project_surface_id")...)

	return errors
}
