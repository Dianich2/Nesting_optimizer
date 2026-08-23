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
		deps.GetSurfaceByIDUseCase,
		deps.ListSurfacesUseCase,
		deps.UpdateSurfaceUseCase,
		deps.DeleteSurfaceUseCase,
	)

	projectSurfaceHandler := handler.NewProjectSurfaceHandler(
		deps.CreateProjectSurfaceUseCase,
		deps.GetProjectSurfaceByIDUseCase,
		deps.ListProjectSurfacesUseCase,
		deps.UpdateProjectSurfaceUseCase,
		deps.DeleteProjectSurfaceUseCase,
	)

	patternHandler := handler.NewPatternHandler(
		deps.CreatePatternUseCase,
		deps.GetPatternByIDUseCase,
		deps.ListPatternsUseCase,
		deps.UpdatePatternUseCase,
		deps.DeletePatternUseCase,
	)

	projectPatternHandler := handler.NewProjectPatternHandler(
		deps.CreateProjectPatternUseCase,
		deps.GetProjectPatternByIDUseCase,
		deps.ListProjectPatternsUseCase,
		deps.UpdateProjectPatternUseCase,
		deps.DeleteProjectPatternUseCase,
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
	api.Get("/surfaces/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), surfaceHandler.GetByID)
	api.Get("/surfaces", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), surfaceHandler.ListSurfaces)
	api.Patch("/surfaces/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), surfaceHandler.Update)
	api.Delete("/surfaces/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), surfaceHandler.Delete)

	api.Post("/projects/:project_id/surfaces", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectSurfaceHandler.Create)
	api.Get("/projects/:project_id/surfaces/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectSurfaceHandler.GetByID)
	api.Get("/projects/:project_id/surfaces", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectSurfaceHandler.ListProjectSurfaces)
	api.Patch("/projects/:project_id/surfaces/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectSurfaceHandler.Update)
	api.Delete("/projects/:project_id/surfaces/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectSurfaceHandler.Delete)

	api.Post("/patterns", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), patternHandler.Create)
	api.Get("/patterns/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), patternHandler.GetByID)
	api.Get("/patterns", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), patternHandler.ListPatterns)
	api.Patch("/patterns/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), patternHandler.Update)
	api.Delete("/patterns/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), patternHandler.Delete)

	api.Post("/projects/:project_id/patterns", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectPatternHandler.Create)
	api.Get("/projects/:project_id/patterns/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectPatternHandler.GetByID)
	api.Get("/projects/:project_id/patterns", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectPatternHandler.ListProjectPatterns)
	api.Patch("/projects/:project_id/patterns/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectPatternHandler.Update)
	api.Delete("/projects/:project_id/patterns/:id", middleware.AuthRequired(deps.JWTManager, deps.SessionRepository), projectPatternHandler.Delete)

	return app
}
