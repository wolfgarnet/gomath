package gomath

func TriangleFromConvexHullAndLength[T Float](center Vector2[T], polygon []Vector2[T], distance T) Polygon[T] {
	v1, v2 := ExtremePoints(center, polygon)
	v1 = v1.Sub(center).Normalize().MultiplyScalar(distance)
	v2 = v2.Sub(center).Normalize().MultiplyScalar(distance)

	return NewPolygon([]Vector2[T]{center, v1, v2})
}
