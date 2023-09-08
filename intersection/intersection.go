package intersection

import (
	"fmt"
	"github.com/wolfgarnet/gomath"
)

func IntersectingSquaredAABB[T gomath.Number](aabb gomath.SquaredAABB2D[T], other interface{}) bool {
	switch o := other.(type) {
	case gomath.Circle[T]:
		return CircleAABB(o, aabb)
	case gomath.LineSegment[T]:
		return LineSegmentInsideOrIntersectsSquaredAABB(o, aabb)
	case gomath.Vector2[T]:
		return AABB2DPoint(aabb, o)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func IntersectingCircle[T gomath.Number](circle gomath.Circle[T], other interface{}) bool {
	switch t := other.(type) {
	case gomath.SquaredAABB2D[T]:
		return CircleAABB(circle, t)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", other))
	}
}

func Intersecting2D[T gomath.Number](first, second interface{}) bool {
	switch f1 := first.(type) {
	case gomath.SquaredAABB2D[T]:
		return IntersectingSquaredAABB(f1, second)
	case gomath.Circle[T]:
		return IntersectingCircle(f1, second)
	case gomath.LineSegment[T]:
		return IntersectingLineSegment2D(f1, second)
	case gomath.Vector2[T]:
		return IntersectingPoint2D(f1, second)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", first))
	}
}
