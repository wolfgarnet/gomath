package intersection

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
		intersections, p1, p2 := RayCircle(ray, t)
		switch intersections {
		case 1:
			return true, ray.Origin.DistanceTo(p1)
		case 2:
			d1 := ray.Origin.DistanceTo(p1)
			d2 := ray.Origin.DistanceTo(p2)
			if d1 > d2 {
				return true, d1
			}
			return true, d2
		default:
			return false, 0
		}
	case *gomath.Circle[T]:
		intersections, p1, p2 := RayCircle(ray, *t)
		switch intersections {
		case 1:
			return true, ray.Origin.DistanceTo(p1)
		case 2:
			d1 := ray.Origin.DistanceTo(p1)
			d2 := ray.Origin.DistanceTo(p2)
			if d1 > d2 {
				return true, d1
			}
			return true, d2
		default:
			return false, 0
		}

	default:
		panic(fmt.Sprintf("unknown type %T", other))
	}
}
