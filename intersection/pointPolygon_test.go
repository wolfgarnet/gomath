package intersection

import (
	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
	"testing"
)

func TestPointPolygon2D(t *testing.T) {
	tests := []struct {
		points     []gomath.Vector2[float32]
		point      gomath.Vector2[float32]
		intersects bool
	}{
		{
			points:     []gomath.Vector2[float32]{{-1, 0}, {1, 0}, {1, 1}, {-1, 1}},
			point:      gomath.Vector2[float32]{0, -1},
			intersects: false,
		},
		{
			points:     []gomath.Vector2[float32]{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}},
			point:      gomath.Vector2[float32]{0, 0},
			intersects: true,
		},
		{
			points:     []gomath.Vector2[float32]{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}},
			point:      gomath.Vector2[float32]{-2, 0},
			intersects: false,
		},
		{
			points:     []gomath.Vector2[float32]{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}},
			point:      gomath.Vector2[float32]{2, 0},
			intersects: false,
		},
		{
			points:     []gomath.Vector2[float32]{{188, -67}, {187, 68}, {4, 0}},
			point:      gomath.Vector2[float32]{48, -16}, // 80,-16 80,16 48,16
			intersects: true,
		},
	}

	for i, test := range tests {
		polygon := gomath.Polygon[float32]{
			Points: test.points,
		}
		intersects := PointPolygon(test.point, polygon)
		assert.Equalf(t, test.intersects, intersects, "Failed test %v", i)
	}
}
