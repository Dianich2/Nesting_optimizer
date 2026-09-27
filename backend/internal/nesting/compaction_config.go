package nesting

import "fmt"

type CompactionConfig struct {
	MaxPasses         int
	InitialStepFactor float64
	MinStepFactor     float64
}

func DefaultCompactionConfig() CompactionConfig {
	return CompactionConfig{
		MaxPasses:         4,
		InitialStepFactor: 0.1,
		MinStepFactor:     0.0001,
	}
}

func validateCompactionConfig(
	config CompactionConfig,
) error {
	if config.MaxPasses <= 0 {
		return fmt.Errorf(
			"compaction max passes must be greater than zero",
		)
	}

	if config.InitialStepFactor <= 0 ||
		config.InitialStepFactor > 1 {
		return fmt.Errorf(
			"compaction initial step factor must be between zero and one",
		)
	}

	if config.MinStepFactor <= 0 {
		return fmt.Errorf(
			"compaction minimum step factor must be greater than zero",
		)
	}

	if config.MinStepFactor >= config.InitialStepFactor {
		return fmt.Errorf(
			"compaction minimum step factor must be less than initial step factor",
		)
	}

	return nil
}
