package collider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircleCircle(t *testing.T) {
	tests := map[string]struct {
		r1, r2              float32
		o1x, o1y, o2x, o2y  float32
		expected            bool
		expectedPenetration float32
		expectedNormalX     float32
		expectedNormalY     float32
	}{
		"not intersecting": {
			r1: 1, r2: 1, o1x: 0, o1y: 0, o2x: 3, o2y: 0,
			expected: false,
		},
		"intersecting": {
			r1: 1, r2: 1, o1x: 0, o1y: 0, o2x: 1.5, o2y: 0,
			expected:            true,
			expectedPenetration: 0.5,
			expectedNormalX:     1,
			expectedNormalY:     0,
		},
		"touching": {
			r1: 1, r2: 1, o1x: 0, o1y: 0, o2x: 2, o2y: 0,
			expected: false, // dist < r1+r2
		},
		"same position": {
			r1: 1, r2: 1, o1x: 0, o1y: 0, o2x: 0, o2y: 0,
			expected:            true,
			expectedPenetration: 2,
			expectedNormalX:     0,
			expectedNormalY:     0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := CircleCircle(tc.r1, tc.r2, tc.o1x, tc.o1y, tc.o2x, tc.o2y)
			assert.Equal(t, tc.expected, got)

			pen, nx, ny, ok := CircleCirclePenetration(tc.r1, tc.r2, tc.o1x, tc.o1y, tc.o2x, tc.o2y)
			assert.Equal(t, tc.expected, ok)
			if ok {
				assert.InDelta(t, tc.expectedPenetration, pen, 0.0001)
				assert.InDelta(t, tc.expectedNormalX, nx, 0.0001)
				assert.InDelta(t, tc.expectedNormalY, ny, 0.0001)
			}
		})
	}
}

func TestCircleAABB(t *testing.T) {
	tests := map[string]struct {
		r                   float32
		cx, cy              float32
		hw, hh              float32
		ax, ay              float32
		expected            bool
		expectedPenetration float32
		expectedNormalX     float32
		expectedNormalY     float32
	}{
		"not intersecting": {
			r: 1, cx: -2, cy: 0,
			hw: 1, hh: 1, ax: 0, ay: 0,
			expected: false,
		},
		"intersecting from left": {
			r: 1, cx: -1.5, cy: 0,
			hw: 1, hh: 1, ax: 0, ay: 0,
			expected:            true,
			expectedPenetration: 0.5,
			expectedNormalX:     -1,
			expectedNormalY:     0,
		},
		"inside AABB": {
			r: 1, cx: 0, cy: 0,
			hw: 5, hh: 5, ax: 0, ay: 0,
			expected:            true,
			expectedPenetration: 6, // hw - |cx-ax| + r = 5 - 0 + 1 = 6
			expectedNormalX:     0,
			expectedNormalY:     -1,
		},
		"inside AABB slightly up": {
			r: 1, cx: 0, cy: 1,
			hw: 5, hh: 5, ax: 0, ay: 0,
			expected:            true,
			expectedPenetration: 5, // hw - |cx-ax| + r = 5 - 0 + 1 = 6
			expectedNormalX:     0,
			expectedNormalY:     1,
		},
		"inside AABB slightly down": {
			r: 1, cx: 0, cy: -1,
			hw: 5, hh: 5, ax: 0, ay: 0,
			expected:            true,
			expectedPenetration: 5, // hw - |cx-ax| + r = 5 - 0 + 1 = 6
			expectedNormalX:     0,
			expectedNormalY:     -1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := CircleAABB(tc.r, tc.cx, tc.cy, tc.hw, tc.hh, tc.ax, tc.ay)
			assert.Equal(t, tc.expected, got)

			pen, nx, ny, ok := CircleAABBPenetration(tc.r, tc.cx, tc.cy, tc.hw, tc.hh, tc.ax, tc.ay)
			assert.Equal(t, tc.expected, ok)
			if ok {
				assert.InDelta(t, tc.expectedPenetration, pen, 0.0001, "penetration")
				assert.InDelta(t, tc.expectedNormalX, nx, 0.0001, "normalX")
				assert.InDelta(t, tc.expectedNormalY, ny, 0.0001, "normalY")
			}
		})
	}
}

func TestCircleLineSegmentPenetration(t *testing.T) {
	tests := map[string]struct {
		circleRadius         float32
		circlex, circley     float32
		segment1x, segment1y float32
		segment2x, segment2y float32
		overlapping          bool
		penetration          float32
		normalx, normaly     float32
	}{
		"segment over circle overlapping": {
			circleRadius: 2,
			circlex:      0, circley: 0,
			segment1x: 0, segment1y: 1,
			segment2x: 0, segment2y: 1,
			overlapping: true,
			penetration: 1,
			normalx:     0, normaly: -1,
		},
		"segment slightly under circle overlapping": {
			circleRadius: 1.1,
			circlex:      0, circley: 0,
			segment1x: 0, segment1y: -1,
			segment2x: 0, segment2y: -1,
			overlapping: true,
			penetration: 0.1,
			normalx:     0, normaly: 1,
		},
		"segment over left overlapping": {
			circleRadius: 1.0,
			circlex:      0, circley: 0,
			segment1x: 0, segment1y: 1,
			segment2x: 1, segment2y: 0,
			overlapping: true,
			penetration: 0.29289323,
			normalx:     -0.70710677, normaly: -0.70710677,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			penetration, normalx, normaly, overlapping := CircleLineSegmentPenetration(test.circleRadius, test.circlex, test.circley, test.segment1x, test.segment1y, test.segment2x, test.segment2y)
			require.Equal(t, test.overlapping, overlapping)
			if overlapping {
				assert.InDelta(t, test.penetration, penetration, 0.0001, "penetration")
				assert.InDelta(t, test.normalx, normalx, 0.0001, "normalx")
				assert.InDelta(t, test.normaly, normaly, 0.0001, "normaly")
			}
		})
	}
}
