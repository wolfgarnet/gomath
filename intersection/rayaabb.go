package intersection

import "github.com/wolfgarnet/gomath"

func RayAABB[T gomath.Float](ray gomath.Ray[T], aabb gomath.AABB[T]) bool {
	for _, s := range aabb.Sides() {
		if _, _, intersects := RayLineSegment(ray, s.P1, s.P2); intersects {
			return true
		}
	}

	return false
}

func RayAABB2[T gomath.Float](ray gomath.Ray[T], aabb gomath.AABB[T]) bool {
	for _, s := range aabb.Sides() {
		if _, _, intersects := RayLineSegment(ray, s.P1, s.P2); intersects {
			return true
		}
	}

	return false
}
