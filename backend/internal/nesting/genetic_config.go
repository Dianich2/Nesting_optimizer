package nesting

import "fmt"

type GeneticConfig struct {
	PopulationSize int
	Generations    int
	EliteCount     int
	TournamentSize int
	MutationRate   float64
}

func DefaultGeneticConfig() GeneticConfig {
	return GeneticConfig{
		PopulationSize: 30,
		Generations:    50,
		EliteCount:     2,
		TournamentSize: 3,
		MutationRate:   0.2,
	}
}

func validateGeneticConfig(
	config GeneticConfig,
) error {
	if config.PopulationSize <= 0 {
		return fmt.Errorf(
			"genetic population size must be greater than zero",
		)
	}

	if config.Generations <= 0 {
		return fmt.Errorf(
			"genetic generations must be greater than zero",
		)
	}

	if config.EliteCount < 0 ||
		config.EliteCount > config.PopulationSize {
		return fmt.Errorf(
			"genetic elite count must be between zero and population size",
		)
	}

	if config.TournamentSize <= 0 {
		return fmt.Errorf(
			"genetic tournament size must be greater than zero",
		)
	}

	if config.MutationRate < 0 ||
		config.MutationRate > 1 {
		return fmt.Errorf(
			"genetic mutation rate must be between zero and one",
		)
	}

	return nil
}
