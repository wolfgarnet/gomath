package gomath

import "math"

// This file rounds Vector4 out with the completeness helpers the other vector
// types already carry, plus the Vector3 bridge.

// ToVector4 lifts a 3D vector into homogeneous coordinates with the given W.
// Use w=1 for a position, so that it picks up the translation column of a
// transform, and w=0 for a direction, so that it does not. It mirrors
// Vector2.ToVector3.
func (v Vector3[T]) ToVector4(w T) Vector4[T] {
	return Vector4[T]{
		X: v.X,
		Y: v.Y,
		Z: v.Z,
		W: w,
	}
}

// Negate returns the vector pointing the opposite way.
func (v Vector4[T]) Negate() Vector4[T] {
	return Vector4[T]{X: -v.X, Y: -v.Y, Z: -v.Z, W: -v.W}
}

// Abs returns the component-wise absolute value.
func (v Vector4[T]) Abs() Vector4[T] {
	return Vector4[T]{X: Abs(v.X), Y: Abs(v.Y), Z: Abs(v.Z), W: Abs(v.W)}
}

// Min returns the component-wise minimum of v and v2.
func (v Vector4[T]) Min(v2 Vector4[T]) Vector4[T] {
	return Vector4[T]{X: min(v.X, v2.X), Y: min(v.Y, v2.Y), Z: min(v.Z, v2.Z), W: min(v.W, v2.W)}
}

// Max returns the component-wise maximum of v and v2.
func (v Vector4[T]) Max(v2 Vector4[T]) Vector4[T] {
	return Vector4[T]{X: max(v.X, v2.X), Y: max(v.Y, v2.Y), Z: max(v.Z, v2.Z), W: max(v.W, v2.W)}
}

// Clamp constrains each component to the range given by lo and hi.
func (v Vector4[T]) Clamp(lo, hi Vector4[T]) Vector4[T] {
	return Vector4[T]{
		X: Clamp(v.X, lo.X, hi.X),
		Y: Clamp(v.Y, lo.Y, hi.Y),
		Z: Clamp(v.Z, lo.Z, hi.Z),
		W: Clamp(v.W, lo.W, hi.W),
	}
}

// IsZero reports whether every component is exactly zero.
func (v Vector4[T]) IsZero() bool {
	return v.X == 0 && v.Y == 0 && v.Z == 0 && v.W == 0
}

// ApproxEquals reports whether every component of v is within tol of the
// corresponding component of v2.
func (v Vector4[T]) ApproxEquals(v2 Vector4[T], tol T) bool {
	return Abs(v.X-v2.X) <= tol &&
		Abs(v.Y-v2.Y) <= tol &&
		Abs(v.Z-v2.Z) <= tol &&
		Abs(v.W-v2.W) <= tol
}

// IsNaN reports whether any component is NaN.
func (v Vector4[T]) IsNaN() bool {
	return math.IsNaN(float64(v.X)) ||
		math.IsNaN(float64(v.Y)) ||
		math.IsNaN(float64(v.Z)) ||
		math.IsNaN(float64(v.W))
}

// IsFinite reports whether every component is neither NaN nor an infinity.
func (v Vector4[T]) IsFinite() bool {
	return !v.IsNaN() &&
		!math.IsInf(float64(v.X), 0) &&
		!math.IsInf(float64(v.Y), 0) &&
		!math.IsInf(float64(v.Z), 0) &&
		!math.IsInf(float64(v.W), 0)
}

// Zero4 returns the zero vector {0, 0, 0, 0}.
func Zero4[T Number]() Vector4[T] { return Vector4[T]{} }

// One4 returns the vector {1, 1, 1, 1}.
func One4[T Number]() Vector4[T] { return Vector4[T]{1, 1, 1, 1} }

// UnitX4 returns {1, 0, 0, 0}.
func UnitX4[T Number]() Vector4[T] { return Vector4[T]{1, 0, 0, 0} }

// UnitY4 returns {0, 1, 0, 0}.
func UnitY4[T Number]() Vector4[T] { return Vector4[T]{0, 1, 0, 0} }

// UnitZ4 returns {0, 0, 1, 0}.
func UnitZ4[T Number]() Vector4[T] { return Vector4[T]{0, 0, 1, 0} }

// UnitW4 returns {0, 0, 0, 1}.
func UnitW4[T Number]() Vector4[T] { return Vector4[T]{0, 0, 0, 1} }
