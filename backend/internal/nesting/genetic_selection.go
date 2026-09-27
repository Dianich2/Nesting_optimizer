package nesting

func tournamentSelection(
	population []EvaluatedChromosome,
	tournamentSize int,
	rng Random,
) EvaluatedChromosome {
	firstCandidateIdx := rng.Intn(len(population))
	best := population[firstCandidateIdx]

	for i := 1; i < tournamentSize; i++ {
		curCandidateIdx := rng.Intn(len(population))
		curCandidate := population[curCandidateIdx]

		if isFitnessBetter(curCandidate.Fitness, best.Fitness) {
			best = curCandidate
		}
	}

	return best
}
