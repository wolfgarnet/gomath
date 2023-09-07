package gomath

import "math"

func Acos32(v float32) float32 {
	return float32(math.Acos(float64(v)))
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
