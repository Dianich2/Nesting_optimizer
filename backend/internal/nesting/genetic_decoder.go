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

	fitness, err := calculatePlacedGeometriesFitness(
		engine,
		problem,
		placements,
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
		Fitness:    fitness,
	}, nil
}
