package nesting

type Algorithm string

const (
	BaselineAlgorithm           Algorithm = "baseline"
	NFPGreedyAlgorithm          Algorithm = "nfp_greedy"
	GeneticAlgorithm            Algorithm = "genetic"
	SimulatedAnnealingAlgorithm Algorithm = "simulated_annealing"
	HybridAlgorithm             Algorithm = "hybrid"
)

func (a Algorithm) IsValid() bool {
	switch a {
	case BaselineAlgorithm,
		NFPGreedyAlgorithm,
		GeneticAlgorithm,
		SimulatedAnnealingAlgorithm,
		HybridAlgorithm:
		return true
	default:
		return false
	}
}
