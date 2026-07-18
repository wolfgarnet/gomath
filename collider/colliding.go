package collider

import (
	"fmt"

	"github.com/wolfgarnet/gomath"
)

type ColliderType uint8

const (
	ColliderAABB ColliderType = iota
	ColliderCircle
	ColliderLineSegment
)

type Collider[T gomath.Float] struct {
	Type   ColliderType
	Offset gomath.Vector2[T]
	Shape  any
}

func ColliderPenetration[T gomath.Float](first Collider[T], second Collider[T]) (penetration T, surfaceNormal gomath.Vector2[T], overlapping bool) {
	switch first.Type {
	case ColliderAABB:
		return ColliderAABBPenetration(first.Shape.(gomath.AABBCollider[T]), first.Offset, second)
	case ColliderCircle:
		return ColliderCirclePenetration(first.Shape.(gomath.CircleCollider[T]), first.Offset, second)
	//case ColliderLineSegment:
	default:
		fmt.Printf("Unsupported collider type for penetration: %T\n", first.Type)
		return penetration, surfaceNormal, false
	}
}

func ColliderCirclePenetration[T gomath.Float](circle gomath.CircleCollider[T], offset gomath.Vector2[T], other Collider[T]) (penetration T, surfaceNormal gomath.Vector2[T], overlapping bool) {
	switch other.Type {
	case ColliderCircle:
		//pen, nx, ny, ok := collider.CircleCirclePenetration(circle.Radius, offset.X, offset.Y, other.Shape.(gomath.CircleCollider[T]).Radius, other.Offset.X, other.Offset.Y)
		//return pen, gomath.Vector2[T]{X: nx, Y: ny}, ok
	case ColliderAABB:
		return CircleAABBCollidersPenetration(circle, offset, other.Shape.(gomath.AABBCollider[T]), other.Offset)
	//case ColliderLineSegment

	default:
		fmt.Printf("Unsupported collider type for circle penetration: %T\n", other.Type)
		return penetration, surfaceNormal, false
	}
	return
}

func ColliderAABBPenetration[T gomath.Float](aabb gomath.AABBCollider[T], offset gomath.Vector2[T], other Collider[T]) (penetration T, surfaceNormal gomath.Vector2[T], overlapping bool) {
	switch other.Type {
	case ColliderAABB:
		otherAABB := other.Shape.(gomath.AABBCollider[T])
		return AABBAABBPenetration(
			gomath.AABB[T]{Center: offset, HalfWidth: aabb.HalfWidth, HalfHeight: aabb.HalfHeight},
			gomath.AABB[T]{Center: other.Offset, HalfWidth: otherAABB.HalfWidth, HalfHeight: otherAABB.HalfHeight},
		)
	case ColliderCircle:
		penetration, surfaceNormal, overlapping = CircleAABBCollidersPenetration(other.Shape.(gomath.CircleCollider[T]), other.Offset, aabb, offset)
		return penetration, surfaceNormal.MultiplyScalar(-1), overlapping
	//case ColliderLineSegment:

	default:
		fmt.Printf("Unsupported collider type for penetration: %T\n", other.Type)
		return penetration, surfaceNormal, false
	}
}

func ColliderTest[T gomath.Float](first Collider[T], second Collider[T]) bool {
	switch first.Type {
	case ColliderCircle:
		return CircleColliderTest(first.Shape.(gomath.CircleCollider[T]), first.Offset, second)
	case ColliderAABB:
		return AABBColliderTest(first.Shape.(gomath.AABBCollider[T]), first.Offset, second)
	case ColliderLineSegment:
		return LineSegmentColliderTest(first.Shape.(gomath.LineSegment[T]), second)

	default:
		fmt.Printf("Unsupported collider type for intersection: %T\n", first.Type)
		return false
	}
}

func LineSegmentColliderTest[T gomath.Float](segment gomath.LineSegment[T], collider Collider[T]) bool {
	switch collider.Type {
	case ColliderLineSegment:
		_, intersecting := LineSegmentLineSegment2D(segment.P1, segment.P2, collider.Shape.(gomath.LineSegment[T]).P1, collider.Shape.(gomath.LineSegment[T]).P2)
		return intersecting
	case ColliderAABB:
		return AABBLineSegmentColliders(collider.Shape.(gomath.AABBCollider[T]), collider.Offset, segment)
	case ColliderCircle:
		return CircleLineSegmentCollider(collider.Shape.(gomath.CircleCollider[T]), collider.Offset, segment)

	default:
		fmt.Printf("Unsupported collider type for line segment intersection: %T\n", collider.Type)
		return false
	}
}

func AABBColliderTest[T gomath.Float](aabb gomath.AABBCollider[T], offset gomath.Vector2[T], other Collider[T]) bool {
	switch other.Type {
	case ColliderCircle:
		return CircleAABBColliders(other.Shape.(gomath.CircleCollider[T]), other.Offset, aabb, offset)
	case ColliderAABB:
		otherAABB := other.Shape.(gomath.AABBCollider[T])
		return AABBAABB(
			gomath.AABB[T]{Center: offset, HalfWidth: aabb.HalfWidth, HalfHeight: aabb.HalfHeight},
			gomath.AABB[T]{Center: other.Offset, HalfWidth: otherAABB.HalfWidth, HalfHeight: otherAABB.HalfHeight},
		)
	case ColliderLineSegment:
		return AABBLineSegmentColliders(aabb, offset, other.Shape.(gomath.LineSegment[T]))

	default:
		fmt.Printf("Unsupported collider type for AABB intersection: %T\n", other.Type)
		return false
	}
}

func CircleColliderTest[T gomath.Float](circle gomath.CircleCollider[T], offset gomath.Vector2[T], other Collider[T]) bool {
	switch other.Type {
	case ColliderCircle:
		otherCircle := other.Shape.(gomath.CircleCollider[T])
		return CircleCircle(
			gomath.Circle[T]{Center: offset, Radius: circle.Radius},
			gomath.Circle[T]{Center: other.Offset, Radius: otherCircle.Radius},
		)
	case ColliderLineSegment:
		return CircleLineSegmentCollider(circle, offset, other.Shape.(gomath.LineSegment[T]))
	case ColliderAABB:
		return CircleAABBColliders(circle, offset, other.Shape.(gomath.AABBCollider[T]), other.Offset)

	default:
		fmt.Printf("Unsupported collider type for circle intersection: %T\n", other.Type)
		return false
	}
}
