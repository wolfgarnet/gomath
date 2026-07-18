package collider

import (
	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
	"testing"
)

func newEllipsoid(x, y float32, a, b float32) gomath.Ellipsoid[float32] {
	return gomath.Ellipsoid[float32]{
		Position: gomath.Vector2[float32]{
			X: x,
			Y: y,
		},
		Radii: gomath.Vector2[float32]{
			X: a,
			Y: b,
		},
	}
}

func TestEllipsoidEllipsoid(t *testing.T) {
	tests := map[string]struct {
		e1       gomath.Ellipsoid[float32]
		e2       gomath.Ellipsoid[float32]
		overlaps bool
	}{
		"same position": {
			e1:       newEllipsoid(0, 0, 1, 1),
			e2:       newEllipsoid(0, 0, 1, 1),
			overlaps: true,
		},
		"same y": {
			e1:       newEllipsoid(0, 0, 1, 1),
			e2:       newEllipsoid(1, 0, 1, 1),
			overlaps: true,
		},
		"same y too far": {
			e1:       newEllipsoid(0, 0, 1, 1),
			e2:       newEllipsoid(3, 0, 1, 1),
			overlaps: false,
		},
		"diagonal": {
			e1:       newEllipsoid(0, 0, 1, 1),
			e2:       newEllipsoid(1, 1, 1, 1),
			overlaps: true,
		},
		"diagonal too far": {
			e1:       newEllipsoid(0, 0, 1, 1),
			e2:       newEllipsoid(3, 3, 1, 1),
			overlaps: false,
		},
		"diagonal too far 2": {
			e1:       newEllipsoid(0, 0, 3, 1),
			e2:       newEllipsoid(3, 3, 1, 1),
			overlaps: false,
		},
		"diagonal far": {
			e1:       newEllipsoid(0, 0, 3, 1),
			e2:       newEllipsoid(3, 3, 1, 3),
			overlaps: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			overlaps := EllipsoidEllipsoid(test.e1, test.e2)
			assert.Equal(t, test.overlaps, overlaps)
		})
	}
}
