package nesting

import (
	"fmt"
	"math"
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/peterstace/simplefeatures/geom"
)

type nfpReference struct {
	anchor domaingeometry.Point
	result domaingeometry.MultiPolygon
}

type checkingNFPBuilder struct {
	delegate geometry.NFPBuilder
	first    map[[32]byte]nfpReference

	comparisons              uint64
	mismatches               uint64
	maxDiffArea              float64
	firstError               string
	vertexMismatches         uint64
	firstVertexDetails       string
	candidateComparisons     uint64
	candidateSetMismatches   uint64
	candidateOrderMismatches uint64
}

var _ geometry.NFPBuilder = (*checkingNFPBuilder)(nil)

func newCheckingNFPBuilder(
	delegate geometry.NFPBuilder,
) *checkingNFPBuilder {
	return &checkingNFPBuilder{
		delegate: delegate,
		first:    make(map[[32]byte]nfpReference),
	}
}

func (b *checkingNFPBuilder) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	// Build the actual NFP every time.
	actual, err := b.delegate.BuildExternal(stationary, moving)
	if err != nil {
		return actual, err
	}

	key := diagnosticNFPKey(stationary, moving)

	anchor := stationary.Exterior.Points[0]

	reference, exists := b.first[key]
	if !exists {
		b.first[key] = nfpReference{
			anchor: anchor,
			result: actual,
		}
		return actual, nil
	}

	b.comparisons++

	dx := anchor.X - reference.anchor.X
	dy := anchor.Y - reference.anchor.Y

	expected := translateNFPMultiPolygon(
		reference.result,
		dx,
		dy,
	)

	expectedCandidates := nfpCandidatesForComparison(
		expected,
		moving,
	)

	actualCandidates := nfpCandidatesForComparison(
		actual,
		moving,
	)

	b.candidateComparisons++

	if !sameCandidateSet(
		expectedCandidates,
		actualCandidates,
	) {
		b.candidateSetMismatches++
	}

	if !sameCandidateSequence(
		expectedCandidates,
		actualCandidates,
	) {
		b.candidateOrderMismatches++
	}

	const vertexTolerance = 1e-8

	if err := compareNFPVertices(
		expected,
		actual,
		vertexTolerance,
	); err != nil {
		b.vertexMismatches++

		if b.firstVertexDetails == "" {
			b.firstVertexDetails = fmt.Sprintf(
				"comparison: %d\n"+
					"reason: %v\n"+
					"translation: dx=%.17g, dy=%.17g\n"+
					"expected NFP: %+v\n"+
					"actual NFP: %+v",
				b.comparisons,
				err,
				dx,
				dy,
				expected,
				actual,
			)
		}

		if b.firstError == "" {
			b.firstError = fmt.Sprintf(
				"comparison %d: vertex mismatch: %v",
				b.comparisons,
				err,
			)
		}
	}

	diffArea, err := nfpSymmetricDifferenceArea(
		expected,
		actual,
	)
	if err != nil {
		b.mismatches++
		if b.firstError == "" {
			b.firstError = fmt.Sprintf(
				"comparison %d: %v",
				b.comparisons,
				err,
			)
		}
		return actual, nil
	}

	if diffArea > b.maxDiffArea {
		b.maxDiffArea = diffArea
	}

	// Diagnostic tolerance, not a production cache rule.
	const tolerance = 1e-7

	if math.IsNaN(diffArea) ||
		math.IsInf(diffArea, 0) ||
		diffArea > tolerance {
		b.mismatches++

		if b.firstError == "" {
			b.firstError = fmt.Sprintf(
				"comparison %d: symmetric difference area %.12g",
				b.comparisons,
				diffArea,
			)
		}
	}

	return actual, nil
}

func translateNFPMultiPolygon(
	multi domaingeometry.MultiPolygon,
	dx, dy float64,
) domaingeometry.MultiPolygon {
	result := domaingeometry.MultiPolygon{
		Polygons: make(
			[]domaingeometry.Polygon,
			len(multi.Polygons),
		),
	}

	translateRing := func(
		ring domaingeometry.Ring,
	) domaingeometry.Ring {
		translated := domaingeometry.Ring{
			Points: make(
				[]domaingeometry.Point,
				len(ring.Points),
			),
		}

		for i, point := range ring.Points {
			translated.Points[i] = domaingeometry.Point{
				X: point.X + dx,
				Y: point.Y + dy,
			}
		}

		return translated
	}

	for i, polygon := range multi.Polygons {
		result.Polygons[i].Exterior = translateRing(
			polygon.Exterior,
		)

		result.Polygons[i].Holes = make(
			[]domaingeometry.Ring,
			len(polygon.Holes),
		)

		for j, hole := range polygon.Holes {
			result.Polygons[i].Holes[j] = translateRing(hole)
		}
	}

	return result
}

func nfpSymmetricDifferenceArea(
	expected domaingeometry.MultiPolygon,
	actual domaingeometry.MultiPolygon,
) (float64, error) {
	codec := simplefeatures.NewCodec()

	convert := func(
		multi domaingeometry.MultiPolygon,
	) (geom.Geometry, error) {
		polygons := make([]geom.Polygon, 0, len(multi.Polygons))

		for _, polygon := range multi.Polygons {
			data, err := codec.EncodeWKB(polygon)
			if err != nil {
				return geom.Geometry{}, err
			}

			decoded, err := geom.UnmarshalWKB(data)
			if err != nil {
				return geom.Geometry{}, err
			}

			sfPolygon, ok := decoded.AsPolygon()
			if !ok {
				return geom.Geometry{}, fmt.Errorf(
					"decoded geometry is not polygon",
				)
			}

			polygons = append(polygons, sfPolygon)
		}

		return geom.NewMultiPolygon(polygons).AsGeometry(), nil
	}

	expectedGeom, err := convert(expected)
	if err != nil {
		return 0, err
	}

	actualGeom, err := convert(actual)
	if err != nil {
		return 0, err
	}

	difference, err := geom.SymmetricDifference(
		expectedGeom,
		actualGeom,
	)
	if err != nil {
		return 0, err
	}

	return difference.Area(), nil
}

func (b *checkingNFPBuilder) Report(t *testing.T) {
	t.Helper()

	t.Logf("NFP comparisons: %d", b.comparisons)
	t.Logf("NFP mismatches: %d", b.mismatches)
	t.Logf("Maximum symmetric difference area: %.12g", b.maxDiffArea)

	t.Logf(
		"NFP candidate comparisons: %d",
		b.candidateComparisons,
	)

	t.Logf(
		"NFP candidate set mismatches: %d",
		b.candidateSetMismatches,
	)

	t.Logf(
		"NFP candidate sequence mismatches: %d",
		b.candidateOrderMismatches,
	)

	if b.firstError != "" {
		t.Logf("First mismatch: %s", b.firstError)
	}

	if b.comparisons == 0 {
		t.Fatal("no repeated NFP pairs were checked")
	}

	if b.mismatches > 0 {
		t.Fatalf(
			"NFP equivalence failed for %d of %d comparisons",
			b.mismatches,
			b.comparisons,
		)
	}

	t.Logf(
		"NFP vertex mismatches: %d",
		b.vertexMismatches,
	)

	if b.firstVertexDetails != "" {
		t.Logf(
			"First vertex mismatch details:\n%s",
			b.firstVertexDetails,
		)
	}

	if b.vertexMismatches > 0 {
		t.Logf(
			"Relative NFP diagnostic: %d vertex mismatches; "+
				"translated NFP results must not be used as "+
				"an exact replacement for BuildExternal",
			b.vertexMismatches,
		)
	}
}

func compareNFPVertices(
	expected domaingeometry.MultiPolygon,
	actual domaingeometry.MultiPolygon,
	tolerance float64,
) error {
	if len(expected.Polygons) != len(actual.Polygons) {
		return fmt.Errorf(
			"polygon count differs: expected %d, actual %d",
			len(expected.Polygons),
			len(actual.Polygons),
		)
	}

	type vertex struct {
		x float64
		y float64
	}

	collect := func(multi domaingeometry.MultiPolygon) (
		exteriors []vertex,
		holes []vertex,
		holeCount int,
	) {
		for _, polygon := range multi.Polygons {
			for _, point := range polygon.Exterior.Points {
				exteriors = append(exteriors, vertex{point.X, point.Y})
			}

			holeCount += len(polygon.Holes)

			for _, hole := range polygon.Holes {
				for _, point := range hole.Points {
					holes = append(holes, vertex{point.X, point.Y})
				}
			}
		}

		return exteriors, holes, holeCount
	}

	expectedExterior, expectedHoles, expectedHoleCount := collect(expected)
	actualExterior, actualHoles, actualHoleCount := collect(actual)

	if expectedHoleCount != actualHoleCount {
		return fmt.Errorf(
			"hole count differs: expected %d, actual %d",
			expectedHoleCount,
			actualHoleCount,
		)
	}

	compare := func(name string, first, second []vertex) error {
		if len(first) != len(second) {
			return fmt.Errorf(
				"%s vertex count differs: expected %d, actual %d",
				name, len(first), len(second),
			)
		}

		used := make([]bool, len(second))

		for _, a := range first {
			matched := false

			for j, b := range second {
				if used[j] {
					continue
				}

				if math.Abs(a.x-b.x) <= tolerance &&
					math.Abs(a.y-b.y) <= tolerance {
					used[j] = true
					matched = true
					break
				}
			}

			if !matched {
				return fmt.Errorf(
					"%s vertex (%.15g, %.15g) has no match",
					name, a.x, a.y,
				)
			}
		}

		return nil
	}

	if err := compare(
		"exterior",
		expectedExterior,
		actualExterior,
	); err != nil {
		return err
	}

	return compare("hole", expectedHoles, actualHoles)
}

func nfpCandidatesForComparison(
	multi domaingeometry.MultiPolygon,
	moving domaingeometry.Polygon,
) []Candidate {
	vertices := ringVertices(moving.Exterior)

	minX, maxX := vertices[0].X, vertices[0].X
	minY, maxY := vertices[0].Y, vertices[0].Y

	for _, point := range vertices[1:] {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	centerX := minX/2 + maxX/2
	centerY := minY/2 + maxY/2

	candidates := make([]Candidate, 0)
	seen := make(map[Candidate]struct{})

	addRing := func(ring domaingeometry.Ring) {
		for _, point := range ring.Points {
			candidate := Candidate{
				X: point.X - centerX,
				Y: point.Y - centerY,
			}

			if _, exists := seen[candidate]; exists {
				continue
			}

			seen[candidate] = struct{}{}
			candidates = append(candidates, candidate)
		}
	}

	for _, polygon := range multi.Polygons {
		addRing(polygon.Exterior)

		for _, hole := range polygon.Holes {
			addRing(hole)
		}
	}

	return candidates
}

func sameCandidateSequence(
	first []Candidate,
	second []Candidate,
) bool {
	if len(first) != len(second) {
		return false
	}

	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}

	return true
}

func sameCandidateSet(
	first []Candidate,
	second []Candidate,
) bool {
	if len(first) != len(second) {
		return false
	}

	seen := make(map[Candidate]struct{}, len(first))

	for _, candidate := range first {
		seen[candidate] = struct{}{}
	}

	for _, candidate := range second {
		if _, exists := seen[candidate]; !exists {
			return false
		}
	}

	return true
}
