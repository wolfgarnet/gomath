package collider

import (
	"fmt"
	"github.com/wolfgarnet/gomath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolygonPolygon2D(t *testing.T) {
	tests := map[string]struct {
		p1, p2         []gomath.Vector2[float32]
		penetration    float32
		hasPenetration bool
	}{
		"1a": {
			p1:             []gomath.Vector2[float32]{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
			p2:             []gomath.Vector2[float32]{{1, 1}, {2, 1}, {2, 2}, {1, 2}},
			hasPenetration: true,
			penetration:    0.0,
		},
		"1b": {
			p1:             []gomath.Vector2[float32]{{0, 1}, {1, 1}, {1, 2}, {0, 2}},
			p2:             []gomath.Vector2[float32]{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
			hasPenetration: true,
			penetration:    0.0,
		},
		"2": {
			p1:             []gomath.Vector2[float32]{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
			p2:             []gomath.Vector2[float32]{{0.5, 0.5}, {2, 1}, {2, 2}, {1, 2}},
			hasPenetration: true,
			penetration:    0.5,
		},
		"3": {
			p1:             []gomath.Vector2[float32]{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
			p2:             []gomath.Vector2[float32]{{2, 2}, {3, 2}, {3, 3}, {2, 3}},
			hasPenetration: false,
			penetration:    0,
		},
		"4": {
			p1:             []gomath.Vector2[float32]{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
			p2:             []gomath.Vector2[float32]{{-2, -2}, {-3, -2}, {-3, -3}, {-2, -3}},
			hasPenetration: false,
			penetration:    0,
		},
		"5": {
			p1:             []gomath.Vector2[float32]{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
			p2:             []gomath.Vector2[float32]{{-2, 2}, {-3, 2}, {-3, 3}, {-2, 3}},
			hasPenetration: false,
			penetration:    0,
		},
		"6": {
			p1:             []gomath.Vector2[float32]{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			p2:             []gomath.Vector2[float32]{{1, 1}, {9, 1}, {9, 9}, {1, 9}},
			hasPenetration: true,
			penetration:    9,
		},
		"7": {
			p1:             []gomath.Vector2[float32]{{-2.5, 0.3}, {-2.5, -0.3}, {2.8, -0.3}, {2.8, 0.3}},
			p2:             []gomath.Vector2[float32]{{-6.0, -5.7}, {-5.7, -6.1}, {-1.1, -1.5}, {-1.5, -1.1}},
			hasPenetration: false,
			penetration:    0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p1 := gomath.Polygon[float32]{Points: test.p1}
			p2 := gomath.Polygon[float32]{Points: test.p2}
			penetration, surfaceNormal, hit := PolygonPolygon(p1, p2)
			fmt.Printf("Normal: %v\n", surfaceNormal)

			require.Equalf(t, test.hasPenetration, hit, "Test %v failed", name)
			assert.Equalf(t, test.penetration, penetration, "Test %v failed", name)
		})
	}
}
