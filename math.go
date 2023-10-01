package gomath

import "math"

const epsilon = 0.9

func IsEpsilon[T Float](v T) bool {
	return Abs(v) < epsilon
}

func Sqrt32(v float32) float32 {
	return float32(math.Sqrt(float64(v)))
}

func Sqrt[T Number](v T) T {
	return T(math.Sqrt(float64(v)))
}

func Sqr[T Number](x T) T { return x * x }

func Acos32(v float32) float32 {
	return float32(math.Acos(float64(v)))
}

func Acos[T Number](v T) T {
	return T(math.Acos(float64(v)))
}

func Abs[T Number](v T) T {
	return T(math.Abs(float64(v)))
}

func Cos[T Number](theta T) T {
	return T(math.Cos(float64(theta)))
}

func Sin[T Number](theta T) T {
	return T(math.Sin(float64(theta)))
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
	max := math.Max(float64(a), float64(b))
	return T(max)
}

func Min[T Number](a, b T) T {
	min := math.Min(float64(a), float64(b))
	return T(min)
}
