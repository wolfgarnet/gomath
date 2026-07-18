package gomath

import "math"

// This file collects general-purpose 3D vector operations that round out
// Vector3 into a complete-ish library for games and geometry, mirroring the
// Vector2 additions. All operations are value-receiver and non-mutating.
//
// As with Vector2, methods are declared on Vector3[T Number], so operations
// that only make sense in floating point still compile for integer T and follow
// the same truncating behaviour as Length/Normalize. Use a floating-point T for
// anything involving magnitude, angles or trigonometry. Purely element-wise
// operations (Add, Sub, Mul, Cross, Min/Max, Manhattan/ChebyshevDistance, ...)
// are exact for integers.

// Cross returns the 3D vector cross product v x v2. The result is perpendicular
// to both operands, with a magnitude equal to the area of the parallelogram
// they span, and follows the right-hand rule.
func (v Vector3[T]) Cross(v2 Vector3[T]) Vector3[T] {
	return Vector3[T]{
		X: v.Y*v2.Z - v.Z*v2.Y,
		Y: v.Z*v2.X - v.X*v2.Z,
		Z: v.X*v2.Y - v.Y*v2.X,
	}
}

// Negate returns the vector pointing in the opposite direction.
func (v Vector3[T]) Negate() Vector3[T] {
	return Vector3[T]{X: -v.X, Y: -v.Y, Z: -v.Z}
}

// Abs returns the component-wise absolute value.
func (v Vector3[T]) Abs() Vector3[T] {
	return Vector3[T]{X: Abs(v.X), Y: Abs(v.Y), Z: Abs(v.Z)}
}

// Min returns the component-wise minimum of v and v2.
func (v Vector3[T]) Min(v2 Vector3[T]) Vector3[T] {
	return Vector3[T]{X: min(v.X, v2.X), Y: min(v.Y, v2.Y), Z: min(v.Z, v2.Z)}
}

// Max returns the component-wise maximum of v and v2.
func (v Vector3[T]) Max(v2 Vector3[T]) Vector3[T] {
	return Vector3[T]{X: max(v.X, v2.X), Y: max(v.Y, v2.Y), Z: max(v.Z, v2.Z)}
}

// Clamp returns the vector with each component clamped into the box spanned by
// lo and hi (component-wise).
func (v Vector3[T]) Clamp(lo, hi Vector3[T]) Vector3[T] {
	return Vector3[T]{X: Clamp(v.X, lo.X, hi.X), Y: Clamp(v.Y, lo.Y, hi.Y), Z: Clamp(v.Z, lo.Z, hi.Z)}
}

// ClampLength returns the vector scaled so its length lies within [minLen, maxLen].
// Direction is preserved; the zero vector is returned unchanged.
func (v Vector3[T]) ClampLength(minLen, maxLen T) Vector3[T] {
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
func (v Vector3[T]) MoveTowards(target Vector3[T], maxDelta T) Vector3[T] {
	d := target.Sub(v)
	dist := d.Length()
	if dist == 0 || dist <= maxDelta {
		return target
	}
	return v.Add(d.MultiplyScalar(maxDelta / dist))
}

// Project returns the vector projection of v onto onto. If onto is the zero
// vector, the zero vector is returned.
func (v Vector3[T]) Project(onto Vector3[T]) Vector3[T] {
	d := onto.Dot(onto)
	if d == 0 {
		return Vector3[T]{}
	}
	return onto.MultiplyScalar(v.Dot(onto) / d)
}

// Reject returns the component of v orthogonal to onto, i.e. v - v.Project(onto).
func (v Vector3[T]) Reject(onto Vector3[T]) Vector3[T] {
	return v.Sub(v.Project(onto))
}

// ManhattanDistance returns the L1 (taxicab) distance to v2.
func (v Vector3[T]) ManhattanDistance(v2 Vector3[T]) T {
	return Abs(v.X-v2.X) + Abs(v.Y-v2.Y) + Abs(v.Z-v2.Z)
}

// ChebyshevDistance returns the L-infinity (chessboard) distance to v2.
func (v Vector3[T]) ChebyshevDistance(v2 Vector3[T]) T {
	return max(Abs(v.X-v2.X), Abs(v.Y-v2.Y), Abs(v.Z-v2.Z))
}

// Floor returns the component-wise floor.
func (v Vector3[T]) Floor() Vector3[T] {
	return Vector3[T]{X: Floor(v.X), Y: Floor(v.Y), Z: Floor(v.Z)}
}

// Ceil returns the component-wise ceiling.
func (v Vector3[T]) Ceil() Vector3[T] {
	return Vector3[T]{
		X: T(math.Ceil(float64(v.X))),
		Y: T(math.Ceil(float64(v.Y))),
		Z: T(math.Ceil(float64(v.Z))),
	}
}

// Round returns the component-wise rounding to the nearest integer, halves away
// from zero.
func (v Vector3[T]) Round() Vector3[T] {
	return Vector3[T]{X: Round(v.X), Y: Round(v.Y), Z: Round(v.Z)}
}

// IsZero reports whether all components are zero.
func (v Vector3[T]) IsZero() bool {
	return v.X == 0 && v.Y == 0 && v.Z == 0
}

// IsNaN reports whether any component is NaN.
func (v Vector3[T]) IsNaN() bool {
	return math.IsNaN(float64(v.X)) || math.IsNaN(float64(v.Y)) || math.IsNaN(float64(v.Z))
}

// IsFinite reports whether all components are finite (neither NaN nor Inf).
func (v Vector3[T]) IsFinite() bool {
	for _, c := range [3]float64{float64(v.X), float64(v.Y), float64(v.Z)} {
		if math.IsNaN(c) || math.IsInf(c, 0) {
			return false
		}
	}
	return true
}

// ScaleToLength returns the vector scaled to the given length, preserving
// direction. Unlike SetLength, a length of 0 (or a zero vector) returns the
// vector unchanged rather than collapsing it to the zero vector.
func (v Vector3[T]) ScaleToLength(length T) Vector3[T] {
	if length == 0 || v.IsZero() {
		return v
	}
	return v.MultiplyScalar(length / v.Length())
}

// DirectionTo returns the unit vector pointing from v towards other.
func (v Vector3[T]) DirectionTo(other Vector3[T]) Vector3[T] {
	return other.Sub(v).Normalize()
}

// RotateAxis rotates the vector counter-clockwise by phi radians about the given
// axis (looking down the axis towards the origin), using Rodrigues' rotation
// formula. The axis is assumed to be of unit length.
func (v Vector3[T]) RotateAxis(axis Vector3[T], phi T) Vector3[T] {
	c := Cos(phi)
	s := Sin(phi)
	// v*cos + (axis x v)*sin + axis*(axis . v)*(1 - cos)
	cross := axis.Cross(v)
	dot := axis.Dot(v)
	return v.MultiplyScalar(c).
		Add(cross.MultiplyScalar(s)).
		Add(axis.MultiplyScalar(dot * (1 - c)))
}

// Zero3 returns the zero vector {0, 0, 0}.
func Zero3[T Number]() Vector3[T] { return Vector3[T]{0, 0, 0} }

// One3 returns the vector {1, 1, 1}.
func One3[T Number]() Vector3[T] { return Vector3[T]{1, 1, 1} }

// Up3 returns {0, 1, 0}. The package convention is Y-up, matching Vector2's Up.
func Up3[T Number]() Vector3[T] { return Vector3[T]{0, 1, 0} }

// Down3 returns {0, -1, 0}.
func Down3[T Number]() Vector3[T] { return Vector3[T]{0, -1, 0} }

// Left3 returns {-1, 0, 0}.
func Left3[T Number]() Vector3[T] { return Vector3[T]{-1, 0, 0} }

// Right3 returns {1, 0, 0}.
func Right3[T Number]() Vector3[T] { return Vector3[T]{1, 0, 0} }

// Forward3 returns {0, 0, 1}. The package convention is +Z forward.
func Forward3[T Number]() Vector3[T] { return Vector3[T]{0, 0, 1} }

// Back3 returns {0, 0, -1}.
func Back3[T Number]() Vector3[T] { return Vector3[T]{0, 0, -1} }

// UnitX3 returns {1, 0, 0}.
func UnitX3[T Number]() Vector3[T] { return Vector3[T]{1, 0, 0} }

// UnitY3 returns {0, 1, 0}.
func UnitY3[T Number]() Vector3[T] { return Vector3[T]{0, 1, 0} }

// UnitZ3 returns {0, 0, 1}.
func UnitZ3[T Number]() Vector3[T] { return Vector3[T]{0, 0, 1} }
