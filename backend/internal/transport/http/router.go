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

	placementHandler := handler.NewPlacementHandler(
		deps.CreatePlacementUseCase,
		deps.GetPlacementByIDUseCase,
		deps.ListPlacementsUseCase,
		deps.UpdatePlacementUseCase,
		deps.DeletePlacementUseCase,
	)

	app.Get("/health", healthHandler.Check)
	app.Get("/swagger/*", swaggo.HandlerDefault)

	api := app.Group("/api/v1")
	protected := api.Group(
		"",
		middleware.AuthRequired(
			deps.JWTManager,
			deps.SessionRepository,
		),
	)

	registerPublicRoutes(api, userHandler)
	registerUserRoutes(protected, userHandler)
	registerProjectRoutes(protected, projectHandler)
	registerSurfaceRoutes(protected, surfaceHandler)
	registerProjectSurfaceRoutes(protected, projectSurfaceHandler)
	registerPatternRoutes(protected, patternHandler)
	registerProjectPatternRoutes(protected, projectPatternHandler)
	registerPlacementRoutes(protected, placementHandler)

	return app
}
