package gomath

// This file rounds the Ray type out with self-contained helpers (point
// sampling, closest-point, normalization). Cross-shape intersection (ray-AABB,
// ray-circle, ray-segment) lives in the gomath/intersection package and is
// intentionally not duplicated here.
//
// ClosestPoint/DistanceToPoint assume nothing about Direction's length (they
// divide by Direction.Dot(Direction)); PointAt measures distance in units of
// Direction, so use a unit Direction (see Normalized) for true distances. These
// helpers round-trip through float64 for magnitude and truncate for integer T.

// PointAt returns the point at the given distance along the ray, measured in
// units of Direction: Origin + Direction*distance.
func (r Ray[T]) PointAt(distance T) Vector2[T] {
	return r.Origin.Add(r.Direction.MultiplyScalar(distance))
}

// Normalized returns a copy of the ray with a unit-length Direction.
func (r Ray[T]) Normalized() Ray[T] {
	return Ray[T]{Origin: r.Origin, Direction: r.Direction.Normalize()}
}

// ClosestPoint returns the point on the ray (at or ahead of the origin) closest
// to point.
func (r Ray[T]) ClosestPoint(point Vector2[T]) Vector2[T] {
	den := r.Direction.Dot(r.Direction)
	if den == 0 {
		// Degenerate ray (zero direction).
		return r.Origin
	}
	t := point.Sub(r.Origin).Dot(r.Direction) / den
	if t <= 0 {
		return r.Origin
	}
	return r.Origin.Add(r.Direction.MultiplyScalar(t))
}

// DistanceToPointSquared returns the squared distance from point to the ray.
func (r Ray[T]) DistanceToPointSquared(point Vector2[T]) T {
	return r.ClosestPoint(point).DistanceToSquared(point)
}

// DistanceToPoint returns the distance from point to the ray.
func (r Ray[T]) DistanceToPoint(point Vector2[T]) T {
	return Sqrt(r.DistanceToPointSquared(point))
}

// Equals reports whether r and other have the same origin and direction.
func (r Ray[T]) Equals(other Ray[T]) bool {
	return r == other
}
