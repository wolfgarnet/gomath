package gomath

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetClosestPointOnLineSegment(t *testing.T) {
	tests := map[string]struct {
		start   Vector2[float32]
		end     Vector2[float32]
		point   Vector2[float32]
		closest Vector2[float32]
	}{
		"on": {
			start:   Vector2[float32]{X: 0, Y: 0},
			end:     Vector2[float32]{X: 2, Y: 0},
			point:   Vector2[float32]{X: 1, Y: 0},
			closest: Vector2[float32]{X: 1, Y: 0},
		},
		"too far +x": {
			start:   Vector2[float32]{X: 0, Y: 0},
			end:     Vector2[float32]{X: 2, Y: 0},
			point:   Vector2[float32]{X: 3, Y: 0},
			closest: Vector2[float32]{X: 2, Y: 0},
		},
		"too far -x": {
			start:   Vector2[float32]{X: 0, Y: 0},
			end:     Vector2[float32]{X: 2, Y: 0},
			point:   Vector2[float32]{X: -1, Y: 0},
			closest: Vector2[float32]{X: 0, Y: 0},
		},
		"too far +y": {
			start:   Vector2[float32]{X: 0, Y: 0},
			end:     Vector2[float32]{X: 2, Y: 0},
			point:   Vector2[float32]{X: 0, Y: 1},
			closest: Vector2[float32]{X: 0, Y: 0},
		},
		"too far +y middle": {
			start:   Vector2[float32]{X: 0, Y: 0},
			end:     Vector2[float32]{X: 2, Y: 0},
			point:   Vector2[float32]{X: 1, Y: 1},
			closest: Vector2[float32]{X: 1, Y: 0},
		},
		"too far -y": {
			start:   Vector2[float32]{X: 0, Y: 0},
			end:     Vector2[float32]{X: 2, Y: 0},
			point:   Vector2[float32]{X: -1, Y: 1},
			closest: Vector2[float32]{X: 0, Y: 0},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			closest := GetClosestPointOnLineSegment(test.start, test.end, test.point)
			assert.InDelta(t, test.closest.X, closest.X, 0.001)
			assert.InDelta(t, test.closest.Y, closest.Y, 0.001)
		})
	}
}
