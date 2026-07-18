package gomath

import "math"

// This file rounds the Ellipsoid (a 2D axis-aligned ellipse: Position + Radii)
// out with self-contained helpers. Cross-shape queries such as point-to-ellipse
// distance and ellipse-ellipse intersection live in the gomath/intersection
// package and are intentionally not duplicated here.
//
// Contains uses a cross-multiplied test, so it is exact for integer T; Area and
// PointAt go through float64 and truncate for integer T.

// Contains reports whether point lies inside the ellipse, boundary inclusive.
func (e Ellipsoid[T]) Contains(point Vector2[T]) bool {
	dx := point.X - e.Position.X
	dy := point.Y - e.Position.Y
	rx, ry := e.Radii.X, e.Radii.Y
	// (dx/rx)^2 + (dy/ry)^2 <= 1, cross-multiplied to avoid division.
	return dx*dx*ry*ry+dy*dy*rx*rx <= rx*rx*ry*ry
}

// Area returns the area of the ellipse, pi * rx * ry.
func (e Ellipsoid[T]) Area() T {
	return T(math.Pi * float64(e.Radii.X) * float64(e.Radii.Y))
}

// BoundingBox returns the axis-aligned box that tightly encloses the ellipse.
func (e Ellipsoid[T]) BoundingBox() AABB[T] {
	return NewAABB(e.Position, e.Radii.X, e.Radii.Y)
}

// PointAt returns the point on the ellipse boundary at the given parametric
// angle (radians): Position + {rx*cos, ry*sin}. Note the parameter is not the
// geometric angle from the centre except on the axes.
func (e Ellipsoid[T]) PointAt(angle T) Vector2[T] {
	return Vector2[T]{
		X: e.Position.X + T(float64(e.Radii.X)*math.Cos(float64(angle))),
		Y: e.Position.Y + T(float64(e.Radii.Y)*math.Sin(float64(angle))),
	}
}

// Translate returns a copy of the ellipse moved by delta.
func (e Ellipsoid[T]) Translate(delta Vector2[T]) Ellipsoid[T] {
	return Ellipsoid[T]{Position: e.Position.Add(delta), Radii: e.Radii}
}

// IsCircle reports whether the two radii are equal.
func (e Ellipsoid[T]) IsCircle() bool {
	return e.Radii.X == e.Radii.Y
}

// Equals reports whether e and other have the same position and radii.
func (e Ellipsoid[T]) Equals(other Ellipsoid[T]) bool {
	return e == other
}
