package nesting

type Algorithm string

const (
	BaselineAlgorithm Algorithm = "baseline"
)

func (a Algorithm) IsValid() bool {
	switch a {
	case BaselineAlgorithm:
		return true
	default:
		return false
	}
}
