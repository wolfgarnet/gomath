package intersection

import "github.com/wolfgarnet/gomath"

func PointPolygon[T gomath.Float](point gomath.Vector2[T], polygon gomath.Polygon[T]) bool {
	var j int
	var count int
	ray := gomath.Ray[T]{
		Origin:    point,
		Direction: gomath.Vector2[T]{X: 1},
	}
	for i := 0; i < len(polygon.Points); i++ {
		if i == 0 {
			j = len(polygon.Points) - 1
		} else {
			j = i - 1
		}

		_, _, intersects := RayLineSegment(ray, polygon.Points[j], polygon.Points[i])
		if intersects {
			count++
		}
	}

	return count%2 != 0
}

func PointsInsideTriangle(p0, pl, pr gomath.Vector2[float32], points []gomath.Vector2[float32]) bool {
	for _, point := range points {
		if gomath.IsLeft(p0, pl, point) > 0 {
			return false
		}
		if gomath.IsLeft(p0, pr, point) < 0 {
			return false
		}
	}

	return true
}
