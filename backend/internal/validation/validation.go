package validation

import (
	"math"
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/pkg/apperror"
	"strings"
	"unicode/utf8"
)

func ValidateID(
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

func ValidateCoordinate(
	field string,
	value float64,
) []apperror.FieldError {
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return []apperror.FieldError{
			apperror.NewFieldError(
				field,
				apperror.FieldCodeInvalid,
				"coordinate must not be NaN or Inf",
			),
		}
	}

	return nil
}

func ValidateRotation(
	rotation float64,
) []apperror.FieldError {
	if math.IsInf(rotation, 0) || math.IsNaN(rotation) {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"rotation",
				apperror.FieldCodeInvalid,
				"rotation must not be NaN or Inf",
			),
		}
	}

	return nil
}

func ValidateMaxLen(
	field string,
	value string,
	maxLen int,
) []apperror.FieldError {
	if utf8.RuneCountInString(value) > maxLen {
		return []apperror.FieldError{
			apperror.NewFieldError(
				field,
				apperror.FieldCodeTooLong,
				field+" is too long",
			),
		}
	}

	return nil
}

func ValidatePagination(
	page int,
	pageSize int,
) []apperror.FieldError {
	var errors []apperror.FieldError
	if page < config.MinPage {
		errors = append(errors,
			apperror.NewFieldError(
				"page",
				apperror.FieldCodeInvalid,
				"page must be greater than 0",
			),
		)
	}

	if pageSize < config.MinPageSize || pageSize > config.MaxPageSize {
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

func ValidateFieldForEmptiness(
	field string,
	value string,
) []apperror.FieldError {
	var errors []apperror.FieldError
	if value == "" {
		errors = append(
			errors,
			apperror.NewFieldError(
				field,
				apperror.FieldCodeRequired,
				field+" must not be empty",
			),
		)
	}

	return errors
}

func ValidateScale(
	scale float64,
) []apperror.FieldError {
	if math.IsNaN(scale) || math.IsInf(scale, 0) {
		return []apperror.FieldError{
			apperror.NewFieldError(
				"scale",
				apperror.FieldCodeInvalid,
				"scale must not be NaN or Inf",
			),
		}
	}

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

func ValidateEmail(
	email string,
) []apperror.FieldError {
	if email == "" {
		return nil
	}

	if !config.RegexValidateEmail.MatchString(email) {
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

func ValidateLogin(
	login string,
) []apperror.FieldError {
	if login == "" {
		return nil
	}
	if !config.RegexValidateLogin.MatchString(login) {
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

func ValidatePassword(
	field string,
	value string,
) []apperror.FieldError {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	if utf8.RuneCountInString(value) < config.MinPasswordLen {
		return []apperror.FieldError{
			apperror.NewFieldError(
				field,
				apperror.FieldCodeTooShort,
				"password is too short",
			),
		}
	}

	if len(value) > config.MaxPasswordLen {
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

func ValidateName(
	field string,
	value string,
) []apperror.FieldError {
	if value == "" {
		return nil
	}
	if !config.RegexValidateName.MatchString(value) {
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
