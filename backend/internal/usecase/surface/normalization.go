package surface

import "strings"

func normalizeCreateSurfaceInput(
	input CreateSurfaceInput,
) CreateSurfaceInput {
	input.Name = strings.TrimSpace(input.Name)

	return input
}

func normalizeUpdateSurfaceInput(
	input UpdateSurfaceInput,
) UpdateSurfaceInput {
	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)
	}

	return input
}
