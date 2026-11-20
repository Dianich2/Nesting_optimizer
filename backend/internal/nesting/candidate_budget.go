package nesting

import (
	"context"
	"server_nesting_optimizer/internal/geometry"
)

// MaxCandidates bounds BOTH raw enumeration work and stored candidates per
// placement attempt (all rotations + fallback). Duplicates consume work too.
// Over-budget inputs fail explicitly; no truncated candidate list is selected.
const MaxCandidates = 200000

func candidateProduct(limit int, factors ...int) (int, error) {
	if limit < 0 {
		return 0, geometry.ErrComplexity
	}
	for _, n := range factors {
		if n < 0 {
			return 0, geometry.ErrComplexity
		}
	}
	for _, n := range factors {
		if n == 0 {
			return 0, nil
		}
	}
	product := 1
	for _, n := range factors {
		if product > limit/n {
			return 0, geometry.ErrComplexity
		}
		product *= n
	}
	if product > limit {
		return 0, geometry.ErrComplexity
	}
	return product, nil
}

type candidateBuffer struct {
	items       []Candidate
	seen        map[Candidate]struct{}
	work, limit int
}

func newCandidateBuffer(limit int, dedup bool) *candidateBuffer {
	b := &candidateBuffer{limit: limit}
	if dedup {
		b.seen = make(map[Candidate]struct{})
	}
	return b
}
func (b *candidateBuffer) add(ctx context.Context, c Candidate) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if b.work >= b.limit {
		return geometry.ErrComplexity
	}
	b.work++
	if b.seen != nil {
		if _, ok := b.seen[c]; ok {
			return nil
		}
		b.seen[c] = struct{}{}
	}
	// Explicit growth caps capacity too, avoiding append's over-allocation beyond
	// the limit. Old backing arrays become unreachable after each bounded copy.
	if len(b.items) == cap(b.items) {
		capacity := max(16, cap(b.items)*2)
		capacity = min(capacity, b.limit)
		next := make([]Candidate, len(b.items), capacity)
		copy(next, b.items)
		b.items = next
	}
	b.items = append(b.items, c)
	return nil
}
