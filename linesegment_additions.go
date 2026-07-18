package gomath

// This file rounds the LineSegment type out with self-contained geometric
// helpers (point sampling, closest-point, bounds). Cross-shape intersection
// (segment-segment, segment-AABB, ...) lives in the gomath/intersection
// package and is intentionally not duplicated here.
//
// Helpers that need a true magnitude (ClosestPoint, DistanceToPoint,
// UnitDirection) round-trip through float64 and truncate for integer T; the
// element-wise helpers (Midpoint, LengthSquared, Reversed, Translate,
// BoundingBox, Equals) are exact for integers.

// PointAt returns the point a fraction t along the segment: t=0 is P1, t=1 is
// P2. Values outside [0,1] extrapolate along the line.
func (l LineSegment[T]) PointAt(t T) Vector2[T] {
	return l.P1.Lerp(l.P2, t)
}

// Midpoint returns the point halfway between P1 and P2.
func (l LineSegment[T]) Midpoint() Vector2[T] {
	return Vector2[T]{X: (l.P1.X + l.P2.X) / 2, Y: (l.P1.Y + l.P2.Y) / 2}
}

// LengthSquared returns the squared length of the segment (cheaper than Length,
// and exact for integer T).
func (l LineSegment[T]) LengthSquared() T {
	return l.P1.DistanceToSquared(l.P2)
}

// UnitDirection returns the segment's direction as a unit vector (P1 -> P2).
func (l LineSegment[T]) UnitDirection() Vector2[T] {
	return l.P2.Sub(l.P1).Normalize()
}

// ClosestPoint returns the point on the segment closest to point.
func (l LineSegment[T]) ClosestPoint(point Vector2[T]) Vector2[T] {
	d := l.P2.Sub(l.P1)
	den := d.X*d.X + d.Y*d.Y
	if den == 0 {
		// Degenerate segment (P1 == P2).
		return l.P1
	}
	p := point.Sub(l.P1)
	t := (p.X*d.X + p.Y*d.Y) / den
	if t <= 0 {
		return l.P1
	}
	if t >= 1 {
		return l.P2
	}
	return l.P1.Add(d.MultiplyScalar(t))
}

// DistanceToPointSquared returns the squared distance from point to the segment.
func (l LineSegment[T]) DistanceToPointSquared(point Vector2[T]) T {
	return l.ClosestPoint(point).DistanceToSquared(point)
}

// DistanceToPoint returns the distance from point to the segment.
func (l LineSegment[T]) DistanceToPoint(point Vector2[T]) T {
	return Sqrt(l.DistanceToPointSquared(point))
}

// Reversed returns the segment with its endpoints swapped.
func (l LineSegment[T]) Reversed() LineSegment[T] {
	return LineSegment[T]{P1: l.P2, P2: l.P1}
}

// Translate returns a copy of the segment moved by delta.
func (l LineSegment[T]) Translate(delta Vector2[T]) LineSegment[T] {
	return LineSegment[T]{P1: l.P1.Add(delta), P2: l.P2.Add(delta)}
}

// BoundingBox returns the axis-aligned box that tightly encloses the segment.
func (l LineSegment[T]) BoundingBox() AABB[T] {
	return NewAABBFromPoints(l.P1, l.P2)
}

// Equals reports whether l and other have the same endpoints in the same order.
func (l LineSegment[T]) Equals(other LineSegment[T]) bool {
	return l == other
}
