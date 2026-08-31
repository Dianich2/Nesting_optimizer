package nesting

type Algorithm string

const (
	BaselineAlgorithm  Algorithm = "baseline"
	NFPGreedyAlgorithm Algorithm = "nfp_greedy"
)

func (a Algorithm) IsValid() bool {
	switch a {
	case BaselineAlgorithm:
		return true
	case NFPGreedyAlgorithm:
		return true
	default:
		return false
	}
}
