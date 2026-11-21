package simplefeatures

import dg "server_nesting_optimizer/internal/domain/geometry"

// Valid polygons are regular closed sets. Coverage fails exactly when some
// inner exterior boundary lies outside outer, or inner interiors enter an
// outer hole. No rounded intersection coordinates are constructed.
func strictCoveredBy(inner, outer dg.Polygon, s *strictState) (bool, error) {
	x0, y0, x1, y1 := strictBounds(inner)
	u0, v0, u1, v1 := strictBounds(outer)
	if x0 < u0 || y0 < v0 || x1 > u1 || y1 > v1 {
		return false, nil
	}
	a, b := s.edges(dg.Polygon{Exterior: inner.Exterior}), s.edges(outer)
	if s.err != nil {
		return false, s.err
	}
	for _, e := range a {
		if s.pointLocation(e.a, b) < 0 {
			return false, s.err
		}
		if s.err != nil {
			return false, s.err
		}
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
				return false, nil
			}
			contact = contact || (o1 == 0 && onBox(f.a, e)) || (o2 == 0 && onBox(f.b, e)) || (o3 == 0 && onBox(e.a, f)) || (o4 == 0 && onBox(e.b, f))
		}
	}
	if contact && s.intervalsHaveLocation(a, b, -1) {
		return false, s.err
	}
	if s.err != nil {
		return false, s.err
	}
	for _, h := range outer.Holes {
		overlap, err := strictInteriors(inner, dg.Polygon{Exterior: h}, s)
		if err != nil || overlap {
			return false, err
		}
	}
	return true, s.err
}
