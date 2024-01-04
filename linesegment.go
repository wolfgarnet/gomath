package gomath

import (
	"fmt"
)

type LineSegment[T Number] struct {
	P1, P2 Vector2[T]
}

func (l LineSegment[T]) String() string {
	return fmt.Sprintf("%v -> %v", l.P1, l.P2)
}

func NewLineSegmentXY[T Number](x1, y1, x2, y2 T) LineSegment[T] {
	return LineSegment[T]{Vector2[T]{X: x1, Y: y1}, Vector2[T]{X: x2, Y: y2}}
}

func NewLineSegment[T Number](p1, p2 Vector2[T]) LineSegment[T] {
	return LineSegment[T]{
		P1: p1,
		P2: p2,
	}
}

func (l LineSegment[T]) Ray() Ray[T] {
	return Ray[T]{
		Origin:    l.P1,
		Direction: l.P2.Sub(l.P1).Normalize(),
	}
}

func (l LineSegment[T]) TransformToAxis(axis Vector2[T]) (min, max T) {
	min = l.P1.X*axis.X + l.P1.Y*axis.Y
	max = l.P2.X*axis.X + l.P2.Y*axis.Y
	if min > max {
		min, max = max, min
	}

	return
}

func (l LineSegment[T]) Length() T {
	return l.P1.DistanceTo(l.P2)
}

func (l LineSegment[T]) Direction() Vector2[T] {
	return l.P2.Sub(l.P1)
}

func (l LineSegment[T]) Normal() Vector2[T] {
	return FindNormalVector(l.P1, l.P2)
}

func GetClosestPointOnLineSegment[T Number](lineStart, lineEnd, point Vector2[T]) Vector2[T] {
	lineVector := Vector2[T]{X: lineEnd.X - lineStart.X, Y: lineEnd.Y - lineStart.Y}
	pointVector := Vector2[T]{X: point.X - lineStart.X, Y: point.Y - lineStart.Y}

	t := (pointVector.X*lineVector.X + pointVector.Y*lineVector.Y) / (lineVector.X*lineVector.X + lineVector.Y*lineVector.Y)

	if t < 0 {
		return lineStart
	} else if t > 1 {
		return lineEnd
	} else {
		return Vector2[T]{X: lineStart.X + t*lineVector.X, Y: lineStart.Y + t*lineVector.Y}
	}
}
