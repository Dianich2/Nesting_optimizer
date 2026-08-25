package user

import (
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
	"strings"
)

func (uc *CreateUserInput) Validate() []apperror.FieldError {
	errors := uc.validateRequiredFields()

	errors = append(errors, validation.ValidateMaxLen("login", uc.Login, config.MaxLoginLen)...)
	errors = append(errors, validation.ValidateMaxLen("email", uc.Email, config.MaxEmailLen)...)
	errors = append(errors, validation.ValidateMaxLen("first_name", uc.FirstName, config.MaxUserNamesLen)...)
	errors = append(errors, validation.ValidateMaxLen("last_name", uc.LastName, config.MaxUserNamesLen)...)

	errors = append(errors, validation.ValidateEmail(uc.Email)...)
	errors = append(errors, validation.ValidateName("first_name", uc.FirstName)...)
	errors = append(errors, validation.ValidateName("last_name", uc.LastName)...)
	errors = append(errors, validation.ValidateLogin(uc.Login)...)
	errors = append(errors, validation.ValidatePassword("password", uc.Password)...)

	return errors
}

func validateLoginInput(
	l LoginInput,
) []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateFieldForEmptiness("identifier", l.Identifier)...)
	errors = append(errors, validation.ValidateFieldForEmptiness("password", strings.TrimSpace(l.Password))...)

	return errors
}

func (uc *CreateUserInput) validateRequiredFields() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateFieldForEmptiness("login", uc.Login)...)
	errors = append(errors, validation.ValidateFieldForEmptiness("email", uc.Email)...)
	errors = append(errors, validation.ValidateFieldForEmptiness("password", strings.TrimSpace(uc.Password))...)
	errors = append(errors, validation.ValidateFieldForEmptiness("first_name", uc.FirstName)...)
	errors = append(errors, validation.ValidateFieldForEmptiness("last_name", uc.LastName)...)

	return errors
}

func (uc *UpdateProfileInput) Validate() []apperror.FieldError {
	errors := uc.validateRequiredFields()

	if uc.FirstName != nil {
		errors = append(errors, validation.ValidateMaxLen("first_name", *uc.FirstName, config.MaxUserNamesLen)...)
		errors = append(errors, validation.ValidateName("first_name", *uc.FirstName)...)
	}

	if uc.LastName != nil {
		errors = append(errors, validation.ValidateMaxLen("last_name", *uc.LastName, config.MaxUserNamesLen)...)
		errors = append(errors, validation.ValidateName("last_name", *uc.LastName)...)
	}

	return errors
}

func (uc *UpdateProfileInput) validateRequiredFields() []apperror.FieldError {
	var errors []apperror.FieldError

	if uc.FirstName == nil && uc.LastName == nil {
		return []apperror.FieldError{apperror.NewFieldError(
			"profile",
			apperror.FieldCodeRequired,
			"at least one profile field must be provided",
		)}
	}

	if uc.FirstName != nil {
		errors = append(errors, validation.ValidateFieldForEmptiness("first_name", *uc.FirstName)...)
	}

	if uc.LastName != nil {
		errors = append(errors, validation.ValidateFieldForEmptiness("last_name", *uc.LastName)...)
	}

	return errors
}

func (uc *ChangePasswordInput) Validate() []apperror.FieldError {
	errors := uc.validateRequiredFields()
	errors = append(errors, validation.ValidatePassword("new_password", uc.NewPassword)...)

	if (strings.TrimSpace(uc.RepeatNewPassword) != "" && strings.TrimSpace(uc.NewPassword) != "") && uc.NewPassword != uc.RepeatNewPassword {
		errors = append(errors, apperror.NewFieldError(
			"repeat_new_password",
			apperror.FieldCodeInvalid,
			"passwords do not match"),
		)
	}

	return errors
}

func (uc *ChangePasswordInput) validateRequiredFields() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateFieldForEmptiness("old_password", strings.TrimSpace(uc.OldPassword))...)
	errors = append(errors, validation.ValidateFieldForEmptiness("new_password", strings.TrimSpace(uc.NewPassword))...)
	errors = append(errors, validation.ValidateFieldForEmptiness("repeat_new_password", strings.TrimSpace(uc.RepeatNewPassword))...)

	return errors
}

func (uc *DeleteCurrentUserInput) Validate() []apperror.FieldError {
	var errors []apperror.FieldError

	errors = append(errors, validation.ValidateFieldForEmptiness("password", strings.TrimSpace(uc.Password))...)

	return errors
}
