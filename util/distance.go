package util

import (
	"github.com/wolfgarnet/gomath"
	"github.com/wolfgarnet/gomath/intersection"
	"math"
)

func ShortestDistance[T gomath.Float](a gomath.AABB[T], b gomath.AABB[T]) T {
	if intersection.AABBAABB(a, b) {
		return 0
	}
	var closest = T(math.MaxFloat32)
	for _, a1 := range a.Points() {
		for _, a2 := range b.Points() {
			d := a1.DistanceToSquared(a2)
			if d < closest {
				closest = d
			}
		}
	}

	return gomath.Sqrt(closest)
}
