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

func ToGetPlacementByIDInput(
	userID int64,
	projectID int64,
	placementID int64,
) placementusecase.GetPlacementByIDInput {
	return placementusecase.GetPlacementByIDInput{
		UserID:      userID,
		ProjectID:   projectID,
		PlacementID: placementID,
	}
}

func ToGetPlacementByIDResponse(
	resp placementusecase.GetPlacementByIDOutput,
) dto.GetPlacementByIDResponse {
	return dto.GetPlacementByIDResponse{
		ID:               resp.ID,
		ProjectPatternID: resp.ProjectPatternID,
		ProjectSurfaceID: resp.ProjectSurfaceID,
		X:                resp.X,
		Y:                resp.Y,
		Rotation:         resp.Rotation,
		Geometry:         toGeometryPolygon(resp.Geometry),
		CreatedAt:        resp.CreatedAt,
		UpdatedAt:        resp.UpdatedAt,
	}
}

func ToListPlacementsInput(
	userID int64,
	projectID int64,
	projectSurfaceID int64,
) placementusecase.ListPlacementsInput {
	return placementusecase.ListPlacementsInput{
		UserID:           userID,
		ProjectID:        projectID,
		ProjectSurfaceID: projectSurfaceID,
	}
}

func ToListPlacementsResponse(
	resp placementusecase.ListPlacementsOutput,
) dto.ListPlacementsResponse {
	placements := dto.ListPlacementsResponse{
		Items: make([]dto.ListPlacementsItemResponse, 0, len(resp.Items)),
	}

	for _, item := range resp.Items {
		placements.Items = append(placements.Items, dto.ListPlacementsItemResponse{
			ID:               item.ID,
			ProjectPatternID: item.ProjectPatternID,
			ProjectSurfaceID: item.ProjectSurfaceID,
			X:                item.X,
			Y:                item.Y,
			Rotation:         item.Rotation,
			Geometry:         toGeometryPolygon(item.Geometry),
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
		})
	}

	return placements
}

func ToUpdatePlacementInput(
	req dto.UpdatePlacementRequest,
	userID int64,
	projectID int64,
	placementID int64,
) placementusecase.UpdatePlacementInput {
	return placementusecase.UpdatePlacementInput{
		UserID:      userID,
		ProjectID:   projectID,
		PlacementID: placementID,
		X:           req.X,
		Y:           req.Y,
		Rotation:    req.Rotation,
	}
}

func ToUpdatePlacementResponse(
	resp placementusecase.UpdatePlacementOutput,
) dto.UpdatePlacementResponse {
	return dto.UpdatePlacementResponse{
		ID:               resp.ID,
		ProjectPatternID: resp.ProjectPatternID,
		ProjectSurfaceID: resp.ProjectSurfaceID,
		X:                resp.X,
		Y:                resp.Y,
		Rotation:         resp.Rotation,
		Geometry:         toGeometryPolygon(resp.Geometry),
		CreatedAt:        resp.CreatedAt,
		UpdatedAt:        resp.UpdatedAt,
	}
}
