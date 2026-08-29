package http

import (
	"server_nesting_optimizer/internal/transport/http/handler"

	"github.com/gofiber/fiber/v3"
)

func registerPublicRoutes(
	api fiber.Router,
	userHandler *handler.UserHandler,
) {
	api.Post("/users", userHandler.Create)
	api.Post("/auth/login", userHandler.Login)
	api.Post("/auth/refresh", userHandler.Refresh)
}

func registerUserRoutes(
	api fiber.Router,
	userHandler *handler.UserHandler,
) {
	api.Post("/auth/logout", userHandler.Logout)

	api.Get("/users/me", userHandler.GetCurrentUser)
	api.Patch("/users/me", userHandler.UpdateProfile)
	api.Patch("/users/me/password", userHandler.ChangePassword)
	api.Post("/users/me/delete", userHandler.DeleteCurrentUser)
}

func registerProjectRoutes(
	api fiber.Router,
	projectHandler *handler.ProjectHandler,
) {
	api.Post("/projects", projectHandler.Create)
	api.Get("/projects/:id", projectHandler.GetByID)
	api.Get("/projects", projectHandler.ListProjects)
	api.Patch("/projects/:id", projectHandler.Update)
	api.Delete("/projects/:id", projectHandler.DeleteProject)
}

func registerSurfaceRoutes(
	api fiber.Router,
	surfaceHandler *handler.SurfaceHandler,
) {
	api.Post("/surfaces", surfaceHandler.Create)
	api.Get("/surfaces/:id", surfaceHandler.GetByID)
	api.Get("/surfaces", surfaceHandler.ListSurfaces)
	api.Patch("/surfaces/:id", surfaceHandler.Update)
	api.Delete("/surfaces/:id", surfaceHandler.Delete)
}

func registerProjectSurfaceRoutes(
	api fiber.Router,
	projectSurfaceHandler *handler.ProjectSurfaceHandler,
) {
	api.Post("/projects/:project_id/surfaces", projectSurfaceHandler.Create)
	api.Get("/projects/:project_id/surfaces/:id", projectSurfaceHandler.GetByID)
	api.Get("/projects/:project_id/surfaces", projectSurfaceHandler.ListProjectSurfaces)
	api.Patch("/projects/:project_id/surfaces/:id", projectSurfaceHandler.Update)
	api.Delete("/projects/:project_id/surfaces/:id", projectSurfaceHandler.Delete)
}

func registerPatternRoutes(
	api fiber.Router,
	patternHandler *handler.PatternHandler,
) {
	api.Post("/patterns", patternHandler.Create)
	api.Get("/patterns/:id", patternHandler.GetByID)
	api.Get("/patterns", patternHandler.ListPatterns)
	api.Patch("/patterns/:id", patternHandler.Update)
	api.Delete("/patterns/:id", patternHandler.Delete)
}

func registerProjectPatternRoutes(
	api fiber.Router,
	projectPatternHandler *handler.ProjectPatternHandler,
) {
	api.Post("/projects/:project_id/patterns", projectPatternHandler.Create)
	api.Get("/projects/:project_id/patterns/:id", projectPatternHandler.GetByID)
	api.Get("/projects/:project_id/patterns", projectPatternHandler.ListProjectPatterns)
	api.Patch("/projects/:project_id/patterns/:id", projectPatternHandler.Update)
	api.Delete("/projects/:project_id/patterns/:id", projectPatternHandler.Delete)
}

func registerPlacementRoutes(
	api fiber.Router,
	placementHandler *handler.PlacementHandler,
) {
	api.Post("/projects/:project_id/surfaces/:project_surface_id/placements", placementHandler.Create)
	api.Get("/projects/:project_id/placements/:placement_id", placementHandler.GetByID)
	api.Get("/projects/:project_id/surfaces/:project_surface_id/placements", placementHandler.ListPlacements)
	api.Patch("/projects/:project_id/placements/:placement_id", placementHandler.Update)
	api.Delete("/projects/:project_id/placements/:placement_id", placementHandler.Delete)
}

func registerNestingRoutes(
	api fiber.Router,
	nestingHandler *handler.NestingHandler,
) {
	api.Post("/projects/:project_id/surfaces/:project_surface_id/nesting", nestingHandler.Run)
}
