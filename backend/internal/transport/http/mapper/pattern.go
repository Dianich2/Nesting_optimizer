package mapper

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	patternusecase "server_nesting_optimizer/internal/usecase/pattern"
)

func ToCreatePatternInput(
	req dto.CreatePatternRequest,
	userID int64,
) patternusecase.CreatePatternInput {
	return patternusecase.CreatePatternInput{
		UserID:   userID,
		Name:     req.Name,
		Geometry: toDomainPolygon(req.Geometry),
	}
}

func ToCreatePatternResponse(
	resp patternusecase.CreatePatternOutput,
) dto.CreatePatternResponse {
	return dto.CreatePatternResponse{
		ID:        resp.ID,
		Name:      resp.Name,
		Geometry:  toGeometryPolygon(resp.Geometry),
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}

func ToGetPatternInput(
	patternID int64,
	userID int64,
) patternusecase.GetPatternByIDInput {
	return patternusecase.GetPatternByIDInput{
		PatternID: patternID,
		UserID:    userID,
	}
}

func ToGetPatternResponse(
	resp patternusecase.GetPatternByIDOutput,
) dto.GetPatternResponse {
	return dto.GetPatternResponse{
		ID:        resp.ID,
		Name:      resp.Name,
		Geometry:  toGeometryPolygon(resp.Geometry),
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}

func ToListPatternsInput(
	page int,
	pageSize int,
	userID int64,
) patternusecase.ListPatternsInput {
	return patternusecase.ListPatternsInput{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	}
}

func ToListPatternsResponse(
	resp patternusecase.ListPatternsOutput,
) dto.ListPatternsResponse {
	items := make(
		[]dto.ListPatternsItemResponse,
		0,
		len(resp.Items),
	)

	for _, pattern := range resp.Items {
		items = append(
			items,
			dto.ListPatternsItemResponse{
				ID:        pattern.ID,
				Name:      pattern.Name,
				Geometry:  toGeometryPolygon(pattern.Geometry),
				CreatedAt: pattern.CreatedAt,
				UpdatedAt: pattern.UpdatedAt,
			},
		)
	}

	return dto.ListPatternsResponse{
		Items:      items,
		Page:       resp.Page,
		PageSize:   resp.PageSize,
		Total:      resp.Total,
		TotalPages: resp.TotalPages,
	}
}

func ToUpdatePatternInput(
	req dto.UpdatePatternRequest,
	userID int64,
	patternID int64,
) patternusecase.UpdatePatternInput {
	pattern := patternusecase.UpdatePatternInput{
		PatternID: patternID,
		UserID:    userID,
		Name:      req.Name,
	}

	if req.Geometry != nil {
		geometry := toDomainPolygon(*req.Geometry)
		pattern.Geometry = &geometry
	}

	return pattern
}

func ToUpdatePatternResponse(
	resp patternusecase.UpdatePatternOutput,
) dto.UpdatePatternResponse {
	return dto.UpdatePatternResponse{
		ID:        resp.ID,
		Name:      resp.Name,
		Geometry:  toGeometryPolygon(resp.Geometry),
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}

func ToDeletePatternInput(
	patternID int64,
	userID int64,
) patternusecase.DeletePatternInput {
	return patternusecase.DeletePatternInput{
		PatternID: patternID,
		UserID:    userID,
	}
}
