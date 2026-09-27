package nesting

import (
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertGenes(
	t *testing.T,
	want Gene,
	got Gene,
	delta float64,
) {
	t.Helper()

	assert.Equal(t, want.PatternID, got.PatternID)
	assert.Equal(t, want.InstanceID, got.InstanceID)
	assert.InDelta(t, want.Rotation, got.Rotation, delta)
}

func TestExpandPatternInstances(t *testing.T) {
	// arrange
	patterns := []PatternItem{
		{
			PatternID: 1,
			Geometry: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 10, Y: 0},
						{X: 10, Y: 10},
						{X: 0, Y: 10},
					},
				},
			},
			Quantity: 3,
		},
		{
			PatternID: 2,
			Geometry: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			Quantity: 2,
		},
	}

	wantGenes := []Gene{
		{
			InstanceID: 0,
			PatternID:  1,
			Rotation:   0,
		},
		{
			InstanceID: 1,
			PatternID:  1,
			Rotation:   0,
		},
		{
			InstanceID: 2,
			PatternID:  1,
			Rotation:   0,
		},
		{
			InstanceID: 3,
			PatternID:  2,
			Rotation:   0,
		},
		{
			InstanceID: 4,
			PatternID:  2,
			Rotation:   0,
		},
	}

	// act
	genes := expandPatternInstances(patterns)

	// assert
	assert.Len(
		t,
		genes,
		len(wantGenes),
	)

	for i := 0; i < len(wantGenes); i++ {
		assertGenes(
			t,
			wantGenes[i],
			genes[i],
			config.Epsilon,
		)
	}

}

func TestExpandPatternInstancesSingleInstance(t *testing.T) {
	// arrange
	patterns := []PatternItem{
		{
			PatternID: 1,
			Geometry: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			Quantity: 1,
		},
	}

	wantGenes := []Gene{
		{
			InstanceID: 0,
			PatternID:  1,
			Rotation:   0,
		},
	}

	// act
	genes := expandPatternInstances(patterns)

	// assert
	assert.Len(
		t,
		genes,
		len(wantGenes),
	)

	for i := 0; i < len(wantGenes); i++ {
		assertGenes(
			t,
			wantGenes[i],
			genes[i],
			config.Epsilon,
		)
	}
}

func TestExpandPatternInstancesEmptyPatterns(t *testing.T) {
	// arrange
	patterns := []PatternItem{}

	// act
	genes := expandPatternInstances(patterns)

	// assert
	assert.Len(
		t,
		genes,
		0,
	)

	assert.Equal(
		t,
		[]Gene{},
		genes,
	)
}
