package nesting

type Gene struct {
	InstanceID int
	PatternID  int64
	Rotation   float64
}

type Chromosome struct {
	Genes []Gene
}

type EvaluatedChromosome struct {
	Chromosome Chromosome
	Fitness    Fitness
}

func expandPatternInstances(
	patterns []PatternItem,
) []Gene {
	genes := make([]Gene, 0)
	instanceID := 0

	for _, patternItem := range patterns {
		for i := 0; i < patternItem.Quantity; i++ {
			genes = append(genes, Gene{
				InstanceID: instanceID,
				PatternID:  patternItem.PatternID,
				Rotation:   0,
			})

			instanceID++
		}
	}

	return genes
}

func buildPatternLookup(
	patterns []PatternItem,
) map[int64]PatternItem {
	patternLookup := make(map[int64]PatternItem)

	for _, pattern := range patterns {
		patternLookup[pattern.PatternID] = pattern
	}

	return patternLookup
}
