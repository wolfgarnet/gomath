package intersection

import (
	"github.com/wolfgarnet/gomath"
	"math"
)

func RayAABB[T gomath.Float](ray gomath.Ray[T], aabb gomath.AABB[T]) bool {
	for _, s := range aabb.Sides() {
		if _, _, intersects := RayLineSegment(ray, s.P1, s.P2); intersects {
			return true
		}
	}

	return false
}

func RayAABB2[T gomath.Float](ray gomath.Ray[T], aabb gomath.AABB[T]) (intersects bool, ratio T) {
	ratio = math.MaxFloat32
	for _, s := range aabb.Sides() {
		r, _, i := RayLineSegment(ray, s.P1, s.P2)
		if i {
			if r < ratio {
				ratio = r
			}
			intersects = true
		}
	}

	return intersects, ratio
}
