package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCandidatesFromAnchors(t *testing.T) {
	// arrange
	anchors := []domaingeometry.Point{
		{
			X: 100,
			Y: 100,
		},
	}

	rotatedPattern := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{
					X: -10,
					Y: -5,
				},
				{
					X: 10,
					Y: -5,
				},
				{
					X: 10,
					Y: 5,
				},
				{
					X: -10,
					Y: 5,
				},
			},
		},
	}
	rotation := 20.

	wantCandidate := []Candidate{
		{
			X:        110,
			Y:        105,
			Rotation: rotation,
		},
		{
			X:        90,
			Y:        105,
			Rotation: rotation,
		},
		{
			X:        90,
			Y:        95,
			Rotation: rotation,
		},
		{
			X:        110,
			Y:        95,
			Rotation: rotation,
		},
	}

	assert.Equal(
		t,
		wantCandidate,
		candidatesFromAnchors(
			anchors,
			rotatedPattern,
			rotation,
		),
	)
}
