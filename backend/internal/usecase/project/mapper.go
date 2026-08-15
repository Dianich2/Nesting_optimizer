package project

import domainproject "server_nesting_optimizer/internal/domain/project"

func toCreateProjectOutput(
	project domainproject.Project,
) CreateProjectOutput {
	return CreateProjectOutput{
		ID:          project.ID,
		UserID:      project.UserID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func toProject(
	input CreateProjectInput,
) domainproject.Project {
	return domainproject.Project{
		UserID:      input.UserID,
		Name:        input.Name,
		Description: input.Description,
	}
}

func toGetProjectByIDOutput(
	project domainproject.Project,
) GetProjectByIDOutput {
	return GetProjectByIDOutput{
		ID:          project.ID,
		UserID:      project.UserID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func toUpdateProjectOutput(
	project domainproject.Project,
) UpdateProjectOutput {
	return UpdateProjectOutput{
		ID:          project.ID,
		UserID:      project.UserID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}
