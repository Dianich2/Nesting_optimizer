package geometry

import (
	"context"
	"errors"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
)

var ErrComplexity = errors.New("nesting geometry exceeds the computation complexity budget")

type NFPBuilder interface {
	BuildExternal(
		stationary domaingeometry.Polygon,
		moving domaingeometry.Polygon,
	) (domaingeometry.MultiPolygon, error)
}

// ContextNFPBuilder extends existing builders without breaking diagnostic adapters.
type ContextNFPBuilder interface {
	BuildExternalContext(context.Context, domaingeometry.Polygon, domaingeometry.Polygon) (domaingeometry.MultiPolygon, error)
}

func BuildExternal(ctx context.Context, builder NFPBuilder, stationary, moving domaingeometry.Polygon) (domaingeometry.MultiPolygon, error) {
	if err := ctx.Err(); err != nil {
		return domaingeometry.MultiPolygon{}, err
	}
	if contextual, ok := builder.(ContextNFPBuilder); ok {
		return contextual.BuildExternalContext(ctx, stationary, moving)
	}
	result, err := builder.BuildExternal(stationary, moving)
	if ctx.Err() != nil {
		return domaingeometry.MultiPolygon{}, ctx.Err()
	}
	return result, err
}
