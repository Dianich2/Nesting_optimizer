package nesting

import (
	"math"
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func containsCandidate(
	candidates []Candidate,
	want Candidate,
) bool {
	for _, candidate := range candidates {
		if assert.ObjectsAreEqualValues(candidate.Rotation, want.Rotation) &&
			math.Abs(candidate.X-want.X) <= config.Epsilon &&
			math.Abs(candidate.Y-want.Y) <= config.Epsilon {
			return true
		}
	}

	return false
}

func TestGenerateNFPCandidatesWithoutOccupied(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := nfp.NewBuilder(engine)

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 100, Y: 100},
				{X: 0, Y: 100},
			},
		},
	}

	pattern := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -10, Y: -5},
				{X: 10, Y: -5},
				{X: 10, Y: 5},
				{X: -10, Y: 5},
			},
		},
	}

	candidates, err := generateNFPCandidates(
		builder,
		pattern,
		surface,
		nil,
		0,
	)

	require.NoError(t, err)
	require.NotEmpty(t, candidates)
}

func TestGenerateNFPCandidatesFromOccupiedRectangle(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := nfp.NewBuilder(engine)

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -100, Y: -100},
				{X: 100, Y: -100},
				{X: 100, Y: 100},
				{X: -100, Y: 100},
			},
		},
	}

	rotatedPattern := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -1, Y: -0.5},
				{X: 1, Y: -0.5},
				{X: 1, Y: 0.5},
				{X: -1, Y: 0.5},
			},
		},
	}

	occupied := []domaingeometry.Polygon{
		{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 4, Y: 0},
					{X: 4, Y: 3},
					{X: 0, Y: 3},
				},
			},
		},
	}

	candidates, err := generateNFPCandidates(
		builder,
		rotatedPattern,
		surface,
		occupied,
		0,
	)

	require.NoError(t, err)

	want := []Candidate{
		{X: -1, Y: -0.5, Rotation: 0},
		{X: 5, Y: -0.5, Rotation: 0},
		{X: 5, Y: 3.5, Rotation: 0},
		{X: -1, Y: 3.5, Rotation: 0},
	}

	for _, candidate := range want {
		assert.True(
			t,
			containsCandidate(candidates, candidate),
			"missing candidate: %+v",
			candidate,
		)
	}
}

func TestGenerateNFPCandidatesFromMultipleOccupied(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := nfp.NewBuilder(engine)

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -100, Y: -100},
				{X: 100, Y: -100},
				{X: 100, Y: 100},
				{X: -100, Y: 100},
			},
		},
	}

	moving := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -1, Y: -1},
				{X: 1, Y: -1},
				{X: 1, Y: 1},
				{X: -1, Y: 1},
			},
		},
	}

	occupied := []domaingeometry.Polygon{
		{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 4, Y: 0},
					{X: 4, Y: 4},
					{X: 0, Y: 4},
				},
			},
		},
		{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 20, Y: 20},
					{X: 24, Y: 20},
					{X: 24, Y: 24},
					{X: 20, Y: 24},
				},
			},
		},
	}

	candidates, err := generateNFPCandidates(
		builder,
		moving,
		surface,
		occupied,
		0,
	)

	require.NoError(t, err)

	assert.True(t, containsCandidate(
		candidates,
		Candidate{X: -1, Y: -1, Rotation: 0},
	))

	assert.True(t, containsCandidate(
		candidates,
		Candidate{X: 19, Y: 19, Rotation: 0},
	))
}

func TestGenerateNFPCandidatesFromNFPHole(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := nfp.NewBuilder(engine)

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -200, Y: -200},
				{X: 200, Y: -200},
				{X: 200, Y: 200},
				{X: -200, Y: 200},
			},
		},
	}

	occupied := []domaingeometry.Polygon{
		{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 100, Y: 0},
					{X: 100, Y: 100},
					{X: 0, Y: 100},
				},
			},
			Holes: []domaingeometry.Ring{
				{
					Points: []domaingeometry.Point{
						{X: 25, Y: 25},
						{X: 25, Y: 75},
						{X: 75, Y: 75},
						{X: 75, Y: 25},
					},
				},
			},
		},
	}

	moving := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -5, Y: -5},
				{X: 5, Y: -5},
				{X: 5, Y: 5},
				{X: -5, Y: 5},
			},
		},
	}

	candidates, err := generateNFPCandidates(
		builder,
		moving,
		surface,
		occupied,
		0,
	)

	require.NoError(t, err)

	for _, want := range []Candidate{
		{X: 30, Y: 30, Rotation: 0},
		{X: 70, Y: 30, Rotation: 0},
		{X: 70, Y: 70, Rotation: 0},
		{X: 30, Y: 70, Rotation: 0},
	} {
		assert.True(
			t,
			containsCandidate(candidates, want),
			"missing hole candidate: %+v",
			want,
		)
	}
}

func TestGenerateNFPCandidatesNoDuplicates(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := nfp.NewBuilder(engine)

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 10, Y: 10},
				{X: 0, Y: 10},
			},
		},
	}

	pattern := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -1, Y: -1},
				{X: 1, Y: -1},
				{X: 1, Y: 1},
				{X: -1, Y: 1},
			},
		},
	}

	candidates, err := generateNFPCandidates(
		builder,
		pattern,
		surface,
		nil,
		0,
	)

	require.NoError(t, err)

	counts := make(map[Candidate]int)

	for _, candidate := range candidates {
		counts[candidate]++
	}

	for candidate, count := range counts {
		assert.Equal(
			t,
			1,
			count,
			"duplicate candidate: %+v",
			candidate,
		)
	}
}

func TestGenerateNFPCandidatesForRotationsNormalizesAndDeduplicates(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := nfp.NewBuilder(engine)

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 100, Y: 100},
				{X: 0, Y: 100},
			},
		},
	}

	pattern := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -10, Y: -5},
				{X: 10, Y: -5},
				{X: 10, Y: 5},
				{X: -10, Y: 5},
			},
		},
	}

	candidates, err := generateNFPCandidatesForRotations(
		engine,
		builder,
		pattern,
		surface,
		nil,
		[]float64{0, 90, 360, -360},
	)

	require.NoError(t, err)
	require.NotEmpty(t, candidates)

	hasZero := false
	hasNinety := false

	for _, candidate := range candidates {
		switch candidate.Rotation {
		case 0:
			hasZero = true
		case 90:
			hasNinety = true
		default:
			t.Fatalf(
				"unexpected normalized rotation: %v",
				candidate.Rotation,
			)
		}
	}

	assert.True(t, hasZero)
	assert.True(t, hasNinety)

	counts := make(map[Candidate]int)
	for _, candidate := range candidates {
		counts[candidate]++
	}

	for candidate, count := range counts {
		assert.Equal(
			t,
			1,
			count,
			"duplicate candidate: %+v",
			candidate,
		)
	}
}
