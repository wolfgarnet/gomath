package collider

import (
	"fmt"
	"github.com/wolfgarnet/gomath"
)

func IntersectingRay2[T gomath.Float](ray gomath.Ray[T], other any) (intersects bool, distance T) {
	switch t := other.(type) {
	case gomath.AABB[T]:
		return RayAABB2(ray, t)
	case *gomath.AABB[T]:
		return RayAABB2(ray, *t)
	case gomath.Circle[T]:
		return rayCircleDistance(ray, t)
	case *gomath.Circle[T]:
		return rayCircleDistance(ray, *t)

	default:
		panic(fmt.Sprintf("unknown type %T", other))
	}
}

// rayCircleDistance returns the distance to the nearest point where the ray meets
// the circle.
func rayCircleDistance[T gomath.Float](ray gomath.Ray[T], circle gomath.Circle[T]) (intersects bool, distance T) {
	intersections, p1, p2 := RayCircle(ray, circle)
	switch intersections {
	case 1:
		return true, ray.Origin.DistanceTo(p1)
	case 2:
		return true, gomath.Min(ray.Origin.DistanceTo(p1), ray.Origin.DistanceTo(p2))
	default:
		return false, 0
	}
}
