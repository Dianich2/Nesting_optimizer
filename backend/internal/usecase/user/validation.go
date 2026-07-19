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

func validateCreateUserInput(
	uc CreateUserInput,
) []apperror.FieldError {
	errors := validateRequiredFields(uc)

	errors = append(errors, validateLoginLen(uc.Login)...)
	errors = append(errors, validateEmailLen(uc.Email)...)
	errors = append(errors, validateFirstNameLen(uc.FirstName)...)
	errors = append(errors, validateLastNameLen(uc.LastName)...)

	errors = append(errors, validateEmail(uc.Email)...)
	errors = append(errors, validateName("first_name", uc.FirstName)...)
	errors = append(errors, validateName("last_name", uc.LastName)...)
	errors = append(errors, validateLogin(uc.Login)...)
	errors = append(errors, validatePassword(uc.Password)...)

	return errors
}

func validateRequiredFields(
	uc CreateUserInput,
) []apperror.FieldError {
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
	password string,
) []apperror.FieldError {
	if strings.TrimSpace(password) == "" {
		return nil
	}
	if utf8.RuneCountInString(password) < 8 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"password",
				apperror.FieldCodeTooShort,
				"password is too short",
			),
		}
	}

	if len(password) > 72 {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"password",
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
