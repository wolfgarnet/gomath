package gomath

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestDistanceFromPointToAABB(t *testing.T) {
	tests := map[string]struct {
		point    Vector2[float64]
		aabb     AABB[float64]
		expected float64
	}{
		"PointInsideAABB": {
			point:    Vector2[float64]{0.5, 0.5},
			aabb:     AABB[float64]{Vector2[float64]{0.5, 0.5}, 1.0, 1.0},
			expected: 0.0,
		},
		"PointOnAABBEdge": {
			point:    Vector2[float64]{0.0, 0.0},
			aabb:     AABB[float64]{Vector2[float64]{0.5, 0.5}, 1.0, 1.0},
			expected: 0.0,
		},
		"PointOutsideAABB1": {
			point:    Vector2[float64]{3.0, 2.0},
			aabb:     AABB[float64]{Vector2[float64]{1, 1}, 1, 1},
			expected: 1,
		},
		"PointOutsideAABB1 a bit down": {
			point:    Vector2[float64]{3.0, 0.0},
			aabb:     AABB[float64]{Vector2[float64]{1, 1}, 1, 1},
			expected: 1.0,
		},
		"PointOutsideAABB1 more down": {
			point:    Vector2[float64]{3.0, -2.0},
			aabb:     AABB[float64]{Vector2[float64]{1, 1}, 1, 1},
			expected: math.Sqrt(5.0),
		},
		"PointOutsideAABB down under": {
			point:    Vector2[float64]{1.0, -3.0},
			aabb:     AABB[float64]{Vector2[float64]{1, 1}, 1, 1},
			expected: 3.0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := DistanceFromPointToAABB[float64](test.point, test.aabb)
			assert.Equal(t, test.expected, result)
		})
	}
}
