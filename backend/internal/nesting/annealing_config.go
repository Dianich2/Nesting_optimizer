package nesting

import "fmt"

type AnnealingConfig struct {
	InitialTemperature float64
	MinTemperature     float64
	CoolingRate        float64
	MaxIterations      int
}

func DefaultAnnealingConfig() AnnealingConfig {
	return AnnealingConfig{
		InitialTemperature: 1.0,
		MinTemperature:     0.001,
		CoolingRate:        0.995,
		MaxIterations:      1000,
	}
}

func validateAnnealingConfig(
	config AnnealingConfig,
) error {
	if config.InitialTemperature <= 0 {
		return fmt.Errorf(
			"annealing initial temperature must be greater than zero",
		)
	}

	if config.MinTemperature <= 0 {
		return fmt.Errorf(
			"annealing minimum temperature must be greater than zero",
		)
	}

	if config.MinTemperature >= config.InitialTemperature {
		return fmt.Errorf(
			"annealing minimum temperature must be less than initial temperature",
		)
	}

	if config.CoolingRate <= 0 ||
		config.CoolingRate >= 1 {
		return fmt.Errorf(
			"annealing cooling rate must be between zero and one",
		)
	}

	if config.MaxIterations <= 0 {
		return fmt.Errorf(
			"annealing max iterations must be greater than zero",
		)
	}

	return nil
}
