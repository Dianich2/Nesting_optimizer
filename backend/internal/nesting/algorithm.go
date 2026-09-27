package nesting

type Algorithm string

const (
	BaselineAlgorithm  Algorithm = "baseline"
	NFPGreedyAlgorithm Algorithm = "nfp_greedy"
	GeneticAlgorithm   Algorithm = "genetic"
)

func (a Algorithm) IsValid() bool {
	switch a {
	case BaselineAlgorithm,
		NFPGreedyAlgorithm,
		GeneticAlgorithm:
		return true
	default:
		return false
	}
}
