package collider

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wolfgarnet/gomath"
)

func circle(x, y, r float32) gomath.Circle[float32] {
	return gomath.Circle[float32]{Center: gomath.Vector2[float32]{X: x, Y: y}, Radius: r}
}

func aabb(x, y, hw, hh float32) gomath.AABB[float32] {
	return gomath.AABB[float32]{Center: gomath.Vector2[float32]{X: x, Y: y}, HalfWidth: hw, HalfHeight: hh}
}

func seg(x1, y1, x2, y2 float32) gomath.LineSegment[float32] {
	return gomath.LineSegment[float32]{P1: gomath.Vector2[float32]{X: x1, Y: y1}, P2: gomath.Vector2[float32]{X: x2, Y: y2}}
}

// These pin the struct penetration functions against the same expected values as
// the former collider package's tests, confirming the byte-for-byte port.

func TestCircleCirclePenetration(t *testing.T) {
	p, n, o := CircleCirclePenetration(circle(0, 0, 1), circle(1.5, 0, 1))
	assert.True(t, o)
	assert.InDelta(t, 0.5, p, 1e-4)
	assert.InDelta(t, 1, n.X, 1e-4)
	assert.InDelta(t, 0, n.Y, 1e-4)

	_, _, o = CircleCirclePenetration(circle(0, 0, 1), circle(3, 0, 1))
	assert.False(t, o)
}

func TestCircleAABBPenetration(t *testing.T) {
	// From the left.
	p, n, o := CircleAABBPenetration(circle(-1.5, 0, 1), aabb(0, 0, 1, 1))
	assert.True(t, o)
	assert.InDelta(t, 0.5, p, 1e-4)
	assert.InDelta(t, 1, n.X, 1e-4)

	// From the right.
	_, n, _ = CircleAABBPenetration(circle(5, 0, 1), aabb(0, 0, 5, 5))
	assert.InDelta(t, -1, n.X, 1e-4)

	// Centre inside.
	p, n, o = CircleAABBPenetration(circle(0, 0, 1), aabb(0, 0, 5, 5))
	assert.True(t, o)
	assert.InDelta(t, 6, p, 1e-4)
	assert.InDelta(t, 1, n.Y, 1e-4)
}

func TestAABBAABBPenetration(t *testing.T) {
	p, n, o := AABBAABBPenetration(aabb(0, 0, 1, 1), aabb(1.5, 0, 1, 1))
	assert.True(t, o)
	assert.InDelta(t, 0.5, p, 1e-4)
	assert.InDelta(t, 1, n.X, 1e-4)

	_, n, _ = AABBAABBPenetration(aabb(0, 0, 1, 1), aabb(0, 1.5, 1, 1))
	assert.InDelta(t, 1, n.Y, 1e-4)

	_, _, o = AABBAABBPenetration(aabb(0, 0, 1, 1), aabb(3, 0, 1, 1))
	assert.False(t, o)
}

func TestCircleLineSegmentPenetration(t *testing.T) {
	// Segment point directly above the circle centre; normal points up toward it.
	p, n, o := CircleLineSegmentPenetration(circle(0, 0, 2), seg(0, 1, 0, 1))
	assert.True(t, o)
	assert.InDelta(t, 1, p, 1e-4)
	assert.InDelta(t, 0, n.X, 1e-4)
	assert.InDelta(t, 1, n.Y, 1e-4)
}

func TestAABBLineSegmentPenetration(t *testing.T) {
	// Segment crossing the box overlaps; a segment well outside does not.
	_, _, o := AABBLineSegmentPenetration(aabb(0, 0, 1, 1), seg(-2, 0, 2, 0))
	assert.True(t, o)

	_, _, o = AABBLineSegmentPenetration(aabb(0, 0, 1, 1), seg(5, 5, 6, 6))
	assert.False(t, o)
}
