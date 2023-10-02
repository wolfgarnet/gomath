package intersection

import (
	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
	"testing"
)

func TestRayLineSegment(t *testing.T) {
	tests := []struct {
		ray         gomath.Ray2D[float32]
		lineSegment gomath.LineSegment[float32]
		//intersection *gomath.Vector2[float32]
		rayRatio   float32
		lineRatio  float32
		intersects bool
	}{
		{
			ray: gomath.Ray2D[float32]{
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
		{
			ray: gomath.Ray2D[float32]{
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
		{
			ray: gomath.Ray2D[float32]{
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

	for i, test := range tests {
		rayRatio, lineRatio, intersects := RayLineSegment2(test.ray, test.lineSegment)
		assert.Equalf(t, test.intersects, intersects, "Failed test %v", i)
		if test.intersects {
			assert.Equalf(t, test.rayRatio, rayRatio, "Failed test %v", i)
			assert.Equalf(t, test.lineRatio, lineRatio, "Failed test %v", i)
		}
	}
}
