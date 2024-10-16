package gomath

var (
	VectorUp32    Vector2[float32] = Vector2[float32]{0, 1}
	VectorRight32 Vector2[float32] = Vector2[float32]{1, 0}
)

func assertAlmostEqual[T Number](a, b, tolerance T) bool {
	return Abs(a-b) < tolerance
}

func assertAlmostEqualVectors[T Number](a, b Vector2[T], tolerance T) bool {
	return (Abs(a.X-b.X) < tolerance) && (Abs(a.Y-b.Y) < tolerance)
}
