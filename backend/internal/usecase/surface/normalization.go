package surface

import "strings"

func normalizeCreateSurfaceInput(
	input CreateSurfaceInput,
) CreateSurfaceInput {
	input.Name = strings.TrimSpace(input.Name)

	return input
}
