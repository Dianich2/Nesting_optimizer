package project

import "strings"

func normalizeCreateProjectInput(
	input CreateProjectInput,
) CreateProjectInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	return input
}
