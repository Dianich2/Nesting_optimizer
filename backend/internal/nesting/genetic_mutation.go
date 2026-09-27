package nesting

func mutateSwap(
	chromosome Chromosome,
	firstIdx int,
	secondIdx int,
) Chromosome {
	mutatedChromosome := Chromosome{
		Genes: make([]Gene, len(chromosome.Genes)),
	}
	copy(mutatedChromosome.Genes, chromosome.Genes)

	mutatedChromosome.Genes[firstIdx], mutatedChromosome.Genes[secondIdx] =
		mutatedChromosome.Genes[secondIdx], mutatedChromosome.Genes[firstIdx]

	return mutatedChromosome
}

func mutateRotation(
	chromosome Chromosome,
	geneIdx int,
	rotation float64,
) Chromosome {
	mutatedChromosome := Chromosome{
		Genes: make([]Gene, len(chromosome.Genes)),
	}
	copy(mutatedChromosome.Genes, chromosome.Genes)

	mutatedChromosome.Genes[geneIdx].Rotation = rotation

	return mutatedChromosome
}
