package pattern

import domainpattern "server_nesting_optimizer/internal/domain/pattern"

func toCreatePatternOutput(
	pattern domainpattern.Pattern,
) CreatePatternOutput {
	return CreatePatternOutput{
		ID:        pattern.ID,
		Name:      pattern.Name,
		Geometry:  pattern.Geometry,
		CreatedAt: pattern.CreatedAt,
		UpdatedAt: pattern.UpdatedAt,
	}
}

func toPattern(
	input CreatePatternInput,
) domainpattern.Pattern {
	return domainpattern.Pattern{
		UserID:   input.UserID,
		Name:     input.Name,
		Geometry: input.Geometry,
	}
}

func toGetPatternByIDOutput(
	pattern domainpattern.Pattern,
) GetPatternByIDOutput {
	return GetPatternByIDOutput{
		ID:        pattern.ID,
		Name:      pattern.Name,
		Geometry:  pattern.Geometry,
		CreatedAt: pattern.CreatedAt,
		UpdatedAt: pattern.UpdatedAt,
	}
}

func toUpdatePatternOutput(
	pattern domainpattern.Pattern,
) UpdatePatternOutput {
	return UpdatePatternOutput{
		ID:        pattern.ID,
		Name:      pattern.Name,
		Geometry:  pattern.Geometry,
		CreatedAt: pattern.CreatedAt,
		UpdatedAt: pattern.UpdatedAt,
	}
}
