package nesting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsScoreBetter(t *testing.T) {
	tests := []struct {
		name               string
		candidateScore     CandidateScore
		bestCandidateScore CandidateScore
		wantIsScoreBetter  bool
	}{
		{
			name: "best candidate usedArea bigger than candidate usedArea",
			candidateScore: CandidateScore{
				UsedArea:   900,
				UsedHeight: 30,
				UsedWidth:  30,
			},
			bestCandidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 40,
				UsedWidth:  30,
			},
			wantIsScoreBetter: true,
		},
		{
			name: "best candidate usedArea smaller than candidate usedArea",
			candidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 40,
				UsedWidth:  30,
			},
			bestCandidateScore: CandidateScore{
				UsedArea:   900,
				UsedHeight: 30,
				UsedWidth:  30,
			},
			wantIsScoreBetter: false,
		},
		{
			name: "best candidate usedHeight bigger than candidate usedHeight",
			candidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 30,
				UsedWidth:  40,
			},
			bestCandidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 60,
				UsedWidth:  20,
			},
			wantIsScoreBetter: true,
		},
		{
			name: "best candidate usedHeight smaller than candidate usedHeight",
			candidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 60,
				UsedWidth:  20,
			},
			bestCandidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 30,
				UsedWidth:  40,
			},
			wantIsScoreBetter: false,
		},
		{
			name: "equals candidates",
			candidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 60,
				UsedWidth:  20,
			},
			bestCandidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 60,
				UsedWidth:  20,
			},
			wantIsScoreBetter: false,
		},
		{
			name: "areas are approximately equal",
			candidateScore: CandidateScore{
				UsedArea:   1200.0000005,
				UsedHeight: 30,
				UsedWidth:  40,
			},
			bestCandidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 40,
				UsedWidth:  30,
			},
			wantIsScoreBetter: true,
		},
		{
			name: "heights are approximately equal",
			candidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 30.00000001,
				UsedWidth:  40,
			},
			bestCandidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 30,
				UsedWidth:  40,
			},
			wantIsScoreBetter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isScoreBetter := isScoreBetter(
				tt.candidateScore,
				tt.bestCandidateScore,
			)

			assert.Equal(
				t,
				tt.wantIsScoreBetter,
				isScoreBetter,
			)

		})
	}
}
