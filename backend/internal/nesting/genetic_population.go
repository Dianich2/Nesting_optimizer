package nesting

func generateInitialPopulation(
	genes []Gene,
	allowedRotations []float64,
	populationSize int,
	rng Random,
) []Chromosome {
	population := make([]Chromosome, 0, populationSize)

	for i := 0; i < populationSize; i++ {
		curGenes := make([]Gene, len(genes))
		copy(curGenes, genes)

		rng.Shuffle(
			len(curGenes),
			func(i int, j int) {
				curGenes[i], curGenes[j] = curGenes[j], curGenes[i]
			},
		)

		for j := range curGenes {
			randIdx := rng.Intn(len(allowedRotations))
			curGenes[j].Rotation = allowedRotations[randIdx]
		}

		population = append(population, Chromosome{
			Genes: curGenes,
		})
	}

	return population
}
