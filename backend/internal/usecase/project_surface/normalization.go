package projectsurface

import "strings"

func normalizeUpdateProjectSurfaceInput(
	input UpdateProjectSurfaceInput,
) UpdateProjectSurfaceInput {
	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)
	}

	return input
}
