package collider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
)

func TestAABBPoint(t *testing.T) {
	tests := map[string]struct {
		ax, ay, halfwidth, halfheight float32
		px, py                        float32
		expected                      bool
	}{
		"inside": {
			ax: 0, ay: 0, halfwidth: 5, halfheight: 5,
			px: 0, py: 0,
			expected: true,
		},
		"outside x": {
			ax: 0, ay: 0, halfwidth: 5, halfheight: 5,
			px: 6, py: 0,
			expected: false,
		},
		"outside y": {
			ax: 0, ay: 0, halfwidth: 5, halfheight: 5,
			px: 0, py: 6,
			expected: false,
		},
		"edge": {
			ax: 0, ay: 0, halfwidth: 1, halfheight: 1,
			px: 0, py: 1,
			expected: true,
		},
		"edge+": {
			ax: 0, ay: 0, halfwidth: 1, halfheight: 1,
			px: 0, py: 1.000001,
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := AABBPoint(tc.ax, tc.ay, tc.halfwidth, tc.halfheight, tc.px, tc.py)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestAABBAABB(t *testing.T) {
	tests := map[string]struct {
		b1x, b1y, b2x, b2y, b1halfwidth, b1halfheight, b2halfwidth, b2halfheight float32
		penetration                                                              float32
		normal                                                                   gomath.Vector2[float32]
		overlaps                                                                 bool
	}{
		"on top of each other": {
			b1x: 0, b1y: 0, b2x: 0, b2y: 0, b1halfwidth: 1, b1halfheight: 1, b2halfwidth: 1, b2halfheight: 1,
			overlaps:    true,
			penetration: 2.0,
			normal:      gomath.Vector2[float32]{X: 0, Y: 1},
		},
		"next to each other": {
			b1x: 0, b1y: 0, b2x: 3, b2y: 0, b1halfwidth: 1, b1halfheight: 1, b2halfwidth: 1, b2halfheight: 1,
			overlaps: false,
		},
		"next to each other overlaps": {
			b1x: 0, b1y: 0, b2x: 1.5, b2y: 0, b1halfwidth: 1, b1halfheight: 1, b2halfwidth: 1, b2halfheight: 1,
			overlaps:    true,
			penetration: 0.5,
			normal:      gomath.Vector2[float32]{X: -1, Y: 0},
		},
		"over each other": {
			b1x: 0, b1y: 0, b2x: 0, b2y: 3, b1halfwidth: 1, b1halfheight: 1, b2halfwidth: 1, b2halfheight: 1,
			overlaps: false,
		},
		"over each other overlaps": {
			b1x: 0, b1y: 0, b2x: 0, b2y: 1.5, b1halfwidth: 1, b1halfheight: 1, b2halfwidth: 1, b2halfheight: 1,
			overlaps:    true,
			penetration: 0.5,
			normal:      gomath.Vector2[float32]{X: 0, Y: -1},
		},
		"diagonal": {
			b1x: 0, b1y: 0, b2x: 2, b2y: 3, b1halfwidth: 1, b1halfheight: 1, b2halfwidth: 1, b2halfheight: 1,
			overlaps: false,
		},
		"diagonal overlaps": {
			b1x: 0, b1y: 0, b2x: 1.5, b2y: 1.2, b1halfwidth: 1, b1halfheight: 1, b2halfwidth: 1, b2halfheight: 1,
			overlaps:    true,
			penetration: 0.5,
			normal:      gomath.Vector2[float32]{X: -1, Y: 0},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := AABBAABB(test.b1x, test.b1y, test.b1halfwidth, test.b1halfheight, test.b2x, test.b2y, test.b2halfwidth, test.b2halfheight)
			assert.Equal(t, test.overlaps, got)

			// Penetration
			penetration, normal, overlaps := AABBAABBPenetration(test.b1x, test.b1y, test.b1halfwidth, test.b1halfheight, test.b2x, test.b2y, test.b2halfwidth, test.b2halfheight)
			assert.Equal(t, test.overlaps, overlaps)
			if overlaps {
				assert.InDelta(t, test.penetration, penetration, 0.00001)
				assert.InDelta(t, test.normal.X, normal.X, 0.00001)
				assert.InDelta(t, test.normal.Y, normal.Y, 0.00001)
			}
		})
	}
}

func TestAABB2LineSegment(t *testing.T) {
	tests := map[string]struct {
		minx, miny, maxx, maxy float32
		ls                     gomath.LineSegment[float32]
		expected               bool
	}{
		"inside": {
			minx: 0, miny: 0, maxx: 1, maxy: 1,
			ls: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 0.5, Y: 0.5},
				P2: gomath.Vector2[float32]{X: 0.5, Y: 0.5},
			},
			expected: true,
		},
		"outside": {
			minx: 0, miny: 0, maxx: 1, maxy: 1,
			ls: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 2, Y: 2},
				P2: gomath.Vector2[float32]{X: 2, Y: 2},
			},
			expected: false,
		},
		"vertical": {
			minx: 0, miny: 0, maxx: 1, maxy: 1,
			ls: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 0.5, Y: 0},
				P2: gomath.Vector2[float32]{X: 0.5, Y: 1},
			},
			expected: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			intersects := AABB2LineSegment[float32](test.minx, test.miny, test.maxx, test.maxy, test.ls.P1.X, test.ls.P1.Y, test.ls.P2.X, test.ls.P2.Y)
			assert.Equal(t, test.expected, intersects)
		})
	}
}

func TestAABBLineSegmentPenetration(t *testing.T) {
	tests := map[string]struct {
		ax, ay, hw, hh float32
		sx, sy, ex, ey float32
		overlaps       bool
		penetration    float32
		normal         gomath.Vector2[float32]
	}{
		"outside": {
			ax: 0, ay: 0, hw: 1, hh: 1,
			sx: 2, sy: 0, ex: 3, ey: 0,
			overlaps: false,
		},
		"inside vertical": {
			ax: 0, ay: 0, hw: 5, hh: 5,
			sx: 0, sy: -1, ex: 0, ey: 1,
			overlaps: true,
			// my code for both points inside:
			// midX=0, midY=0. dx=0, dy=0. px=5, py=5.
			// px < py is false (5 < 5). So else: penetration = py = 5, normal = {0, 1}
			penetration: 5,
			normal:      gomath.Vector2[float32]{X: 0, Y: 1},
		},
		"intersecting from right": {
			ax: 0, ay: 0, hw: 1, hh: 1,
			sx: 2, sy: 0, ex: 0.5, ey: 0,
			overlaps:    true,
			penetration: 0.5, // point 0.5, 0 is inside. dx=0.5, dy=0. px=1-0.5=0.5, py=1-0=1. px < py. penetration=0.5, normal={1,0}
			normal:      gomath.Vector2[float32]{X: 1, Y: 0},
		},
		"crossing horizontal": {
			ax: 0, ay: 0, hw: 1, hh: 1,
			sx: -2, sy: 0, ex: 2, ey: 0,
			overlaps:    true,
			penetration: 1, // closest point to center is 0,0. px=1, py=1. px < py is false. penetration=1, normal={0,1}
			normal:      gomath.Vector2[float32]{X: 0, Y: 1},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			penetration, normal, overlaps := AABBLineSegmentPenetration(test.ax, test.ay, test.hw, test.hh, test.sx, test.sy, test.ex, test.ey)
			assert.Equal(t, test.overlaps, overlaps)
			if overlaps {
				assert.InDelta(t, test.penetration, penetration, 0.0001)
				assert.InDelta(t, test.normal.X, normal.X, 0.0001)
				assert.InDelta(t, test.normal.Y, normal.Y, 0.0001)
			}
		})
	}
}
