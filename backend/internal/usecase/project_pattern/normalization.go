package projectpattern

import "strings"

func normalizeUpdateProjectPatternInput(
	input UpdateProjectPatternInput,
) UpdateProjectPatternInput {
	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)
	}

	return input
}
