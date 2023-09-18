package intersection

import (
	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
	"testing"
)

func TestCircleCircle(t *testing.T) {
	tests := map[string]struct {
		circle1     gomath.Circle[float32]
		circle2     gomath.Circle[float32]
		overlapping bool
		penetration float32
		normal      gomath.Vector2[float32]
	}{
		"from right": {
			circle1:     gomath.NewCircle[float32](0, 0, 1),
			circle2:     gomath.NewCircle[float32](2, 0, 1.1),
			overlapping: true,
			penetration: 0.1,
			normal:      gomath.NewVector2[float32](1, 0),
		},
		"from right no overlap": {
			circle1:     gomath.NewCircle[float32](0, 0, 1),
			circle2:     gomath.NewCircle[float32](2, 0, 0.9),
			overlapping: false,
		},
		"from top": {
			circle1:     gomath.NewCircle[float32](0, 0, 1),
			circle2:     gomath.NewCircle[float32](0, 2, 1.1),
			overlapping: true,
			penetration: 0.1,
			normal:      gomath.NewVector2[float32](0, 1),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			penetration, normal, overlapping := CircleCircleWithPenetration(test.circle1, test.circle2)
			assert.Equal(t, test.overlapping, overlapping)
			if !overlapping {
				return
			}

			assert.InDelta(t, test.penetration, penetration, 0.001)
			assert.Equal(t, test.normal, normal)
		})
	}
}

func TestCircleAABB(t *testing.T) {
	tests := map[string]struct {
		circle      gomath.Circle[float32]
		box         gomath.SquaredAABB2D[float32]
		overlapping bool
		penetration float32
		normal      gomath.Vector2[float32]
	}{
		"box from right": {
			circle:      gomath.NewCircle[float32](0, 0, 1),
			box:         gomath.NewSquaredAABB2DXY[float32](2, 0, 1.1),
			overlapping: true,
			penetration: 0.1,
			normal:      gomath.NewVector2[float32](1, 0),
		},
		"box from top": {
			circle:      gomath.NewCircle[float32](0, 0, 1),
			box:         gomath.NewSquaredAABB2DXY[float32](0, 2, 1.1),
			overlapping: true,
			penetration: 0.1,
			normal:      gomath.NewVector2[float32](0, 1),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			penetration, normal, overlapping := CircleSquaredAABBWithPenetration(test.circle, test.box)
			assert.Equal(t, test.overlapping, overlapping)
			if !overlapping {
				return
			}

			assert.InDelta(t, test.penetration, penetration, 0.001)
			assert.Equal(t, test.normal, normal)
		})
	}
}
