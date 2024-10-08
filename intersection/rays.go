package intersection

import (
	"fmt"
	"github.com/wolfgarnet/gomath"
)

func IntersectingRay2[T gomath.Float](ray gomath.Ray[T], other any) (intersects bool, distance T) {
	switch t := other.(type) {
	case gomath.AABB[T]:
		return RayAABB2(ray, t)

	default:
		panic(fmt.Sprintf("unknown type %T", other))
	}
}
