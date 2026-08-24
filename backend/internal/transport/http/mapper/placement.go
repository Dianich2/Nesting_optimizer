package mapper

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	placementusecase "server_nesting_optimizer/internal/usecase/placement"
)

func ToCreatePlacementInput(
	req dto.CreatePlacementRequest,
	userID int64,
	projectID int64,
	projectSurfaceID int64,
) placementusecase.CreatePlacementInput {
	return placementusecase.CreatePlacementInput{
		UserID:           userID,
		ProjectID:        projectID,
		ProjectPatternID: req.ProjectPatternID,
		ProjectSurfaceID: projectSurfaceID,
		X:                req.X,
		Y:                req.Y,
		Rotation:         req.Rotation,
	}
}

func ToCreatePlacementResponse(
	resp placementusecase.CreatePlacementOutput,
) dto.CreatePlacementResponse {
	return dto.CreatePlacementResponse{
		ID:               resp.ID,
		ProjectPatternID: resp.ProjectPatternID,
		ProjectSurfaceID: resp.ProjectSurfaceID,
		X:                resp.X,
		Y:                resp.Y,
		Rotation:         resp.Rotation,
		CreatedAt:        resp.CreatedAt,
		UpdatedAt:        resp.UpdatedAt,
	}
}
