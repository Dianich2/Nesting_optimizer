package user

import (
	"regexp"
	"server_nesting_optimizer/pkg/apperror"
	"strings"
	"unicode/utf8"
)

var (
	regexValidateEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	regexValidateLogin = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]*[a-z0-9])?$`)
	regexValidateName  = regexp.MustCompile(`^\p{L}+([-'’ ]\p{L}+)*$`)
)

func (uc *CreateUserInput) Validate() []apperror.FieldError {
	errors := uc.validateRequiredFields()

	errors = append(errors, validateLoginLen(uc.Login)...)
	errors = append(errors, validateEmailLen(uc.Email)...)
	errors = append(errors, validateFirstNameLen(uc.FirstName)...)
	errors = append(errors, validateLastNameLen(uc.LastName)...)

	errors = append(errors, validateEmail(uc.Email)...)
	errors = append(errors, validateName("first_name", uc.FirstName)...)
	errors = append(errors, validateName("last_name", uc.LastName)...)
	errors = append(errors, validateLogin(uc.Login)...)
	errors = append(errors, validatePassword("password", uc.Password)...)

	return errors
}

func (uc *CreateUserInput) validateRequiredFields() []apperror.FieldError {
	var errors []apperror.FieldError

	if uc.Login == "" {
		errors = append(errors, apperror.NewFieldError(
			"login",
			apperror.FieldCodeRequired,
			"login must not be empty"),
		)
	}

	if uc.Email == "" {
		errors = append(errors, apperror.NewFieldError(
			"email",
			apperror.FieldCodeRequired,
			"email must not be empty"),
		)
	}

	if strings.TrimSpace(uc.Password) == "" {
		errors = append(errors, apperror.NewFieldError(
			"password",
			apperror.FieldCodeRequired,
			"password must not be empty"),
		)
	}

	if uc.FirstName == "" {
		errors = append(errors, apperror.NewFieldError(
			"first_name",
			apperror.FieldCodeRequired,
			"first_name must not be empty"),
		)
	}

	if uc.LastName == "" {
		errors = append(errors, apperror.NewFieldError(
			"last_name",
			apperror.FieldCodeRequired,
			"last_name must not be empty"),
		)
	}

	return errors
}

func validateLoginLen(
	login string,
) []apperror.FieldError {
	if utf8.RuneCountInString(login) > 100 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"login",
				apperror.FieldCodeTooLong,
				"login is too long",
			),
		}
	}

	return nil
}

func validateEmailLen(
	email string,
) []apperror.FieldError {
	if len(email) > 254 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"email",
				apperror.FieldCodeTooLong,
				"email is too long",
			),
		}
	}

	return nil
}

func validateFirstNameLen(
	firstName string,
) []apperror.FieldError {
	if utf8.RuneCountInString(firstName) > 50 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"first_name",
				apperror.FieldCodeTooLong,
				"first_name is too long",
			),
		}
	}

	return nil
}

func validateLastNameLen(
	lastName string,
) []apperror.FieldError {
	if utf8.RuneCountInString(lastName) > 50 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"last_name",
				apperror.FieldCodeTooLong,
				"last_name is too long",
			),
		}
	}

	return nil
}

func validateEmail(
	email string,
) []apperror.FieldError {
	if email == "" {
		return nil
	}

	if !regexValidateEmail.MatchString(email) {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"email",
				apperror.FieldCodeInvalid,
				"invalid email",
			),
		}
	}

	return nil
}

func validateLogin(
	login string,
) []apperror.FieldError {
	if login == "" {
		return nil
	}
	if !regexValidateLogin.MatchString(login) {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"login",
				apperror.FieldCodeInvalid,
				"invalid login",
			),
		}
	}

	return nil
}

func validatePassword(
	field string,
	value string,
) []apperror.FieldError {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	if utf8.RuneCountInString(value) < 8 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				field,
				apperror.FieldCodeTooShort,
				"password is too short",
			),
		}
	}

	if len(value) > 72 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				field,
				apperror.FieldCodeTooLong,
				"password is too long",
			),
		}
	}

	return nil
}

func validateName(
	field string,
	value string,
) []apperror.FieldError {
	if value == "" {
		return nil
	}
	if !regexValidateName.MatchString(value) {
		return []apperror.FieldError{
			apperror.NewFieldError(
				field,
				apperror.FieldCodeInvalid,
				"invalid "+field,
			),
		}
	}

	return nil
}

func validateLoginInput(
	l LoginInput,
) []apperror.FieldError {
	var errors []apperror.FieldError

	if l.Identifier == "" {
		errors = append(errors, apperror.NewFieldError(
			"identifier",
			apperror.FieldCodeRequired,
			"identifier must not be empty"),
		)
	}

	if strings.TrimSpace(l.Password) == "" {
		errors = append(errors, apperror.NewFieldError(
			"password",
			apperror.FieldCodeRequired,
			"password must not be empty"),
		)
	}

	return errors
}

func (uc *UpdateProfileInput) Validate() []apperror.FieldError {
	errors := uc.validateRequiredFields()

	if uc.FirstName != nil {
		errors = append(errors, validateFirstNameLen(*uc.FirstName)...)
		errors = append(errors, validateName("first_name", *uc.FirstName)...)
	}

	if uc.LastName != nil {
		errors = append(errors, validateLastNameLen(*uc.LastName)...)
		errors = append(errors, validateName("last_name", *uc.LastName)...)
	}

	return errors
}

func (uc *UpdateProfileInput) validateRequiredFields() []apperror.FieldError {
	var details []apperror.FieldError

	if uc.FirstName == nil && uc.LastName == nil {
		return []apperror.FieldError{apperror.NewFieldError(
			"profile",
			apperror.FieldCodeRequired,
			"at least one profile field must be provided",
		)}
	}

	if uc.FirstName != nil && *uc.FirstName == "" {
		details = append(details, apperror.NewFieldError(
			"first_name",
			apperror.FieldCodeRequired,
			"first_name must not be empty",
		))
	}

	if uc.LastName != nil && *uc.LastName == "" {
		details = append(details, apperror.NewFieldError(
			"last_name",
			apperror.FieldCodeRequired,
			"last_name must not be empty",
		))
	}

	return details
}

func (uc *ChangePasswordInput) Validate() []apperror.FieldError {
	errors := uc.validateRequiredFields()
	errors = append(errors, validatePassword("new_password", uc.NewPassword)...)

	if (strings.TrimSpace(uc.RepeatNewPassword) != "" && strings.TrimSpace(uc.NewPassword) != "") && uc.NewPassword != uc.RepeatNewPassword {
		errors = append(errors, apperror.NewFieldError(
			"new_password",
			apperror.FieldCodeInvalid,
			"repeat new password and new password not equal"),
		)
	}

	return errors
}

func (uc *ChangePasswordInput) validateRequiredFields() []apperror.FieldError {
	var errors []apperror.FieldError

	if strings.TrimSpace(uc.OldPassword) == "" {
		errors = append(errors, apperror.NewFieldError(
			"old_password",
			apperror.FieldCodeRequired,
			"old password must not be empty"),
		)
	}

	if strings.TrimSpace(uc.NewPassword) == "" {
		errors = append(errors, apperror.NewFieldError(
			"new_password",
			apperror.FieldCodeRequired,
			"new password must not be empty"),
		)
	}

	if strings.TrimSpace(uc.RepeatNewPassword) == "" {
		errors = append(errors, apperror.NewFieldError(
			"repeat_new_password",
			apperror.FieldCodeRequired,
			"repeat new password must not be empty"),
		)
	}

	return errors
}
