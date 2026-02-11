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
			got := AABBAABB(test.b1x, test.b1y, test.b2x, test.b2y, test.b1halfwidth, test.b1halfheight, test.b2halfwidth, test.b2halfheight)
			assert.Equal(t, test.overlaps, got)

			// Penetration
			penetration, normal, overlaps := AABBAABBPenetration(test.b1x, test.b1y, test.b2x, test.b2y, test.b1halfwidth, test.b1halfheight, test.b2halfwidth, test.b2halfheight)
			assert.Equal(t, test.overlaps, overlaps)
			if overlaps {
				assert.InDelta(t, test.penetration, penetration, 0.00001)
				assert.InDelta(t, test.normal.X, normal.X, 0.00001)
				assert.InDelta(t, test.normal.Y, normal.Y, 0.00001)
			}
		})
	}
}
