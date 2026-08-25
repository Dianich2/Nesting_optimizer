package mapper

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	surfaceusecase "server_nesting_optimizer/internal/usecase/surface"
)

func ToCreateSurfaceInput(
	req dto.CreateSurfaceRequest,
	userID int64,
) surfaceusecase.CreateSurfaceInput {
	return surfaceusecase.CreateSurfaceInput{
		UserID:   userID,
		Name:     req.Name,
		Geometry: toDomainPolygon(req.Geometry),
	}
}

func ToCreateSurfaceResponse(
	resp surfaceusecase.CreateSurfaceOutput,
) dto.CreateSurfaceResponse {
	return dto.CreateSurfaceResponse{
		ID:        resp.ID,
		Name:      resp.Name,
		Geometry:  toGeometryPolygon(resp.Geometry),
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}

func ToGetSurfaceInput(
	surfaceID int64,
	userID int64,
) surfaceusecase.GetSurfaceByIDInput {
	return surfaceusecase.GetSurfaceByIDInput{
		SurfaceID: surfaceID,
		UserID:    userID,
	}
}

func ToGetSurfaceResponse(
	resp surfaceusecase.GetSurfaceByIDOutput,
) dto.GetSurfaceResponse {
	return dto.GetSurfaceResponse{
		ID:        resp.ID,
		Name:      resp.Name,
		Geometry:  toGeometryPolygon(resp.Geometry),
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}

func ToListSurfacesInput(
	page int,
	pageSize int,
	userID int64,
) surfaceusecase.ListSurfacesInput {
	return surfaceusecase.ListSurfacesInput{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	}
}

func ToListSurfacesResponse(
	resp surfaceusecase.ListSurfacesOutput,
) dto.ListSurfacesResponse {
	items := make(
		[]dto.ListSurfacesItemResponse,
		0,
		len(resp.Items),
	)

	for _, surface := range resp.Items {
		items = append(
			items,
			dto.ListSurfacesItemResponse{
				ID:        surface.ID,
				Name:      surface.Name,
				Geometry:  toGeometryPolygon(surface.Geometry),
				CreatedAt: surface.CreatedAt,
				UpdatedAt: surface.UpdatedAt,
			},
		)
	}

	return dto.ListSurfacesResponse{
		Items:      items,
		Page:       resp.Page,
		PageSize:   resp.PageSize,
		Total:      resp.Total,
		TotalPages: resp.TotalPages,
	}
}

func ToUpdateSurfaceInput(
	req dto.UpdateSurfaceRequest,
	userID int64,
	surfaceID int64,
) surfaceusecase.UpdateSurfaceInput {
	surface := surfaceusecase.UpdateSurfaceInput{
		SurfaceID: surfaceID,
		UserID:    userID,
		Name:      req.Name,
	}

	if req.Geometry != nil {
		geometry := toDomainPolygon(*req.Geometry)
		surface.Geometry = &geometry
	}

	return surface
}

func ToUpdateSurfaceResponse(
	resp surfaceusecase.UpdateSurfaceOutput,
) dto.UpdateSurfaceResponse {
	return dto.UpdateSurfaceResponse{
		ID:        resp.ID,
		Name:      resp.Name,
		Geometry:  toGeometryPolygon(resp.Geometry),
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}

func ToDeleteSurfaceInput(
	surfaceID int64,
	userID int64,
) surfaceusecase.DeleteSurfaceInput {
	return surfaceusecase.DeleteSurfaceInput{
		SurfaceID: surfaceID,
		UserID:    userID,
	}
}
