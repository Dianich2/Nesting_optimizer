package nesting

import (
	"context"
	"math/rand"
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
)

const benchmarkSeed int64 = 20260927

type benchmarkRandom struct {
	rng *rand.Rand
}

func newBenchmarkRandom(seed int64) *benchmarkRandom {
	return &benchmarkRandom{
		rng: rand.New(
			rand.NewSource(seed),
		),
	}
}

func (r *benchmarkRandom) Intn(n int) int {
	return r.rng.Intn(n)
}

func (r *benchmarkRandom) Shuffle(
	n int,
	swap func(i int, j int),
) {
	r.rng.Shuffle(n, swap)
}

func (r *benchmarkRandom) Float64() float64 {
	return r.rng.Float64()
}

type benchmarkProblem struct {
	name    string
	problem Problem
}

type benchmarkProfile struct {
	geneticConfig    GeneticConfig
	annealingConfig  AnnealingConfig
	compactionConfig CompactionConfig
}

func productionBenchmarkProfile() benchmarkProfile {
	return benchmarkProfile{
		geneticConfig:    DefaultGeneticConfig(),
		annealingConfig:  DefaultAnnealingConfig(),
		compactionConfig: DefaultCompactionConfig(),
	}
}

func comparisonBenchmarkProfile() benchmarkProfile {
	return benchmarkProfile{
		geneticConfig: GeneticConfig{
			PopulationSize: 8,
			Generations:    10,
			EliteCount:     1,
			TournamentSize: 2,
			MutationRate:   0.2,
		},
		annealingConfig: AnnealingConfig{
			InitialTemperature: 1,
			MinTemperature:     0.01,
			CoolingRate:        0.95,
			MaxIterations:      100,
		},
		compactionConfig: DefaultCompactionConfig(),
	}
}

func BenchmarkOptimizers(b *testing.B) {
	runOptimizerBenchmarks(
		b,
		productionBenchmarkProfile(),
	)
}

func BenchmarkOptimizersComparison(b *testing.B) {
	runOptimizerBenchmarks(
		b,
		comparisonBenchmarkProfile(),
	)
}

func runOptimizerBenchmarks(
	b *testing.B,
	profile benchmarkProfile,
) {
	problems := benchmarkProblems()

	algorithms := []Algorithm{
		BaselineAlgorithm,
		NFPGreedyAlgorithm,
		GeneticAlgorithm,
		SimulatedAnnealingAlgorithm,
		HybridAlgorithm,
	}

	for _, benchmarkCase := range problems {
		benchmarkCase := benchmarkCase

		b.Run(benchmarkCase.name, func(b *testing.B) {
			for _, algorithm := range algorithms {
				algorithm := algorithm

				b.Run(string(algorithm), func(b *testing.B) {
					engine := simplefeatures.NewEngine()

					optimizer := newBenchmarkOptimizer(
						b,
						engine,
						algorithm,
						profile,
					)

					ctx := context.Background()

					var totalPlaced float64
					var totalUnplaced float64
					var totalUtilization float64
					var totalCompactness float64
					var totalUsedArea float64
					var totalUsedHeight float64

					b.ReportAllocs()
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						result, err := optimizer.Optimize(
							ctx,
							benchmarkCase.problem,
						)
						if err != nil {
							b.Fatalf(
								"optimize with %s: %v",
								algorithm,
								err,
							)
						}

						b.StopTimer()

						fitness, err := calculatePlacementsFitness(
							engine,
							benchmarkCase.problem,
							result.Placements,
						)
						if err != nil {
							b.Fatalf(
								"calculate fitness for %s: %v",
								algorithm,
								err,
							)
						}

						if fitness.UsedArea > 0 {
							totalCompactness +=
								result.Metrics.PlacedArea /
									fitness.UsedArea
						}

						totalPlaced += float64(
							result.Metrics.PlacedCount,
						)

						totalUnplaced += float64(
							fitness.UnplacedCount,
						)

						totalUtilization +=
							result.Metrics.Utilization

						totalUsedArea +=
							fitness.UsedArea

						totalUsedHeight +=
							fitness.UsedHeight

						b.StartTimer()
					}

					b.StopTimer()

					runs := float64(b.N)

					b.ReportMetric(
						totalPlaced/runs,
						"placed/op",
					)

					b.ReportMetric(
						totalUnplaced/runs,
						"unplaced/op",
					)

					b.ReportMetric(
						(totalUtilization/runs)*100,
						"util_pct/op",
					)

					b.ReportMetric(
						(totalCompactness/runs)*100,
						"compact_pct/op",
					)

					b.ReportMetric(
						totalUsedArea/runs,
						"used_area/op",
					)

					b.ReportMetric(
						totalUsedHeight/runs,
						"used_height/op",
					)
				})
			}
		})
	}
}

func newBenchmarkOptimizer(
	b *testing.B,
	engine geometry.Engine,
	algorithm Algorithm,
	profile benchmarkProfile,
) Optimizer {
	b.Helper()

	switch algorithm {
	case BaselineAlgorithm:
		return NewBaselineOptimizer(
			engine,
		)

	case NFPGreedyAlgorithm:
		nfpBuilder := nfp.NewBuilder(
			engine,
		)

		return NewNFPGreedyOptimizer(
			engine,
			nfpBuilder,
		)

	case GeneticAlgorithm:
		nfpBuilder := nfp.NewBuilder(
			engine,
		)

		optimizer, err := NewGeneticOptimizer(
			engine,
			nfpBuilder,
			profile.geneticConfig,
		)
		if err != nil {
			b.Fatalf(
				"create genetic optimizer: %v",
				err,
			)
		}

		optimizer.rng = newBenchmarkRandom(
			benchmarkSeed + 1,
		)

		return optimizer

	case SimulatedAnnealingAlgorithm:
		nfpBuilder := nfp.NewBuilder(
			engine,
		)

		optimizer, err :=
			NewSimulatedAnnealingOptimizer(
				engine,
				nfpBuilder,
				profile.annealingConfig,
			)
		if err != nil {
			b.Fatalf(
				"create simulated annealing optimizer: %v",
				err,
			)
		}

		optimizer.rng = newBenchmarkRandom(
			benchmarkSeed + 2,
		)

		return optimizer

	case HybridAlgorithm:
		nfpBuilder := nfp.NewBuilder(
			engine,
		)

		geneticOptimizer, err :=
			NewGeneticOptimizer(
				engine,
				nfpBuilder,
				profile.geneticConfig,
			)
		if err != nil {
			b.Fatalf(
				"create hybrid genetic optimizer: %v",
				err,
			)
		}

		geneticOptimizer.rng =
			newBenchmarkRandom(
				benchmarkSeed + 3,
			)

		annealingOptimizer, err :=
			NewSimulatedAnnealingOptimizer(
				engine,
				nfpBuilder,
				profile.annealingConfig,
			)
		if err != nil {
			b.Fatalf(
				"create hybrid annealing optimizer: %v",
				err,
			)
		}

		annealingOptimizer.rng =
			newBenchmarkRandom(
				benchmarkSeed + 4,
			)

		hybridOptimizer, err :=
			NewHybridOptimizer(
				engine,
				geneticOptimizer,
				annealingOptimizer,
				profile.compactionConfig,
			)
		if err != nil {
			b.Fatalf(
				"create hybrid optimizer: %v",
				err,
			)
		}

		return hybridOptimizer

	default:
		b.Fatalf(
			"unsupported benchmark algorithm: %s",
			algorithm,
		)

		return nil
	}
}

func benchmarkProblems() []benchmarkProblem {
	return []benchmarkProblem{
		{
			name: "simple_control",
			problem: Problem{
				Surface: benchmarkRectangle(
					100,
					100,
				),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  4,
						Geometry: benchmarkRectangle(
							20,
							20,
						),
					},
					{
						PatternID: 2,
						Quantity:  2,
						Geometry: benchmarkRectangle(
							30,
							20,
						),
					},
				},
				AllowedRotations: []float64{
					0,
					90,
				},
			},
		},
		{
			name: "mixed_polygons",
			problem: Problem{
				Surface: benchmarkRectangle(
					140,
					100,
				),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  3,
						Geometry:  benchmarkLShape(),
					},
					{
						PatternID: 2,
						Quantity:  3,
						Geometry: benchmarkTriangle(
							30,
							25,
						),
					},
					{
						PatternID: 3,
						Quantity:  3,
						Geometry: benchmarkRectangle(
							25,
							15,
						),
					},
				},
				AllowedRotations: []float64{
					0,
					90,
					180,
					270,
				},
			},
		},
		{
			name: "with_obstacle",
			problem: Problem{
				Surface: benchmarkRectangle(
					120,
					100,
				),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  4,
						Geometry: benchmarkRectangle(
							30,
							20,
						),
					},
					{
						PatternID: 2,
						Quantity:  2,
						Geometry:  benchmarkLShape(),
					},
				},
				Obstacles: []domaingeometry.Polygon{
					benchmarkRectangleAt(
						40,
						30,
						40,
						40,
					),
				},
				AllowedRotations: []float64{
					0,
					90,
					180,
					270,
				},
			},
		},
		{
			name: "constrained_space",
			problem: Problem{
				Surface: benchmarkRectangle(
					80,
					60,
				),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  4,
						Geometry: benchmarkRectangle(
							30,
							20,
						),
					},
					{
						PatternID: 2,
						Quantity:  4,
						Geometry: benchmarkTriangle(
							25,
							20,
						),
					},
					{
						PatternID: 3,
						Quantity:  3,
						Geometry:  benchmarkLShape(),
					},
				},
				AllowedRotations: []float64{
					0,
					90,
					180,
					270,
				},
			},
		},
	}
}

func benchmarkRectangle(
	width float64,
	height float64,
) domaingeometry.Polygon {
	return benchmarkRectangleAt(
		0,
		0,
		width,
		height,
	)
}

func benchmarkRectangleAt(
	x float64,
	y float64,
	width float64,
	height float64,
) domaingeometry.Polygon {
	return domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{
					X: x,
					Y: y,
				},
				{
					X: x + width,
					Y: y,
				},
				{
					X: x + width,
					Y: y + height,
				},
				{
					X: x,
					Y: y + height,
				},
			},
		},
	}
}

func benchmarkTriangle(
	width float64,
	height float64,
) domaingeometry.Polygon {
	return domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: width,
					Y: 0,
				},
				{
					X: width / 2,
					Y: height,
				},
			},
		},
	}
}

func benchmarkLShape() domaingeometry.Polygon {
	return domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 30,
					Y: 0,
				},
				{
					X: 30,
					Y: 10,
				},
				{
					X: 10,
					Y: 10,
				},
				{
					X: 10,
					Y: 30,
				},
				{
					X: 0,
					Y: 30,
				},
			},
		},
	}
}
