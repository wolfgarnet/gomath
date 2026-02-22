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

// GetClosestPointOnLineSegment returns the point on the line segment closest for the given point.
func GetClosestPointOnLineSegment[T Float](segmentStart, segmentEnd, point Vector2[T]) Vector2[T] {
	dx := segmentEnd.X - segmentStart.X
	dy := segmentEnd.Y - segmentStart.Y

	den := dx*dx + dy*dy
	if den == 0 {
		// Segment is a point.
		return segmentStart
	}

	px := point.X - segmentStart.X
	py := point.Y - segmentStart.Y

	t := (px*dx + py*dy) / den

	// Clamp t to [0, 1]
	if t <= 0 {
		return segmentStart
	}
	if t >= 1 {
		return segmentEnd
	}

	return Vector2[T]{
		X: segmentStart.X + t*dx,
		Y: segmentStart.Y + t*dy,
	}
}
