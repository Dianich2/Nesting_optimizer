package placement

import (
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
	"time"
)

func placementConflict(reason apperror.FieldCode, message string) *apperror.AppError {
	err := apperror.Conflict(message)
	err.Details = []apperror.FieldError{apperror.NewFieldError("placement", reason, message)}
	return err
}

func coverageConflict(engine geometry.Engine, polygon, surface dg.Polygon) error {
	if len(surface.Holes) > 0 {
		insideExterior, err := engine.CoveredBy(polygon, dg.Polygon{Exterior: surface.Exterior})
		if err != nil {
			return apperror.Internal("failed to check exterior coverage", err)
		}
		if insideExterior {
			return placementConflict("placement_in_hole", "placement enters a project surface hole")
		}
	}
	return placementConflict("placement_outside_surface", "placement is outside project surface")
}

func checkVersion(expected string, updatedAt time.Time) error {
	if expected == "" {
		return nil
	}
	version, err := time.Parse(time.RFC3339Nano, expected)
	if err != nil {
		return apperror.Validation("invalid placement version", apperror.NewFieldError("placement_version", apperror.FieldCodeInvalid, "expected an RFC3339 timestamp"))
	}
	if !version.Equal(updatedAt) {
		return placementConflict("placement_version_conflict", "placement was modified")
	}
	return nil
}
