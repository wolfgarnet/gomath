package gomath

// This file rounds the Polygon type out into a production-grade primitive for
// games: area/centroid, point containment, bounds, perimeter and safe copies.
// All additions are value-receiver and non-mutating.
//
// Shoelace-based quantities (SignedArea, Area, Centroid) are exact for integer T
// only when the true value is integral; ContainsPoint and Perimeter go through
// division / square roots and truncate for integer T. Use a floating-point T for
// geometry.

// SignedArea returns the signed area of the polygon via the shoelace formula.
// It is positive for counter-clockwise winding and negative for clockwise.
func (p Polygon[T]) SignedArea() T {
	n := len(p.Points)
	if n < 3 {
		return 0
	}
	var sum T
	for i := range p.Points {
		j := (i + 1) % n
		sum += p.Points[i].X*p.Points[j].Y - p.Points[j].X*p.Points[i].Y
	}
	return sum / 2
}

// Centroid returns the area-weighted centroid (centre of mass) of the polygon.
// For a degenerate (zero-area) polygon it falls back to the average of the
// vertices.
func (p Polygon[T]) Centroid() Vector2[T] {
	n := len(p.Points)
	if n == 0 {
		return Vector2[T]{}
	}

	var cx, cy, a2 T // a2 accumulates twice the signed area
	for i := range p.Points {
		j := (i + 1) % n
		cross := p.Points[i].X*p.Points[j].Y - p.Points[j].X*p.Points[i].Y
		cx += (p.Points[i].X + p.Points[j].X) * cross
		cy += (p.Points[i].Y + p.Points[j].Y) * cross
		a2 += cross
	}

	if a2 == 0 {
		// Degenerate: average the vertices instead.
		var sum Vector2[T]
		for _, pt := range p.Points {
			sum = sum.Add(pt)
		}
		return sum.DivScalar(T(n))
	}
	// centroid = (1/(6A)) * (cx, cy) and 6A = 3*a2.
	return Vector2[T]{X: cx / (3 * a2), Y: cy / (3 * a2)}
}

// ContainsPoint reports whether point lies inside the polygon, using the
// even-odd crossing-number rule. Works for convex and concave polygons; the
// result for points exactly on an edge is unspecified.
func (p Polygon[T]) ContainsPoint(point Vector2[T]) bool {
	n := len(p.Points)
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		pi, pj := p.Points[i], p.Points[j]
		if (pi.Y > point.Y) != (pj.Y > point.Y) {
			xCross := (pj.X-pi.X)*(point.Y-pi.Y)/(pj.Y-pi.Y) + pi.X
			if point.X < xCross {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}

// BoundingBox returns the axis-aligned box that tightly encloses the polygon.
func (p Polygon[T]) BoundingBox() AABB[T] {
	return NewAABBFromPoints(p.Points...)
}

// Perimeter returns the total edge length of the polygon.
func (p Polygon[T]) Perimeter() T {
	n := len(p.Points)
	if n < 2 {
		return 0
	}
	var sum T
	for i := range p.Points {
		j := (i + 1) % n
		sum += p.Points[i].DistanceTo(p.Points[j])
	}
	return sum
}

// Translate returns a copy of the polygon with every vertex moved by delta. The
// original is not modified.
func (p Polygon[T]) Translate(delta Vector2[T]) Polygon[T] {
	pts := make([]Vector2[T], len(p.Points))
	for i, pt := range p.Points {
		pts[i] = pt.Add(delta)
	}
	return Polygon[T]{Points: pts}
}

// Clone returns a deep copy of the polygon with its own backing slice, so
// mutating one does not affect the other.
func (p Polygon[T]) Clone() Polygon[T] {
	pts := make([]Vector2[T], len(p.Points))
	copy(pts, p.Points)
	return Polygon[T]{Points: pts}
}

// EdgeSegment returns edge index as a LineSegment.
func (p Polygon[T]) EdgeSegment(index int) LineSegment[T] {
	a, b := p.Edge(index)
	return NewLineSegment(a, b)
}
