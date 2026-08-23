package projectpattern

import domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"

func toCreateProjectPatternOutput(
	projectPattern domainprojectpattern.ProjectPattern,
) CreateProjectPatternOutput {
	return CreateProjectPatternOutput{
		ID:              projectPattern.ID,
		ProjectID:       projectPattern.ProjectID,
		SourcePatternID: *projectPattern.SourcePatternID,
		Name:            projectPattern.Name,
		Geometry:        projectPattern.Geometry,
		CreatedAt:       projectPattern.CreatedAt,
		UpdatedAt:       projectPattern.UpdatedAt,
	}
}

func toGetProjectPatternByIDOutput(
	projectPattern domainprojectpattern.ProjectPattern,
) GetProjectPatternByIDOutput {
	return GetProjectPatternByIDOutput{
		ID:              projectPattern.ID,
		ProjectID:       projectPattern.ProjectID,
		SourcePatternID: projectPattern.SourcePatternID,
		Name:            projectPattern.Name,
		Geometry:        projectPattern.Geometry,
		CreatedAt:       projectPattern.CreatedAt,
		UpdatedAt:       projectPattern.UpdatedAt,
	}
}

func toUpdateProjectPatternOutput(
	projectPattern domainprojectpattern.ProjectPattern,
) UpdateProjectPatternOutput {
	return UpdateProjectPatternOutput{
		ID:              projectPattern.ID,
		ProjectID:       projectPattern.ProjectID,
		SourcePatternID: projectPattern.SourcePatternID,
		Name:            projectPattern.Name,
		Geometry:        projectPattern.Geometry,
		CreatedAt:       projectPattern.CreatedAt,
		UpdatedAt:       projectPattern.UpdatedAt,
	}
}
