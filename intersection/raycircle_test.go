package intersection

import (
	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
	"testing"
)

func TestRayCircleIntersection(t *testing.T) {
	tests := map[string]struct {
		ray      gomath.Ray[float32]
		circle   gomath.Circle[float32]
		p1       *gomath.Vector2[float32]
		p2       *gomath.Vector2[float32]
		expected int
	}{
		"Intersecting": {
			ray:      gomath.NewRay(gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{1, 0}),
			circle:   gomath.NewCircle[float32](3, 0, 2),
			p1:       &gomath.Vector2[float32]{5, 0},
			p2:       &gomath.Vector2[float32]{1, 0},
			expected: 2,
		},
		"Not intersecting": {
			ray:      gomath.NewRay(gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{1, 0}),
			circle:   gomath.NewCircle[float32](5, 5, 1),
			expected: 0,
		},
		"Tangent": {
			ray:      gomath.NewRay(gomath.Vector2[float32]{0, 1}, gomath.Vector2[float32]{1, 0}),
			circle:   gomath.NewCircle[float32](2, 0, 1),
			p1:       &gomath.Vector2[float32]{X: 2, Y: 1},
			expected: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			intersects, p1, p2 := RayCircle(test.ray, test.circle)
			assert.Equal(t, test.expected, intersects)
			if intersects > 0 {
				assert.Equal(t, *test.p1, p1)
			}
			if intersects > 1 {
				assert.Equal(t, *test.p2, p2)
			}
		})
	}
}
