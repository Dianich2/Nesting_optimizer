package nesting

import (
	"math"
	optimizer "server_nesting_optimizer/internal/nesting"
	"testing"
)

func TestRunNestingInputWorkLimits(t *testing.T) {
	for _, test := range []struct {
		name      string
		quantity  int
		rotations int
		wantError bool
	}{
		{"boundary", 500, 36, false},
		{"too many instances", 501, 1, true},
		{"integer overflow", math.MaxInt, 1, true},
		{"too many rotations", 1, 37, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := RunNestingInput{
				UserID: 1, ProjectID: 1, ProjectSurfaceID: 1,
				Algorithm:        optimizer.BaselineAlgorithm,
				Patterns:         []RunNestingPatternInput{{ProjectPatternID: 1, Quantity: test.quantity}},
				AllowedRotations: make([]float64, test.rotations),
			}
			for i := range input.AllowedRotations {
				input.AllowedRotations[i] = float64(i)
			}
			if gotError := len(input.Validate()) > 0; gotError != test.wantError {
				t.Fatalf("validation errors = %v, expected error %v", input.Validate(), test.wantError)
			}
		})
	}
	input := RunNestingInput{
		UserID: 1, ProjectID: 1, ProjectSurfaceID: 1,
		Algorithm: optimizer.BaselineAlgorithm, AllowedRotations: []float64{0},
		Patterns: []RunNestingPatternInput{
			{ProjectPatternID: 1, Quantity: 250},
			{ProjectPatternID: 2, Quantity: 251},
		},
	}
	if len(input.Validate()) == 0 {
		t.Fatal("the limit must apply to the sum of quantities")
	}
}
