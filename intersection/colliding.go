package intersection

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

func AABBColliderTest[T gomath.Float](aabb gomath.AABBCollider[T], offset gomath.Vector2[T], collider Collider[T]) bool {
	switch collider.Type {
	case ColliderCircle:
		return CircleAABBColliders(collider.Shape.(gomath.CircleCollider[T]), collider.Offset, aabb, offset)
	case ColliderAABB:
		return AABBAABBColliders(aabb, collider.Shape.(gomath.AABBCollider[T]), offset, collider.Offset)
	case ColliderLineSegment:
		return AABBLineSegmentColliders(aabb, offset, collider.Shape.(gomath.LineSegment[T]))

	default:
		fmt.Printf("Unsupported collider type for AABB intersection: %T\n", collider.Type)
		return false
	}
}

func CircleColliderTest[T gomath.Float](circle gomath.CircleCollider[T], offset gomath.Vector2[T], collider Collider[T]) bool {
	switch collider.Type {
	case ColliderCircle:
		return CircleCircleColliders(circle.Radius, collider.Shape.(gomath.CircleCollider[T]).Radius, offset, collider.Shape.(gomath.Circle[T]).Center)
	case ColliderLineSegment:
		return CircleLineSegmentCollider(circle, offset, collider.Shape.(gomath.LineSegment[T]))
	case ColliderAABB:
		return CircleAABBColliders(circle, offset, collider.Shape.(gomath.AABBCollider[T]), collider.Offset)

	default:
		fmt.Printf("Unsupported collider type for circle intersection: %T\n", collider.Type)
		return false
	}
}
