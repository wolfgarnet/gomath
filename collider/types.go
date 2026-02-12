package collider

type Type uint8

const (
	AABB Type = iota
	Circle
	LineSegment
)
