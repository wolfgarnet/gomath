package collider

import (
	"fmt"
	"github.com/wolfgarnet/gomath"
)

func IntegerIntersecting(a, b any) bool {
	switch t := a.(type) {
	case gomath.AABB[int]:
		return IntegerIntersectionAABB(t, b)
	case gomath.Vector2[int]:
		return IntegerIntersectionVector(t, b)
	case gomath.SquaredAABB2D[int]:
		return IntegerIntersectionSquaredAABB(t, b)
	case gomath.LineSegment[int]:
		return IntegerIntersectionLineSegment(t, b)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", a))
	}
}

func IntegerIntersectionAABB(aabb gomath.AABB[int], b any) bool {
	switch t := b.(type) {
	case gomath.AABB[int]:
		return AABBAABB(aabb, t)
	case gomath.Vector2[int]:
		return AABBPoint(aabb, t)
	case gomath.SquaredAABB2D[int]:
		return AABBSquaredAABB(aabb, t)
	case gomath.LineSegment[int]:
		return LineBox(t, aabb)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", aabb))
	}
}

func IntegerIntersectionVector(point gomath.Vector2[int], b any) bool {
	switch t := b.(type) {
	case gomath.AABB[int]:
		return AABBPoint(t, point)
	case gomath.Vector2[int]:
		return point.X == t.X && point.Y == t.Y
	case gomath.SquaredAABB2D[int]:
		return SquaredAABBPoint(t, point)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", point))
	}
}

func IntegerIntersectionSquaredAABB(aabb gomath.SquaredAABB2D[int], b any) bool {
	switch t := b.(type) {
	case gomath.AABB[int]:
		return AABBSquaredAABB(t, aabb)
	case gomath.Vector2[int]:
		return SquaredAABBPoint(aabb, t)
	case gomath.SquaredAABB2D[int]:
		return SquaredAABBSquaredAABB(aabb, t)
	case gomath.LineSegment[int]:
		return LineSquaredBox(t, aabb)
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", aabb))
	}
}

func IntegerIntersectionLineSegment(segment gomath.LineSegment[int], b any) bool {
	switch t := b.(type) {
	case gomath.AABB[int]:
		return LineBox(segment, t)
	case gomath.SquaredAABB2D[int]:
		return LineSquaredBox(segment, t)
	case gomath.LineSegment[int]:
		_, intersects := LineSegmentLineSegment2D(segment.P1, segment.P2, t.P1, t.P2)
		return intersects
	default:
		panic(fmt.Sprintf("Unsupported operand: %T", segment))
	}
}
