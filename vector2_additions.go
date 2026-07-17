package gomath

import "math"

// This file collects general-purpose 2D vector operations that round out
// Vector2 into a complete-ish library for games and geometry. All operations
// are value-receiver and non-mutating, matching the rest of Vector2.
//
// Methods are declared on Vector2[T Number] (Go does not allow a method to add
// type constraints beyond its receiver), so operations that only make sense in
// floating point still compile for integer T and simply follow the same
// truncating behaviour as the existing Length/Angle/Normalize methods. Use a
// floating-point T for anything involving magnitude, angles or trigonometry.

// Cross returns the 2D scalar cross product (the z-component of the 3D cross
// product), v.X*v2.Y - v.Y*v2.X. It is positive when v2 is counter-clockwise
// from v, negative when clockwise, and zero when the two are parallel.
func (v Vector2[T]) Cross(v2 Vector2[T]) T {
	return v.X*v2.Y - v.Y*v2.X
}

// Perp returns the counter-clockwise (left) perpendicular of the vector,
// {-Y, X}. It has the same length as v and is rotated a quarter turn CCW.
func (v Vector2[T]) Perp() Vector2[T] {
	return Vector2[T]{X: -v.Y, Y: v.X}
}

// Negate returns the vector pointing in the opposite direction, {-X, -Y}.
func (v Vector2[T]) Negate() Vector2[T] {
	return Vector2[T]{X: -v.X, Y: -v.Y}
}

// Abs returns the component-wise absolute value.
func (v Vector2[T]) Abs() Vector2[T] {
	return Vector2[T]{X: Abs(v.X), Y: Abs(v.Y)}
}

// Min returns the component-wise minimum of v and v2.
func (v Vector2[T]) Min(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{X: min(v.X, v2.X), Y: min(v.Y, v2.Y)}
}

// Max returns the component-wise maximum of v and v2.
func (v Vector2[T]) Max(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{X: max(v.X, v2.X), Y: max(v.Y, v2.Y)}
}

// Clamp returns the vector with each component clamped into the box spanned by
// lo and hi (component-wise).
func (v Vector2[T]) Clamp(lo, hi Vector2[T]) Vector2[T] {
	return Vector2[T]{X: Clamp(v.X, lo.X, hi.X), Y: Clamp(v.Y, lo.Y, hi.Y)}
}

// ClampLength returns the vector scaled so its length lies within [minLen, maxLen].
// Direction is preserved; the zero vector is returned unchanged.
func (v Vector2[T]) ClampLength(minLen, maxLen T) Vector2[T] {
	l := v.Length()
	if l == 0 {
		return v
	}
	if l < minLen {
		return v.MultiplyScalar(minLen / l)
	}
	if l > maxLen {
		return v.MultiplyScalar(maxLen / l)
	}
	return v
}

// MoveTowards returns v moved towards target by at most maxDelta. If target is
// within maxDelta, target is returned exactly (no overshoot).
func (v Vector2[T]) MoveTowards(target Vector2[T], maxDelta T) Vector2[T] {
	d := target.Sub(v)
	dist := d.Length()
	if dist == 0 || dist <= maxDelta {
		return target
	}
	return v.Add(d.MultiplyScalar(maxDelta / dist))
}

// Project returns the vector projection of v onto onto. If onto is the zero
// vector, the zero vector is returned.
func (v Vector2[T]) Project(onto Vector2[T]) Vector2[T] {
	d := onto.Dot(onto)
	if d == 0 {
		return Vector2[T]{}
	}
	return onto.MultiplyScalar(v.Dot(onto) / d)
}

// Reject returns the component of v orthogonal to onto, i.e. v - v.Project(onto).
func (v Vector2[T]) Reject(onto Vector2[T]) Vector2[T] {
	return v.Sub(v.Project(onto))
}

// Angle returns the orientation of the vector in radians, measured
// counter-clockwise from the positive X axis, in the range (-pi, pi].
func (v Vector2[T]) Angle() T {
	return Atan2(v.Y, v.X)
}

// RotateAround rotates the vector counter-clockwise by phi radians about pivot.
func (v Vector2[T]) RotateAround(pivot Vector2[T], phi T) Vector2[T] {
	return v.Sub(pivot).Rotate(phi).Add(pivot)
}

// ManhattanDistance returns the L1 (taxicab) distance to v2: |dx| + |dy|.
func (v Vector2[T]) ManhattanDistance(v2 Vector2[T]) T {
	return Abs(v.X-v2.X) + Abs(v.Y-v2.Y)
}

// ChebyshevDistance returns the L-infinity (chessboard) distance to v2:
// max(|dx|, |dy|).
func (v Vector2[T]) ChebyshevDistance(v2 Vector2[T]) T {
	return max(Abs(v.X-v2.X), Abs(v.Y-v2.Y))
}

// Floor returns the component-wise floor.
func (v Vector2[T]) Floor() Vector2[T] {
	return Vector2[T]{X: Floor(v.X), Y: Floor(v.Y)}
}

// Ceil returns the component-wise ceiling.
func (v Vector2[T]) Ceil() Vector2[T] {
	return Vector2[T]{X: T(math.Ceil(float64(v.X))), Y: T(math.Ceil(float64(v.Y)))}
}

// Round returns the component-wise rounding to the nearest integer, halves away
// from zero.
func (v Vector2[T]) Round() Vector2[T] {
	return Vector2[T]{X: Round(v.X), Y: Round(v.Y)}
}

// IsNaN reports whether either component is NaN.
func (v Vector2[T]) IsNaN() bool {
	return math.IsNaN(float64(v.X)) || math.IsNaN(float64(v.Y))
}

// IsFinite reports whether both components are finite (neither NaN nor Inf).
func (v Vector2[T]) IsFinite() bool {
	x, y := float64(v.X), float64(v.Y)
	return !math.IsNaN(x) && !math.IsInf(x, 0) && !math.IsNaN(y) && !math.IsInf(y, 0)
}

// FromAngle returns the unit vector pointing at theta radians counter-clockwise
// from the positive X axis: {cos(theta), sin(theta)}.
func FromAngle[T Float](theta T) Vector2[T] {
	return Vector2[T]{X: Cos(theta), Y: Sin(theta)}
}

// FromPolar returns the vector with the given length pointing at theta radians
// counter-clockwise from the positive X axis.
func FromPolar[T Float](theta, length T) Vector2[T] {
	return Vector2[T]{X: length * Cos(theta), Y: length * Sin(theta)}
}

// Zero returns the zero vector {0, 0}.
func Zero[T Number]() Vector2[T] { return Vector2[T]{0, 0} }

// One returns the vector {1, 1}.
func One[T Number]() Vector2[T] { return Vector2[T]{1, 1} }

// Up returns the unit vector {0, 1} (Y points up, matching VectorUp32).
func Up[T Number]() Vector2[T] { return Vector2[T]{0, 1} }

// Down returns the unit vector {0, -1}.
func Down[T Number]() Vector2[T] { return Vector2[T]{0, -1} }

// Left returns the unit vector {-1, 0}.
func Left[T Number]() Vector2[T] { return Vector2[T]{-1, 0} }

// Right returns the unit vector {1, 0} (matching VectorRight32).
func Right[T Number]() Vector2[T] { return Vector2[T]{1, 0} }

// UnitX returns the unit vector {1, 0}.
func UnitX[T Number]() Vector2[T] { return Vector2[T]{1, 0} }

// UnitY returns the unit vector {0, 1}.
func UnitY[T Number]() Vector2[T] { return Vector2[T]{0, 1} }
