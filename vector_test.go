package gomath

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestVector2_Rotate(t *testing.T) {
	tests := map[string]struct {
		vector   Vector2[float32]
		angle    float32
		expected Vector2[float32]
	}{
		"up 90": {
			vector:   Vector2[float32]{X: 0, Y: 1},
			angle:    math.Pi / 2,
			expected: Vector2[float32]{X: -1, Y: 0},
		},
		"up -90": {
			vector:   Vector2[float32]{X: 0, Y: 1},
			angle:    -math.Pi / 2,
			expected: Vector2[float32]{X: 1, Y: 0},
		},
		"right 90": {
			vector:   Vector2[float32]{X: 1, Y: 0},
			angle:    math.Pi / 2,
			expected: Vector2[float32]{X: 0, Y: 1},
		},
		"right -90": {
			vector:   Vector2[float32]{X: 1, Y: 0},
			angle:    -math.Pi / 2,
			expected: Vector2[float32]{X: 0, Y: -1},
		},

		"up 45": {
			vector:   Vector2[float32]{X: 0, Y: 1},
			angle:    math.Pi / 4,
			expected: Vector2[float32]{X: -0.707, Y: 0.707},
		},

		"up 180": {
			vector:   Vector2[float32]{X: 0, Y: 1},
			angle:    math.Pi,
			expected: Vector2[float32]{X: 0, Y: -1},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rotated := test.vector.Rotate(test.angle)
			spew.Dump(rotated)
			assert.InDelta(t, test.expected.X, rotated.X, 0.001)
			assert.InDelta(t, test.expected.Y, rotated.Y, 0.001)
		})
	}
}
