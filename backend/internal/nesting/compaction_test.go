package nesting

import (
	"context"
	"testing"

	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func TestCompactPlacements(t *testing.T) {
	engine := simplefeatures.NewEngine()

	compactionConfig := CompactionConfig{
		MaxPasses:         2,
		InitialStepFactor: 0.1,
		MinStepFactor:     0.0001,
	}

	tests := []struct {
		name       string
		problem    Problem
		placements []Placement
		want       []Placement
		delta      float64
	}{
		{
			name: "empty placements",
			problem: Problem{
				Surface: rectangle(100, 100),
			},
			placements: []Placement{},
			want:       []Placement{},
		},
		{
			name: "single placement moves to bottom left",
			problem: Problem{
				Surface: rectangle(100, 100),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  1,
						Geometry:  rectangle(20, 20),
					},
				},
			},
			placements: []Placement{
				{
					PatternID: 1,
					X:         70,
					Y:         70,
					Rotation:  0,
				},
			},
			want: []Placement{
				{
					PatternID: 1,
					X:         10,
					Y:         10,
					Rotation:  0,
				},
			},
		},
		{
			name: "second placement touches first after compaction",
			problem: Problem{
				Surface: rectangle(100, 100),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  2,
						Geometry:  rectangle(20, 20),
					},
				},
			},
			placements: []Placement{
				{
					PatternID: 1,
					X:         10,
					Y:         10,
					Rotation:  0,
				},
				{
					PatternID: 1,
					X:         70,
					Y:         10,
					Rotation:  0,
				},
			},
			want: []Placement{
				{
					PatternID: 1,
					X:         10,
					Y:         10,
					Rotation:  0,
				},
				{
					PatternID: 1,
					X:         30,
					Y:         10,
					Rotation:  0,
				},
			},
		},
		{
			name: "existing obstacle respected",
			problem: Problem{
				Surface: rectangle(100, 100),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  1,
						Geometry:  rectangle(20, 20),
					},
				},
				Obstacles: []domaingeometry.Polygon{
					rectangle(20, 20),
				},
			},
			placements: []Placement{
				{
					PatternID: 1,
					X:         70,
					Y:         10,
					Rotation:  0,
				},
			},
			want: []Placement{
				{
					PatternID: 1,
					X:         30,
					Y:         10,
					Rotation:  0,
				},
			},
		},
		{
			name: "rotated placement compacted using rotated bounds",
			problem: Problem{
				Surface: rectangle(100, 100),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  1,
						Geometry:  rectangle(40, 20),
					},
				},
			},
			placements: []Placement{
				{
					PatternID: 1,
					X:         70,
					Y:         70,
					Rotation:  90,
				},
			},
			want: []Placement{
				{
					PatternID: 1,
					X:         10,
					Y:         20,
					Rotation:  90,
				},
			},
			delta: 0.01 + config.Epsilon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compactPlacements(
				context.Background(),
				engine,
				tt.problem,
				tt.placements,
				compactionConfig,
			)

			require.NoError(t, err)

			require.Len(
				t,
				got,
				len(tt.want),
			)

			delta := tt.delta
			if delta == 0 {
				delta = config.Epsilon
			}

			for i := range tt.want {
				assertPlacementInDelta(
					t,
					tt.want[i],
					got[i],
					delta,
				)
			}
		})
	}
}

func TestCompactPlacementsDoesNotMutateInput(t *testing.T) {
	engine := simplefeatures.NewEngine()

	placements := []Placement{
		{
			PatternID: 1,
			X:         70,
			Y:         70,
			Rotation:  0,
		},
	}

	original := append(
		[]Placement(nil),
		placements...,
	)

	_, err := compactPlacements(
		context.Background(),
		engine,
		Problem{
			Surface: rectangle(100, 100),
			Patterns: []PatternItem{
				{
					PatternID: 1,
					Quantity:  1,
					Geometry:  rectangle(20, 20),
				},
			},
		},
		placements,
		CompactionConfig{
			MaxPasses:         2,
			InitialStepFactor: 0.1,
			MinStepFactor:     0.0001,
		},
	)
	require.NoError(t, err)

	require.Equal(
		t,
		original,
		placements,
	)
}

func TestCompactPlacementsCanceledContext(t *testing.T) {
	engine := simplefeatures.NewEngine()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := compactPlacements(
		ctx,
		engine,
		Problem{
			Surface: rectangle(100, 100),
			Patterns: []PatternItem{
				{
					PatternID: 1,
					Quantity:  1,
					Geometry:  rectangle(20, 20),
				},
			},
		},
		[]Placement{
			{
				PatternID: 1,
				X:         70,
				Y:         70,
				Rotation:  0,
			},
		},
		DefaultCompactionConfig(),
	)

	require.ErrorIs(
		t,
		err,
		context.Canceled,
	)
}

func TestValidateCompactionConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  CompactionConfig
		wantErr bool
	}{
		{
			name:   "default config",
			config: DefaultCompactionConfig(),
		},
		{
			name: "zero max passes",
			config: CompactionConfig{
				MaxPasses:         0,
				InitialStepFactor: 0.1,
				MinStepFactor:     0.001,
			},
			wantErr: true,
		},
		{
			name: "zero initial step factor",
			config: CompactionConfig{
				MaxPasses:         1,
				InitialStepFactor: 0,
				MinStepFactor:     0.001,
			},
			wantErr: true,
		},
		{
			name: "initial step factor greater than one",
			config: CompactionConfig{
				MaxPasses:         1,
				InitialStepFactor: 1.1,
				MinStepFactor:     0.001,
			},
			wantErr: true,
		},
		{
			name: "zero minimum step factor",
			config: CompactionConfig{
				MaxPasses:         1,
				InitialStepFactor: 0.1,
				MinStepFactor:     0,
			},
			wantErr: true,
		},
		{
			name: "minimum equals initial step factor",
			config: CompactionConfig{
				MaxPasses:         1,
				InitialStepFactor: 0.1,
				MinStepFactor:     0.1,
			},
			wantErr: true,
		},
		{
			name: "minimum greater than initial step factor",
			config: CompactionConfig{
				MaxPasses:         1,
				InitialStepFactor: 0.1,
				MinStepFactor:     0.2,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCompactionConfig(
				tt.config,
			)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
