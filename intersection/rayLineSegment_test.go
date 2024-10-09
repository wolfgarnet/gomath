package intersection

import (
	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
	"testing"
)

func TestRayLineSegment(t *testing.T) {
	tests := map[string]struct {
		ray         gomath.Ray[float32]
		lineSegment gomath.LineSegment[float32]
		//intersection *gomath.Vector2[float32]
		rayRatio   float32
		lineRatio  float32
		intersects bool
	}{
		"ray intersecting line segment in the middle": {
			ray: gomath.Ray[float32]{
				Origin:    gomath.Vector2[float32]{0, 0},
				Direction: gomath.Vector2[float32]{1, 0},
			},
			lineSegment: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{1, -1},
				P2: gomath.Vector2[float32]{1, 1},
			},
			rayRatio:   1,
			lineRatio:  0.5,
			intersects: true,
		},
		"ray intersecting line segment in the middle 10x away": {
			ray: gomath.Ray[float32]{
				Origin:    gomath.Vector2[float32]{0, 0},
				Direction: gomath.Vector2[float32]{1, 0},
			},
			lineSegment: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{10, -1},
				P2: gomath.Vector2[float32]{10, 1},
			},
			rayRatio:   10,
			lineRatio:  0.5,
			intersects: true,
		},
		"ray intersecting line segment in the upper 5x away": {
			ray: gomath.Ray[float32]{
				Origin:    gomath.Vector2[float32]{0, 0},
				Direction: gomath.Vector2[float32]{1, 0},
			},
			lineSegment: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{5, -4},
				P2: gomath.Vector2[float32]{5, 1},
			},
			rayRatio:   5,
			lineRatio:  0.8,
			intersects: true,
		},
		"ray intersecting line segment in the upper 5x away reversed": {
			ray: gomath.Ray[float32]{
				Origin:    gomath.Vector2[float32]{0, 0},
				Direction: gomath.Vector2[float32]{-1, 0},
			},
			lineSegment: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{-5, -4},
				P2: gomath.Vector2[float32]{-5, 1},
			},
			rayRatio:   5,
			lineRatio:  0.8,
			intersects: true,
		},
		"ray intersecting line segment in the upper 5x away reversed 2": {
			ray: gomath.Ray[float32]{
				Origin:    gomath.Vector2[float32]{0, 0},
				Direction: gomath.Vector2[float32]{1, 0},
			},
			lineSegment: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{5, 1},
				P2: gomath.Vector2[float32]{5, -4},
			},
			rayRatio:   5,
			lineRatio:  0.2,
			intersects: true,
		},
		"ray not on line segment": {
			ray: gomath.Ray[float32]{
				Origin:    gomath.Vector2[float32]{0, 0},
				Direction: gomath.Vector2[float32]{1, 0},
			},
			lineSegment: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{-1, -1},
				P2: gomath.Vector2[float32]{-1, 1},
			},
			rayRatio:   10,
			lineRatio:  0.5,
			intersects: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rayRatio, lineRatio, intersects := RayLineSegment2(test.ray, test.lineSegment)
			assert.Equalf(t, test.intersects, intersects, "Failed test %v", name)
			if test.intersects {
				assert.Equalf(t, test.rayRatio, rayRatio, "Failed test %v", name)
				assert.Equalf(t, test.lineRatio, lineRatio, "Failed test %v", name)
			}
		})
	}
}
