package collider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wolfgarnet/gomath"
)

func TestLineSegmentLineSegment2D(t *testing.T) {
	tests := map[string]struct {
		p1, p2, q1, q2 gomath.Vector2[float32]
		intersecting   bool
		collinear      bool
		point          *gomath.Vector2[float32]
	}{
		"test 1":                         {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{1, 0}, gomath.Vector2[float32]{0.5, -0.5}, gomath.Vector2[float32]{0.5, 0.5}, true, false, &gomath.Vector2[float32]{0.5, 0}},
		"collinear 1 y axis":             {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 1}, gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 1}, true, true, nil},
		"collinear 2 y axis end points":  {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 1}, gomath.Vector2[float32]{0, 3}, true, true, nil},
		"test 4":                         {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{2, 0}, true, false, &gomath.Vector2[float32]{0, 0}},
		"collinear 3 y axis endpoints 2": {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 4}, true, true, nil},
		"collinear 4 x axis":             {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{2, 0}, gomath.Vector2[float32]{3, 0}, gomath.Vector2[float32]{5, 0}, false, true, nil},
		"test 6":                         {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{1, 0}, gomath.Vector2[float32]{1, 2}, false, false, nil},
		"test 7":                         {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 3}, gomath.Vector2[float32]{0, 5}, false, true, nil},
		"test 9":                         {gomath.Vector2[float32]{-2.2486298, -2.3608873}, gomath.Vector2[float32]{3.482002, 2.4513066}, gomath.Vector2[float32]{3.6403534, 0.059493814}, gomath.Vector2[float32]{-3.8130138, 0.6856461}, true, false, &gomath.Vector2[float32]{X: 0.90713596, Y: 0.2891097}},
		"test 10":                        {gomath.Vector2[float32]{-0.6, -2.4}, gomath.Vector2[float32]{0.46, 2.38}, gomath.Vector2[float32]{3.56, -0}, gomath.Vector2[float32]{-1.43, 0}, true, false, &gomath.Vector2[float32]{-0.0677824, 0}},
		"not touching":                   {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{1, 0}, gomath.Vector2[float32]{1.0, 1.0}, gomath.Vector2[float32]{3.0, -1.0}, false, false, &gomath.Vector2[float32]{2.0, 0}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			point, collinear, intersecting := LineSegmentLineSegment2D(test.p1.X, test.p1.Y, test.p2.X, test.p2.Y, test.q1.X, test.q1.Y, test.q2.X, test.q2.Y)
			require.Equal(t, test.collinear, collinear, "Collinear")
			require.Equal(t, test.intersecting, intersecting, "Intersecting")
			if intersecting && test.point != nil {
				assert.Equalf(t, test.point.X, point.X, "x position on segments")
				assert.Equalf(t, test.point.Y, point.Y, "y position on segments")
			}
		})
	}
}
