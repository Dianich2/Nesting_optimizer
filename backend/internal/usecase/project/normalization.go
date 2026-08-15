package project

import "strings"

func normalizeCreateProjectInput(
	input CreateProjectInput,
) CreateProjectInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	return input
}

func normalizeUpdateProjectInput(
	input UpdateProjectInput,
) UpdateProjectInput {
	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)
	}

	if input.Description != nil {
		*input.Description = strings.TrimSpace(*input.Description)
	}

	return input
}
