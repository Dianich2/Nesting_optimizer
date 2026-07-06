package httperror

import (
	"server_nesting_optimizer/pkg/apperror"

	"github.com/gofiber/fiber/v3"
)

func Handle(
	c fiber.Ctx,
	err error,
) error {
	appError, ok := apperror.As(err)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(NewErrorResponse(
			string(apperror.AppCodeInternal),
			"internal server error",
			nil,
		))
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

	return c.Status(status).JSON(NewErrorResponse(
		string(appError.Code),
		appError.Message,
		appError.Details,
	))
}
