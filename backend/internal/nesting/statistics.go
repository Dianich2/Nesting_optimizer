package nesting

import (
	"context"
	"sync/atomic"
)

// CandidateStatistics is optional per-execution instrumentation. Normal runs do
// not collect it. Benchmarks use the same observer before and after a change.
type CandidateStatistics struct{ Generated atomic.Uint64 }
type candidateStatisticsKey struct{}

func WithCandidateStatistics(ctx context.Context, statistics *CandidateStatistics) context.Context {
	if statistics == nil {
		return ctx
	}
	return context.WithValue(ctx, candidateStatisticsKey{}, statistics)
}
func recordCandidateStatistics(ctx context.Context, count int) {
	if statistics, ok := ctx.Value(candidateStatisticsKey{}).(*CandidateStatistics); ok {
		statistics.Generated.Add(uint64(count))
	}
}
