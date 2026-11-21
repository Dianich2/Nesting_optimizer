package simplefeatures

import (
	"math"
	"math/big"
	"slices"

	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

// These are per-call work limits, not geometric tolerances. The interface has no
// context; callers can cancel between bounded calls. Exhaustion is an error,
// never evidence that polygons are disjoint. No shared counters/caches exist.
const strictWorkLimit = 1_000_000
const strictExactLimit = 4096

type strictState struct {
	work, exact int
	err         error
}

func (s *strictState) step() bool {
	s.work++
	if s.work > strictWorkLimit {
		s.err = geometry.ErrComplexity
	}
	return s.err == nil
}
func (s *strictState) exactStep() bool {
	s.exact++
	if s.exact > strictExactLimit {
		s.err = geometry.ErrComplexity
	}
	return s.err == nil
}
func rat(x float64) *big.Rat { return new(big.Rat).SetFloat64(x) }
func sign(x float64) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// orient uses the orient2d first-stage error bound (3+16u)u, u=2^-53.
// A certified sign is exact, not an epsilon comparison. Overflow, underflow,
// cancellation and collinearity go to rational arithmetic on ORIGINAL doubles.
func (s *strictState) orient(a, b, c dg.Point) int {
	// Axis-aligned determinants factor into signs of original comparisons.
	// This is exact even for subnormals/overflow and makes shared rectangular
	// edges cheap without treating near-collinear edges as collinear.
	cmp := func(x, y float64) int {
		if x > y {
			return 1
		}
		if x < y {
			return -1
		}
		return 0
	}
	if a == c || b == c || a == b {
		return 0
	}
	if a.X == b.X {
		return cmp(a.X, c.X) * cmp(b.Y, a.Y)
	}
	if a.Y == b.Y {
		return -cmp(a.Y, c.Y) * cmp(b.X, a.X)
	}
	ax, ay, bx, by := a.X-c.X, a.Y-c.Y, b.X-c.X, b.Y-c.Y
	left, right := ax*by, ay*bx
	det := left - right
	normalProduct := func(p, x, y float64) bool {
		return (p == 0 && (x == 0 || y == 0)) || math.Abs(p) >= 0x1p-1022
	}
	bound := 3.3306690738754716e-16 * (math.Abs(left) + math.Abs(right))
	if normalProduct(left, ax, by) && normalProduct(right, ay, bx) &&
		math.Abs(det) >= 0x1p-1022 && !math.IsInf(bound, 0) && math.Abs(det) > bound {
		return sign(det)
	}
	if !s.exactStep() {
		return 0
	}
	x1 := new(big.Rat).Sub(rat(a.X), rat(c.X))
	y1 := new(big.Rat).Sub(rat(a.Y), rat(c.Y))
	x2 := new(big.Rat).Sub(rat(b.X), rat(c.X))
	y2 := new(big.Rat).Sub(rat(b.Y), rat(c.Y))
	return new(big.Rat).Sub(new(big.Rat).Mul(x1, y2), new(big.Rat).Mul(y1, x2)).Sign()
}

type strictEdge struct {
	a, b dg.Point
	left int
}

func openRing(r dg.Ring) []dg.Point {
	p := r.Points
	for len(p) > 1 && p[0] == p[len(p)-1] {
		p = p[:len(p)-1]
	}
	return p
}
func (s *strictState) edges(p dg.Polygon) []strictEdge {
	n := len(p.Exterior.Points)
	for _, h := range p.Holes {
		n += len(h.Points)
	}
	out := make([]strictEdge, 0, n)
	add := func(r dg.Ring, hole bool) {
		points := openRing(r)
		if len(points) < 3 {
			s.err = geometry.ErrInvalidPolygon
			return
		}
		// A lexicographically extreme vertex of a simple valid ring is convex.
		// Remove zero-length edges when finding its neighbours (closed input allowed).
		k := 0
		for i, v := range points {
			if v.X < points[k].X || (v.X == points[k].X && v.Y < points[k].Y) {
				k = i
			}
		}
		prev, next := (k+len(points)-1)%len(points), (k+1)%len(points)
		for prev != k && points[prev] == points[k] {
			prev = (prev + len(points) - 1) % len(points)
		}
		for next != k && points[next] == points[k] {
			next = (next + 1) % len(points)
		}
		left := s.orient(points[prev], points[k], points[next])
		if left == 0 {
			if s.err == nil {
				s.err = geometry.ErrInvalidPolygon
			}
			return
		}
		if hole {
			left = -left
		}
		for i, a := range points {
			b := points[(i+1)%len(points)]
			if a != b {
				out = append(out, strictEdge{a, b, left})
			}
		}
	}
	add(p.Exterior, false)
	for _, h := range p.Holes {
		add(h, true)
	}
	return out
}
func between(x, a, b float64) bool { return min(a, b) <= x && x <= max(a, b) }
func onBox(p dg.Point, e strictEdge) bool {
	return between(p.X, e.a.X, e.b.X) && between(p.Y, e.a.Y, e.b.Y)
}
func boxesMeet(a, b strictEdge) bool {
	return max(a.a.X, a.b.X) >= min(b.a.X, b.b.X) && max(b.a.X, b.b.X) >= min(a.a.X, a.b.X) &&
		max(a.a.Y, a.b.Y) >= min(b.a.Y, b.b.Y) && max(b.a.Y, b.b.Y) >= min(a.a.Y, a.b.Y)
}
func strictBounds(p dg.Polygon) (x0, y0, x1, y1 float64) {
	x0, y0, x1, y1 = math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, v := range p.Exterior.Points {
		x0 = min(x0, v.X)
		y0 = min(y0, v.Y)
		x1 = max(x1, v.X)
		y1 = max(y1, v.Y)
	}
	return
}

// pointLocation returns -1 exterior, 0 boundary, 1 interior. Ray crossing uses
// half-open Y intervals and orientation only: no division/intersection rounding.
// Even-odd parity includes all rings, independent of winding.
func (s *strictState) pointLocation(p dg.Point, edges []strictEdge) int {
	inside := false
	for _, e := range edges {
		if !s.step() {
			return -1
		}
		if onBox(p, e) && s.orient(e.a, e.b, p) == 0 {
			return 0
		}
		if (e.a.Y > p.Y) != (e.b.Y > p.Y) {
			o := s.orient(e.a, e.b, p)
			if (e.b.Y > e.a.Y && o > 0) || (e.b.Y < e.a.Y && o < 0) {
				inside = !inside
			}
		}
	}
	if inside {
		return 1
	}
	return -1
}

// Exact midpoint cannot be rounded to a boundary endpoint, even at one-ULP
// spacing. Only used after a genuine vertex-on-edge contact, not on every pair.
func (s *strictState) midpointLocation(a, b dg.Point, edges []strictEdge) int {
	// Knuth TwoSum certifies whether the midpoint is representable, rather
	// than rounding a potentially one-ULP interval to an endpoint.
	midpoint := func(a, b float64) (float64, bool) {
		sum := a + b
		if math.IsInf(sum, 0) {
			return 0, false
		}
		bv := sum - a
		err := (a - (sum - bv)) + (b - bv)
		mid := sum / 2
		return mid, err == 0 && mid*2 == sum
	}
	xFloat, xExact := midpoint(a.X, b.X)
	yFloat, yExact := midpoint(a.Y, b.Y)
	if xExact && yExact {
		return s.pointLocation(dg.Point{X: xFloat, Y: yFloat}, edges)
	}
	if !s.exactStep() {
		return -1
	}
	two := big.NewRat(2, 1)
	x := new(big.Rat).Quo(new(big.Rat).Add(rat(a.X), rat(b.X)), two)
	y := new(big.Rat).Quo(new(big.Rat).Add(rat(a.Y), rat(b.Y)), two)
	inside := false
	for _, e := range edges {
		if !s.step() || !s.exactStep() {
			return -1
		}
		ax, ay, bx, by := rat(e.a.X), rat(e.a.Y), rat(e.b.X), rat(e.b.Y)
		dx, dy := new(big.Rat).Sub(bx, ax), new(big.Rat).Sub(by, ay)
		qx, qy := new(big.Rat).Sub(x, ax), new(big.Rat).Sub(y, ay)
		o := new(big.Rat).Sub(new(big.Rat).Mul(dx, qy), new(big.Rat).Mul(dy, qx)).Sign()
		if o == 0 && x.Cmp(rat(min(e.a.X, e.b.X))) >= 0 && x.Cmp(rat(max(e.a.X, e.b.X))) <= 0 &&
			y.Cmp(rat(min(e.a.Y, e.b.Y))) >= 0 && y.Cmp(rat(max(e.a.Y, e.b.Y))) <= 0 {
			return 0
		}
		aboveA, aboveB := ay.Cmp(y) > 0, by.Cmp(y) > 0
		if aboveA != aboveB && ((e.b.Y > e.a.Y && o > 0) || (e.b.Y < e.a.Y && o < 0)) {
			inside = !inside
		}
	}
	if inside {
		return 1
	}
	return -1
}

// Once proper crossings are excluded, an edge can enter/leave the other polygon
// only at its vertices. Split at those original (exact float64) vertices and
// classify open subsegments by exact midpoints. This covers crossings through
// shared vertices and T junctions, which vertex containment alone misses.
func (s *strictState) contactIntervals(edges, other []strictEdge) bool {
	return s.intervalsHaveLocation(edges, other, 1)
}

func (s *strictState) intervalsHaveLocation(edges, other []strictEdge, location int) bool {
	for _, e := range edges {
		cuts := []dg.Point{e.a, e.b}
		for _, f := range other {
			if !s.step() {
				return false
			}
			if f.a != e.a && f.a != e.b && onBox(f.a, e) && s.orient(e.a, e.b, f.a) == 0 {
				cuts = append(cuts, f.a)
			}
		}
		if len(cuts) == 2 {
			if s.pointLocation(e.a, other) != 0 || s.pointLocation(e.b, other) != 0 {
				if s.err != nil {
					return false
				}
				continue
			}
		}
		slices.SortFunc(cuts, func(a, b dg.Point) int {
			if a.X < b.X || (a.X == b.X && a.Y < b.Y) {
				return -1
			}
			if a == b {
				return 0
			}
			return 1
		})
		cuts = slices.Compact(cuts)
		for i := 1; i < len(cuts); i++ {
			if s.midpointLocation(cuts[i-1], cuts[i], other) == location {
				return true
			}
			if s.err != nil {
				return false
			}
		}
	}
	return false
}

// Complete for valid polygon boundaries: positive overlap has a proper boundary
// crossing, boundary in the other's interior, or coincident boundary intervals
// with the interiors on the same side. All three cases are checked symmetrically.
func strictInteriors(first, second dg.Polygon, s *strictState) (bool, error) {
	x0, y0, x1, y1 := strictBounds(first)
	u0, v0, u1, v1 := strictBounds(second)
	if x1 <= u0 || u1 <= x0 || y1 <= v0 || v1 <= y0 {
		return false, nil
	}
	a, b := s.edges(first), s.edges(second)
	if s.err != nil {
		return false, s.err
	}
	contact := false
	for _, e := range a {
		for _, f := range b {
			if !s.step() {
				return false, s.err
			}
			if !boxesMeet(e, f) {
				continue
			}
			o1, o2 := s.orient(e.a, e.b, f.a), s.orient(e.a, e.b, f.b)
			o3, o4 := s.orient(f.a, f.b, e.a), s.orient(f.a, f.b, e.b)
			if s.err != nil {
				return false, s.err
			}
			if o1*o2 < 0 && o3*o4 < 0 {
				return true, nil
			}
			if o1 == 0 && o2 == 0 {
				// Strict positive length of collinear overlap; opposite interior sides
				// are legal edge-only touch (including a pattern on a hole boundary).
				sameDirection := false
				overlap := false
				if e.a.X != e.b.X {
					overlap = max(min(e.a.X, e.b.X), min(f.a.X, f.b.X)) < min(max(e.a.X, e.b.X), max(f.a.X, f.b.X))
					sameDirection = (e.b.X > e.a.X) == (f.b.X > f.a.X)
				} else {
					overlap = max(min(e.a.Y, e.b.Y), min(f.a.Y, f.b.Y)) < min(max(e.a.Y, e.b.Y), max(f.a.Y, f.b.Y))
					sameDirection = (e.b.Y > e.a.Y) == (f.b.Y > f.a.Y)
				}
				if overlap && ((e.left == f.left) == sameDirection) {
					return true, nil
				}
			}
			contact = contact || (o1 == 0 && onBox(f.a, e)) || (o2 == 0 && onBox(f.b, e)) || (o3 == 0 && onBox(e.a, f)) || (o4 == 0 && onBox(e.b, f))
		}
	}
	for _, pair := range [][2][]strictEdge{{a, b}, {b, a}} {
		for _, e := range pair[0] {
			if s.pointLocation(e.a, pair[1]) == 1 {
				return true, s.err
			}
			if s.err != nil {
				return false, s.err
			}
		}
	}
	if contact {
		if s.contactIntervals(a, b) || s.contactIntervals(b, a) {
			return true, s.err
		}
	}
	return false, s.err
}
