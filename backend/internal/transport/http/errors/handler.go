package httperror

import (
	"server_nesting_optimizer/pkg/apperror"

	"github.com/gofiber/fiber/v3"
)

func writeInternalError(
	c fiber.Ctx,
) error {
	return c.Status(fiber.StatusInternalServerError).JSON(NewErrorResponse(
		string(apperror.AppCodeInternal),
		"internal server error",
		nil,
	))
}

func Handle(
	c fiber.Ctx,
	err error,
) error {
	appError, ok := apperror.As(err)
	if !ok {
		return writeInternalError(c)
	}

	if appError.Code == apperror.AppCodeInternal {
		return writeInternalError(c)
	}

	status := fiber.StatusInternalServerError

	switch appError.Code {
	case apperror.AppCodeValidation:
		status = fiber.StatusBadRequest
	case apperror.AppCodeUnauthorized:
		status = fiber.StatusUnauthorized
	case apperror.AppCodeForbidden:
		status = fiber.StatusForbidden
	case apperror.AppCodeNotFound:
		status = fiber.StatusNotFound
	case apperror.AppCodeConflict:
		status = fiber.StatusConflict
	}

	if status == fiber.StatusInternalServerError {
		return writeInternalError(c)
	}

	return c.Status(status).JSON(NewErrorResponse(
		string(appError.Code),
		appError.Message,
		appError.Details,
	))
}
