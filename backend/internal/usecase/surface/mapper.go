package surface

import domainsurface "server_nesting_optimizer/internal/domain/surface"

func toCreateSurfaceOutput(
	surface domainsurface.Surface,
) CreateSurfaceOutput {
	return CreateSurfaceOutput{
		ID:        surface.ID,
		UserID:    surface.UserID,
		Name:      surface.Name,
		Geometry:  surface.Geometry,
		CreatedAt: surface.CreatedAt,
		UpdatedAt: surface.UpdatedAt,
	}
}

func toSurface(
	input CreateSurfaceInput,
) domainsurface.Surface {
	return domainsurface.Surface{
		UserID:   input.UserID,
		Name:     input.Name,
		Geometry: input.Geometry,
	}
}
