package nesting

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
)

type strictTransitionPair struct {
	First, Second dg.Polygon
	Old, Strict   bool
}
type strictTransitionEngine struct {
	*sf.Engine
	pairs  []strictTransitionPair
	legacy bool
}

func legacyInteriors(a, b dg.Polygon) (bool, error) {
	codec := sf.NewCodec()
	wa, err := codec.EncodeWKB(a)
	if err != nil {
		return false, err
	}
	wb, err := codec.EncodeWKB(b)
	if err != nil {
		return false, err
	}
	ga, err := geom.UnmarshalWKB(wa)
	if err != nil {
		return false, err
	}
	gb, err := geom.UnmarshalWKB(wb)
	if err != nil {
		return false, err
	}
	relation, err := geom.Relate(ga, gb)
	if err != nil {
		return false, err
	}
	return geom.RelateMatches(relation, "2********")
}
func (e *strictTransitionEngine) InteriorsIntersect(a, b dg.Polygon) (bool, error) {
	if e.legacy {
		return legacyInteriors(a, b)
	}
	strict, err := e.Engine.InteriorsIntersect(a, b)
	if err != nil {
		return false, err
	}
	old, err := legacyInteriors(a, b)
	if err != nil {
		return false, err
	}
	if old != strict && len(e.pairs) < 256 {
		e.pairs = append(e.pairs, strictTransitionPair{a, b, old, strict})
	}
	return strict, nil
}
func TestStrictHybridTransitionEvidence(t *testing.T) {
	path := os.Getenv("NESTING_STRICT_TRANSITION")
	if path == "" {
		t.Skip("explicit bounded before/after evidence")
	}
	var problem Problem
	for _, c := range benchmarkProblems() {
		if c.name == "constrained_space" {
			problem = c.problem
		}
	}
	engine := &strictTransitionEngine{Engine: sf.NewEngine()}
	var e geometry.Engine = engine
	builder := newExactNFPDeterminismBuilder(nfp.NewBuilder(e))
	profile := comparisonBenchmarkProfile()
	ga, err := NewGeneticOptimizer(e, builder, profile.geneticConfig)
	if err != nil {
		t.Fatal(err)
	}
	ga.rng = newBenchmarkRandom(benchmarkSeed + 3)
	sa, err := NewSimulatedAnnealingOptimizer(e, builder, profile.annealingConfig)
	if err != nil {
		t.Fatal(err)
	}
	sa.rng = newBenchmarkRandom(benchmarkSeed + 4)
	hybrid, err := NewHybridOptimizer(e, ga, sa, profile.compactionConfig)
	if err != nil {
		t.Fatal(err)
	}
	result, err := hybrid.Optimize(context.Background(), problem)
	if err != nil {
		t.Fatal(err)
	}
	formed := checkInvariantResult(t, problem, result)
	data, err := json.Marshal(struct {
		Pairs []strictTransitionPair
		Rows  []invariantCase
	}{engine.pairs, []invariantCase{{Name: "hybrid-strict-transition", Algorithm: HybridAlgorithm, Problem: problem, Result: result, Formed: formed}}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("TRANSITION pairs=%d total=%d repeated=%d mismatches=%d placed=%d", len(engine.pairs), builder.total, builder.repeated, builder.mismatches, len(result.Placements))
}
