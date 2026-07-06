package http

import (
	"server_nesting_optimizer/internal/app/container"
	"server_nesting_optimizer/internal/transport/http/handler"

	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
)

func NewRouter(deps *container.Container) *fiber.App {
	app := fiber.New()

	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(deps.CreateUserUseCase)

	app.Get("/health", healthHandler.Check)
	app.Get("/swagger/*", swaggo.HandlerDefault)

	api := app.Group("/api/v1")
	api.Post("/users", userHandler.Create)
	return app
}
