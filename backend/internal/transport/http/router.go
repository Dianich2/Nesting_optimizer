package http

import (
	"server_nesting_optimizer/internal/transport/http/handler"

	"github.com/gofiber/fiber/v3"
)

func NewRouter() *fiber.App {
	app := fiber.New()

	healthHandler := handler.NewHealthHandler()

	app.Get("/health", healthHandler.Check)
	return app
}
