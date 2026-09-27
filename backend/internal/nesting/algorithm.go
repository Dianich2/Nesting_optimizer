package nesting

type Algorithm string

const (
	BaselineAlgorithm           Algorithm = "baseline"
	NFPGreedyAlgorithm          Algorithm = "nfp_greedy"
	GeneticAlgorithm            Algorithm = "genetic"
	SimulatedAnnealingAlgorithm Algorithm = "simulated_annealing"
)

func (a Algorithm) IsValid() bool {
	switch a {
	case BaselineAlgorithm,
		NFPGreedyAlgorithm,
		GeneticAlgorithm,
		SimulatedAnnealingAlgorithm:
		return true
	default:
		return false
	}
}
