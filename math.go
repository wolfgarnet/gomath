package gomath

import "math"

const epsilon2 = 0.001

// IsNearlyEqual reports whether a and b are within 0.001 of each other.
//
// The threshold is absolute, so this is only meaningful for values of roughly
// unit magnitude. For anything else compare against a tolerance scaled to your
// own data.
func IsNearlyEqual[T Float](a, b T) bool {
	return Abs(a-b) < epsilon2
}

// IsEpsilon was removed. It reported Abs(v) < 0.9, which reads as a near-zero
// test but is nothing of the kind, and its single caller — collider.LineAABB —
// used it as a divide-by-zero guard on non-normalised direction components.
// Every component under 0.9 was therefore treated as parallel to its axis, and
// the axis was never clipped against, so the test reported intersections that
// did not exist. See collider/lineaabb_regression_test.go.
//
// For a genuine near-zero test compare against the exported Epsilon (1e-5), or
// better, against a tolerance derived from the magnitudes involved.

func Sqrt32(v float32) float32 {
	return float32(math.Sqrt(float64(v)))
}

func Sqrt[T Number](v T) T {
	return T(math.Sqrt(float64(v)))
}

func Sqr[T Number](x T) T { return x * x }

func Pow[T Number](v T, y T) T {
	return T(math.Pow(float64(v), float64(y)))
}

func Acos32(v float32) float32 {
	return float32(math.Acos(float64(v)))
}

func Acos[T Number](v T) T {
	return T(math.Acos(float64(v)))
}

// Asin returns the arcsine of v in radians. v must be in [-1, 1]; outside that
// the result is NaN.
func Asin[T Number](v T) T {
	return T(math.Asin(float64(v)))
}

// Atan returns the arctangent of v in radians, in (-pi/2, pi/2). When both
// coordinates of a direction are known, prefer Atan2, which resolves the
// quadrant.
func Atan[T Number](v T) T {
	return T(math.Atan(float64(v)))
}

// DegToRad converts degrees to radians. Every angle in this package is in
// radians; this exists for the boundary where degrees come in from data files
// or a UI.
func DegToRad[T Number](degrees T) T {
	return T(float64(degrees) * math.Pi / 180)
}

// RadToDeg converts radians to degrees.
func RadToDeg[T Number](radians T) T {
	return T(float64(radians) * 180 / math.Pi)
}

// Lerp linearly interpolates from a to b by t, where t == 0 returns a and
// t == 1 returns b. t is not clamped, so values outside [0, 1] extrapolate.
func Lerp[T Number](a T, b T, t T) T {
	return a + (b-a)*t
}

// Sign returns -1 for a negative value, 1 for a positive one, and 0 for zero.
func Sign[T Number](v T) T {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}

	return 0
}

// Atan2 returns the angle in radians of the point (x, y) from the positive X
// axis, in (-pi, pi].
//
// The argument order is y THEN x, matching math.Atan2 and every other atan2 in
// existence. The parameters used to be named (x, y) while being passed straight
// through in that order, so the names said the opposite of what the function
// did; the behaviour was always the stdlib's, and every caller already relied on
// it, so this is a rename and not a change.
func Atan2[T Number](y, x T) T {
	return T(math.Atan2(float64(y), float64(x)))
}

func Abs[T Number](v T) T {
	if v < 0 {
		return -v
	}
	return v
}

func Cos[T Number](theta T) T {
	return T(math.Cos(float64(theta)))
}

func Sin[T Number](theta T) T {
	return T(math.Sin(float64(theta)))
}

// Tan returns the tangent of theta radians. It is undefined at odd multiples of
// pi/2, where the true value is infinite.
func Tan[T Number](theta T) T {
	return T(math.Tan(float64(theta)))
}

func Floor[T Number](v T) T {
	return T(math.Floor(float64(v)))
}

func Mod[T Number](v T, d T) T {
	return T(math.Mod(float64(v), float64(d)))
}

func Clamp[T Number](x, a, b T) T {
	if x < a {
		return a
	}
	if x > b {
		return b
	}
	return x
}

func Max[T Number](a, b T) T {
	return max(a, b)
}

func Min[T Number](a, b T) T {
	return min(a, b)
}

func Round[T Number](n T) T {
	return T(math.Round(float64(n)))
}

func Round2[T Number](n T, precision int) T {
	output := T(math.Pow(10, float64(precision)))
	return (Round(n * output)) / output
}
