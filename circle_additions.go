package gomath

import "math"

// This file rounds the Circle type out into a production-grade primitive for
// games: containment, overlap (circle-circle and circle-AABB), closest-point
// and bounds. All additions are value-receiver and non-mutating, and Circle is
// treated as a filled disk.
//
// The containment and overlap tests use squared distances, so they are exact
// for integer T. Operations that need a true magnitude or pi (ClosestPoint,
// Area, Circumference) round-trip through float64 and truncate for integer T -
// use a floating-point T for those.

// Contains reports whether point lies inside the disk, boundary inclusive.
func (c Circle[T]) Contains(point Vector2[T]) bool {
	return point.DistanceToSquared(c.Center) <= c.Radius*c.Radius
}

// ContainsCircle reports whether other lies entirely inside c, boundary
// inclusive.
func (c Circle[T]) ContainsCircle(other Circle[T]) bool {
	if other.Radius > c.Radius {
		return false
	}
	gap := c.Radius - other.Radius
	return c.Center.DistanceToSquared(other.Center) <= gap*gap
}

// Intersects reports whether c and other overlap. Circles that touch at a single
// point count as intersecting.
func (c Circle[T]) Intersects(other Circle[T]) bool {
	sum := c.Radius + other.Radius
	return c.Center.DistanceToSquared(other.Center) <= sum*sum
}

// IntersectsAABB reports whether the disk overlaps the box. Touching counts as
// intersecting.
func (c Circle[T]) IntersectsAABB(box AABB[T]) bool {
	closest := box.ClosestPoint(c.Center)
	return closest.DistanceToSquared(c.Center) <= c.Radius*c.Radius
}

// ClosestPoint returns the point of the filled disk closest to point: point
// itself when it is inside, otherwise its projection onto the boundary.
func (c Circle[T]) ClosestPoint(point Vector2[T]) Vector2[T] {
	if c.Contains(point) {
		return point
	}
	return c.Center.Add(point.Sub(c.Center).Normalize().MultiplyScalar(c.Radius))
}

// Diameter returns 2 * Radius.
func (c Circle[T]) Diameter() T {
	return c.Radius * 2
}

// Area returns the area of the disk, pi * r^2.
func (c Circle[T]) Area() T {
	r := float64(c.Radius)
	return T(math.Pi * r * r)
}

// Circumference returns the perimeter of the circle, 2 * pi * r.
func (c Circle[T]) Circumference() T {
	return T(2 * math.Pi * float64(c.Radius))
}

// BoundingBox returns the axis-aligned box that tightly encloses the circle.
func (c Circle[T]) BoundingBox() AABB[T] {
	return NewAABB(c.Center, c.Radius, c.Radius)
}

// Translate returns a copy of the circle moved by delta.
func (c Circle[T]) Translate(delta Vector2[T]) Circle[T] {
	return Circle[T]{Center: c.Center.Add(delta), Radius: c.Radius}
}

// Equals reports whether c and other have the same centre and radius.
func (c Circle[T]) Equals(other Circle[T]) bool {
	return c == other
}
