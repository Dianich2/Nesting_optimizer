package apperror

import "errors"

type Code string

type FieldCode string

const (
	AppCodeValidation   Code = "validation_error"
	AppCodeNotFound     Code = "not_found_error"
	AppCodeConflict     Code = "conflict_error"
	AppCodeUnauthorized Code = "unauthorized_error"
	AppCodeForbidden    Code = "forbidden_error"
	AppCodeInternal     Code = "internal_error"
)

const (
	FieldCodeRequired      FieldCode = "required"
	FieldCodeInvalid       FieldCode = "invalid"
	FieldCodeTooLong       FieldCode = "too_long"
	FieldCodeTooShort      FieldCode = "too_short"
	FieldCodeAlreadyExists FieldCode = "already_exists"
)

type FieldError struct {
	Field   string    `json:"field"`
	Code    FieldCode `json:"code"`
	Message string    `json:"message"`
}

func NewFieldError(
	field string,
	code FieldCode,
	message string,
) FieldError {
	return FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	}
}

type AppError struct {
	Code    Code
	Message string
	Details []FieldError
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func Validation(
	message string,
	details ...FieldError,
) *AppError {
	return &AppError{
		Code:    AppCodeValidation,
		Message: message,
		Details: details,
	}
}

func Conflict(
	message string,
) *AppError {
	return &AppError{
		Code:    AppCodeConflict,
		Message: message,
	}
}

func NotFound(
	message string,
) *AppError {
	return &AppError{
		Code:    AppCodeNotFound,
		Message: message,
	}
}

func Unauthorized(
	message string,
) *AppError {
	return &AppError{
		Code:    AppCodeUnauthorized,
		Message: message,
	}
}

func Forbidden(
	message string,
) *AppError {
	return &AppError{
		Code:    AppCodeForbidden,
		Message: message,
	}
}

func Internal(
	message string,
	err error,
) *AppError {
	return &AppError{
		Code:    AppCodeInternal,
		Message: message,
		Err:     err,
	}
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}

	return nil, false
}
