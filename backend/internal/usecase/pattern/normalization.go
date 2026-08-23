package pattern

import "strings"

func normalizeCreatePatternInput(
	input CreatePatternInput,
) CreatePatternInput {
	input.Name = strings.TrimSpace(input.Name)

	return input
}

func normalizeUpdatePatternInput(
	input UpdatePatternInput,
) UpdatePatternInput {
	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)
	}

	return input
}
