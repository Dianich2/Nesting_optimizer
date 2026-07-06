package user

import (
	"net/mail"
	"server_nesting_optimizer/pkg/apperror"
	"strings"
)

func validateCreateUserInput(uc CreateUserInput) []apperror.FieldError {
	errors := validateRequiredFields(uc)

	errors = append(errors, validateLoginLen(uc.Login)...)
	errors = append(errors, validateEmailLen(uc.Email)...)
	errors = append(errors, validateFirstNameLen(uc.FirstName)...)
	errors = append(errors, validateLastNameLen(uc.LastName)...)

	errors = append(errors, validateEmail(uc.Email)...)
	return errors
}

func validateRequiredFields(uc CreateUserInput) []apperror.FieldError {
	var errors []apperror.FieldError
	if strings.TrimSpace(uc.Login) == "" {
		errors = append(errors, apperror.NewFieldError("login", apperror.FieldCodeRequired, "login must not be empty"))
	}
	if strings.TrimSpace(uc.Email) == "" {
		errors = append(errors, apperror.NewFieldError("email", apperror.FieldCodeRequired, "email must not be empty"))
	}
	if strings.TrimSpace(uc.Password) == "" {
		errors = append(errors, apperror.NewFieldError("password", apperror.FieldCodeRequired, "password must not be empty"))
	}
	if strings.TrimSpace(uc.FirstName) == "" {
		errors = append(errors, apperror.NewFieldError("first_name", apperror.FieldCodeRequired, "first_name must not be empty"))
	}
	if strings.TrimSpace(uc.LastName) == "" {
		errors = append(errors, apperror.NewFieldError("last_name", apperror.FieldCodeRequired, "last_name must not be empty"))
	}
	return errors
}

func validateLoginLen(login string) []apperror.FieldError {
	if len(login) > 100 {
		return []apperror.FieldError{
			apperror.NewFieldError("login", apperror.FieldCodeTooLong, "login is too long"),
		}
	}
	return nil
}

func validateEmailLen(email string) []apperror.FieldError {
	if len(email) > 254 {
		return []apperror.FieldError{
			apperror.NewFieldError("email", apperror.FieldCodeTooLong, "email is too long"),
		}
	}
	return nil
}

func validateFirstNameLen(firstName string) []apperror.FieldError {
	if len(firstName) > 50 {
		return []apperror.FieldError{
			apperror.NewFieldError("first_name", apperror.FieldCodeTooLong, "first_name is too long"),
		}
	}
	return nil
}

func validateLastNameLen(lastName string) []apperror.FieldError {
	if len(lastName) > 50 {
		return []apperror.FieldError{
			apperror.NewFieldError("last_name", apperror.FieldCodeTooLong, "last_name is too long"),
		}
	}
	return nil
}

func validateEmail(email string) []apperror.FieldError {
	if email == "" {
		return nil
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return []apperror.FieldError{
			apperror.NewFieldError("email", apperror.FieldCodeInvalid, "invalid email"),
		}
	}

	return nil
}
