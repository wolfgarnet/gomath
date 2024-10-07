package gomath

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRay_AtDistance(t *testing.T) {
	tests := map[string]struct {
		ray      Ray[float32]
		distance float32
		length   float32
		expected LineSegment[float32]
	}{
		"basic": {
			ray:      NewRay[float32](Vector2[float32]{}, Vector2[float32]{X: 1}),
			distance: 1,
			length:   1,
			expected: LineSegment[float32]{P1: Vector2[float32]{X: 1, Y: 0}, P2: Vector2[float32]{X: 2, Y: 0}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			segment := LineSegment[float32]{}
			test.ray.AtDistance(test.distance, test.length, &segment)
			assert.Equal(t, test.expected, segment)
		})
	}
}
