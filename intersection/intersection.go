package intersection

import (
	"fmt"
	"github.com/wolfgarnet/gomath"
)

func IntersectingAABB[T gomath.Float](aabb gomath.AABB[T], other interface{}) bool {
	switch o := other.(type) {
	case gomath.AABB[T]:
		return AABBAABB(aabb, o)
	case gomath.SquaredAABB2D[T]:
		return AABBSquaredAABB(aabb, o)
	case gomath.Circle[T]:
		return CircleAABB(o, aabb)
	case gomath.LineSegment[T]:
		return LineSegmentInsideOrIntersectsAABB(o, aabb)
	case gomath.Vector2[T]:
		return AABBPoint(aabb, o)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func IntersectingSquaredAABB[T gomath.Float](saabb gomath.SquaredAABB2D[T], other interface{}) bool {
	switch o := other.(type) {
	case gomath.SquaredAABB2D[T]:
		return SquaredAABBSquaredAABB(saabb, o)
	case gomath.Circle[T]:
		return CircleSquaredAABB(o, saabb)
	case gomath.AABB[T]:
		return AABBSquaredAABB(o, saabb)
	case gomath.LineSegment[T]:
		return LineSegmentInsideOrIntersectsSquaredAABB(o, saabb)
	case gomath.Vector2[T]:
		return SquaredAABBPoint(saabb, o)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func IntersectingCircle[T gomath.Number](circle gomath.Circle[T], other interface{}) bool {
	switch t := other.(type) {
	case gomath.SquaredAABB2D[T]:
		return CircleSquaredAABB(circle, t)
	case gomath.AABB[T]:
		return CircleAABB(circle, t)
	case gomath.Circle[T]:
		return CircleCircle(circle, t)
	case gomath.LineSegment[T]:
		return CircleLineSegment(circle, t)
	case gomath.Vector2[T]:
		return CirclePoint(circle, t)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func IntersectingLineSegment[T gomath.Float](segment gomath.LineSegment[T], other interface{}) bool {
	switch t := other.(type) {
	case gomath.SquaredAABB2D[T]:
		return LineSegmentInsideOrIntersectsSquaredAABB(segment, t)
	case gomath.Circle[T]:
		return CircleLineSegment(t, segment)
	case gomath.AABB[T]:
		return LineSegmentInsideOrIntersectsAABB(segment, t)
	case gomath.LineSegment[T]:
		_, intersects := LineSegmentLineSegment2D(segment.P1, segment.P2, t.P1, t.P2)
		return intersects
	//case gomath.Vector2[T]:
	//return Linesegment
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func IntersectingPoint[T gomath.Float](vector gomath.Vector2[T], other interface{}) bool {
	switch t := other.(type) {
	case gomath.SquaredAABB2D[T]:
		return SquaredAABBPoint(t, vector)
	case gomath.Circle[T]:
		return CirclePoint(t, vector)
	case gomath.AABB[T]:
		return AABBPoint(t, vector)
	//case gomath.LineSegment[T]:
	//case gomath.Vector2[T]:
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func Intersecting2D[T gomath.Float](first, second interface{}) bool {
	switch f1 := first.(type) {
	case gomath.SquaredAABB2D[T]:
		return IntersectingSquaredAABB(f1, second)
	case gomath.AABB[T]:
		return IntersectingAABB(f1, second)
	case gomath.Circle[T]:
		return IntersectingCircle(f1, second)
	case gomath.LineSegment[T]:
		return IntersectingLineSegment(f1, second)
	case gomath.Vector2[T]:
		return IntersectingPoint(f1, second)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", first))
	}
}

func IntersectingSquaredAABB2[T gomath.Float](aabb gomath.SquaredAABB2D[T], other interface{}) (penetration T, surfaceNormal gomath.Vector2[T], overlapping bool) {
	switch o := other.(type) {
	case gomath.SquaredAABB2D[T]:
		return SquaredAABB2DSquaredAABB2DWithPenetration(aabb, o)
	case gomath.AABB[T]:
		return AABBSquaredAABB2DWithPenetration(o, aabb)
	case gomath.Circle[T]:
		penetration, surfaceNormal, overlapping = CircleSquaredAABBWithPenetration(o, aabb)
		return penetration, surfaceNormal.MultiplyScalar(-1), overlapping
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func IntersectingAABB2[T gomath.Float](aabb gomath.AABB[T], other interface{}) (penetration T, surfaceNormal gomath.Vector2[T], overlapping bool) {
	switch o := other.(type) {
	case gomath.SquaredAABB2D[T]:
		return AABBSquaredAABB2DWithPenetration(aabb, o)
	case gomath.AABB[T]:
		return AABB2DAABB2DWithPenetration(aabb, o)
	case gomath.Circle[T]:
		penetration, surfaceNormal, overlapping = CircleAABBWithPenetration(o, aabb)
		return penetration, surfaceNormal.MultiplyScalar(-1), overlapping
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func IntersectingCircle2[T gomath.Float](circle gomath.Circle[T], other interface{}) (penetration T, surfaceNormal gomath.Vector2[T], overlapping bool) {
	switch o := other.(type) {
	case gomath.SquaredAABB2D[T]:
		return CircleSquaredAABBWithPenetration(circle, o)
	case gomath.Circle[T]:
		return CircleCircleWithPenetration(circle, o)
	case gomath.AABB[T]:
		penetration, surfaceNormal, overlapping = CircleAABBWithPenetration(circle, o)
		return penetration, surfaceNormal.MultiplyScalar(-1), overlapping
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func Intersecting2[T gomath.Float](first, second interface{}) (penetration T, surfaceNormal gomath.Vector2[T], overlapping bool) {
	switch f1 := first.(type) {
	case gomath.SquaredAABB2D[T]:
		return IntersectingSquaredAABB2(f1, second)
	case gomath.AABB[T]:
		return IntersectingAABB2(f1, second)
	case gomath.Circle[T]:
		return IntersectingCircle2(f1, second)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", first))
	}
}
