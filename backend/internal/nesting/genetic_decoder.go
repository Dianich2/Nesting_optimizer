package nesting

import (
	"context"
	"fmt"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

type DecodedChromosome struct {
	Placements []Placement
	Fitness    Fitness
}

func decodeChromosome(
	ctx context.Context,
	engine geometry.Engine,
	nfpBuilder geometry.NFPBuilder,
	problem Problem,
	chromosome Chromosome,
	patternLookup map[int64]PatternItem,
) (DecodedChromosome, error) {
	if err := ctx.Err(); err != nil {
		return DecodedChromosome{}, fmt.Errorf(
			"decode chromosome: %w",
			err,
		)
	}

	placements := make([]Placement, 0, len(chromosome.Genes))

	occupied := append(
		[]domaingeometry.Polygon(nil),
		problem.Obstacles...,
	)

	placedGeometries := make(
		[]domaingeometry.Polygon,
		0,
		len(chromosome.Genes),
	)

	unplacedCount := 0

	for _, gene := range chromosome.Genes {
		if err := ctx.Err(); err != nil {
			return DecodedChromosome{}, fmt.Errorf(
				"decode chromosome: %w",
				err,
			)
		}

		patternItem, ok := patternLookup[gene.PatternID]
		if !ok {
			return DecodedChromosome{}, fmt.Errorf(
				"decode chromosome: pattern item %d not found",
				gene.PatternID,
			)
		}

		placement, found, err := placeOnePatternNFP(
			ctx,
			engine,
			nfpBuilder,
			patternItem.Geometry,
			problem.Surface,
			occupied,
			[]float64{gene.Rotation},
		)
		if err != nil {
			return DecodedChromosome{}, fmt.Errorf(
				"decode chromosome: %w",
				err,
			)
		}

		if !found {
			unplacedCount++
			continue
		}

		placements = append(placements, Placement{
			PatternID: gene.PatternID,
			X:         placement.Candidate.X,
			Y:         placement.Candidate.Y,
			Rotation:  placement.Candidate.Rotation,
		})

		occupied = append(
			occupied,
			placement.Geometry,
		)

		placedGeometries = append(
			placedGeometries,
			placement.Geometry,
		)
	}

	usedArea, usedHeight, err := calculateDecodedLayoutSize(
		engine,
		problem.Surface,
		placedGeometries,
	)
	if err != nil {
		return DecodedChromosome{}, fmt.Errorf(
			"decode chromosome: %w",
			err,
		)
	}

	return DecodedChromosome{
		Placements: placements,
		Fitness: Fitness{
			UnplacedCount: unplacedCount,
			UsedArea:      usedArea,
			UsedHeight:    usedHeight,
		},
	}, nil
}

func calculateDecodedLayoutSize(
	engine geometry.Engine,
	surface domaingeometry.Polygon,
	placed []domaingeometry.Polygon,
) (float64, float64, error) {
	if len(placed) == 0 {
		return 0, 0, nil
	}

	surfaceBounds, err := engine.Bounds(surface)
	if err != nil {
		return 0, 0, fmt.Errorf(
			"calculate decoded layout size: %w",
			err,
		)
	}

	firstBounds, err := engine.Bounds(placed[0])
	if err != nil {
		return 0, 0, fmt.Errorf(
			"calculate decoded layout size: %w",
			err,
		)
	}

	maxX := firstBounds.MaxX
	maxY := firstBounds.MaxY

	for i := 1; i < len(placed); i++ {
		bounds, err := engine.Bounds(placed[i])
		if err != nil {
			return 0, 0, fmt.Errorf(
				"calculate decoded layout size: %w",
				err,
			)
		}

		if bounds.MaxX > maxX {
			maxX = bounds.MaxX
		}

		if bounds.MaxY > maxY {
			maxY = bounds.MaxY
		}
	}

	usedWidth := maxX - surfaceBounds.MinX
	usedHeight := maxY - surfaceBounds.MinY
	usedArea := usedWidth * usedHeight

	return usedArea, usedHeight, nil
}
