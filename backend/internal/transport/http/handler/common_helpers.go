package handler

import (
	"errors"
	"server_nesting_optimizer/internal/config"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/middleware"
	"server_nesting_optimizer/pkg/apperror"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func parseBody(context fiber.Ctx, body any) error {
	err := context.Bind().Body(body)
	if err != nil {
		return httperror.Handle(
			context,
			apperror.Validation("invalid request body"),
		)
	}

	return nil
}

func getUserID(context fiber.Ctx) (int64, error) {
	userID := context.Locals(middleware.UserIDLocalKey)
	u, ok := userID.(int64)
	if !ok {
		return 0, httperror.Handle(
			context,
			apperror.Internal("failed to get user id from request context", errors.New("user_id local is missing or invalid")),
		)
	}

	return u, nil
}

func getIDFromPath(context fiber.Ctx, field string) (int64, error) {
	id, err := strconv.ParseInt(context.Params(field), 10, 64)
	if err != nil {
		return 0, httperror.Handle(
			context,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					field,
					apperror.FieldCodeInvalid,
					field+" must be a valid positive integer",
				),
			),
		)
	}

	return id, nil
}

func parsePagination(
	context fiber.Ctx,
) (int, int, error) {
	pageSizeQuery, pageQuery := context.Query("page_size"), context.Query("page")

	if pageSizeQuery == "" {
		pageSizeQuery = config.DefaultPageSize
	}

	if pageQuery == "" {
		pageQuery = config.DefaultPage
	}

	pageSize, err := strconv.Atoi(pageSizeQuery)
	if err != nil {
		return 0, 0, httperror.Handle(
			context,
			apperror.Validation(
				"invalid page size",
				apperror.NewFieldError(
					"page_size",
					apperror.FieldCodeInvalid,
					"page size must be a valid positive integer",
				),
			),
		)
	}

	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		return 0, 0, httperror.Handle(
			context,
			apperror.Validation(
				"invalid page",
				apperror.NewFieldError(
					"page",
					apperror.FieldCodeInvalid,
					"page must be a valid positive integer",
				),
			),
		)
	}

	return page, pageSize, nil
}
