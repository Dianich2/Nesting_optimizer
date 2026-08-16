package http

import (
	"server_nesting_optimizer/internal/app/container"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/handler"
	"server_nesting_optimizer/internal/transport/http/middleware"

	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
)

func NewRouter(deps *container.Container) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return httperror.Handle(c, err)
		},
	})

	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(
		deps.CreateUserUseCase,
		deps.LoginUseCase,
		deps.RefreshUseCase,
		deps.LogoutUseCase,
		deps.GetCurrentUserUseCase,
		deps.UpdateProfileUseCase,
		deps.ChangePasswordUseCase,
		deps.DeleteCurrentUserUseCase,
	)

	projectHandler := handler.NewProjectHandler(
		deps.CreateProjectUseCase,
		deps.GetProjectByIDUseCase,
		deps.ListProjectsUseCase,
		deps.UpdateProjectUseCase,
		deps.DeleteProjectUseCase,
	)

	surfaceHandler := handler.NewSurfaceHandler(
		deps.CreateSurfaceUseCase,
	)

	app.Get("/health", healthHandler.Check)
	app.Get("/swagger/*", swaggo.HandlerDefault)

	api := app.Group("/api/v1")
	api.Post("/users", userHandler.Create)
	api.Post("/auth/login", userHandler.Login)
	api.Post("/auth/refresh", userHandler.Refresh)
	api.Post("/auth/logout", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), userHandler.Logout)
	api.Get("/users/me", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), userHandler.GetCurrentUser)
	api.Patch("/users/me", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), userHandler.UpdateProfile)
	api.Patch("/users/me/password", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), userHandler.ChangePassword)
	api.Post("/users/me/delete", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), userHandler.DeleteCurrentUser)

	api.Post("/projects", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectHandler.Create)
	api.Get("/projects/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectHandler.GetByID)
	api.Get("/projects", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectHandler.ListProjects)
	api.Patch("/projects/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectHandler.Update)
	api.Delete("/projects/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectHandler.DeleteProject)

	api.Post("/surfaces", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), surfaceHandler.Create)

	return app
}
