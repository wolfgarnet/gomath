package intersection

import (
	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
	"testing"
)

func TestAABB2DPoint(t *testing.T) {
	tests := []struct {
		point    gomath.Vector2[float32]
		bb       gomath.SquaredAABB2D[float32]
		contains bool
	}{
		{gomath.Vector2[float32]{0, 0}, gomath.NewSquaredAABB2DXY[float32](0, 0, 1), true},
		{gomath.Vector2[float32]{1, 0}, gomath.NewSquaredAABB2DXY[float32](0, 0, 1), true},
		{gomath.Vector2[float32]{1, 0}, gomath.NewSquaredAABB2DXY[float32](2, 0, 1), false},
		{gomath.Vector2[float32]{1.1, 0}, gomath.NewSquaredAABB2DXY[float32](0, 0, 1), false},
	}

	for i, test := range tests {
		contains := SquaredAABBPoint(test.bb, test.point)
		if contains && !test.contains {
			t.Errorf("Failed test %v. Was contained, shouldn't have been", i)
		}

		if !contains && test.contains {
			t.Errorf("Failed test %v. Was not contained, should have been", i)
		}
	}
}

func TestAABB2DAABB2D(t *testing.T) {
	tests := map[string]struct {
		bb1, bb2      gomath.SquaredAABB2D[float32]
		isOverlapping bool
		penetration   float32
		normal        gomath.Vector2[float32]
	}{
		"1":    {gomath.NewSquaredAABB2DXY[float32](0, 0, 1), gomath.NewSquaredAABB2DXY[float32](0, 0, 1), true, 2, gomath.Vector2[float32]{0, -1}},
		"0.5":  {gomath.NewSquaredAABB2DXY[float32](0, 0, 1), gomath.NewSquaredAABB2DXY[float32](0.5, 0, 1), true, 1.5, gomath.Vector2[float32]{1, 0}},
		"-0.5": {gomath.NewSquaredAABB2DXY[float32](0, 0, 1), gomath.NewSquaredAABB2DXY[float32](-0.5, 0, 1), true, 1.5, gomath.Vector2[float32]{-1, 0}},
		"2":    {gomath.NewSquaredAABB2DXY[float32](0, 0, 1), gomath.NewSquaredAABB2DXY[float32](1, 0, 1), true, 1, gomath.Vector2[float32]{1, 0}},
		"3":    {gomath.NewSquaredAABB2DXY[float32](0, 0, 1), gomath.NewSquaredAABB2DXY[float32](2, 0, 1), true, 0, gomath.Vector2[float32]{1, 0}},
		"4":    {gomath.NewSquaredAABB2DXY[float32](0, 0, 1), gomath.NewSquaredAABB2DXY[float32](2.1, 0, 1), false, 0, gomath.Vector2[float32]{}},
		"5":    {gomath.NewSquaredAABB2DXY[float32](0, 0, 1), gomath.NewSquaredAABB2DXY[float32](0, 2, 1), true, 0, gomath.Vector2[float32]{0, 1}},
		"6":    {gomath.NewSquaredAABB2DXY[float32](0, 0, 1), gomath.NewSquaredAABB2DXY[float32](0, 2.01, 1), false, 0, gomath.Vector2[float32]{}},
		"7":    {gomath.NewSquaredAABB2DXY[float32](0, 0, 1), gomath.NewSquaredAABB2DXY[float32](0, 0, 2), true, 3, gomath.Vector2[float32]{0, -1}},
		"8":    {gomath.NewSquaredAABB2DXY[float32](0, 0, 2), gomath.NewSquaredAABB2DXY[float32](0, 0, 1), true, 3, gomath.Vector2[float32]{0, -1}},
		"9":    {gomath.NewSquaredAABB2DXY[float32](-10, 0, 1), gomath.NewSquaredAABB2DXY[float32](10, 0, 1), false, 1, gomath.Vector2[float32]{}},
		"10":   {gomath.NewSquaredAABB2DXY[float32](200, 0, 2), gomath.NewSquaredAABB2DXY[float32](201, 0, 2), true, 3, gomath.Vector2[float32]{1, 0}},
		"11":   {gomath.NewSquaredAABB2DXY[float32](0, 0, 2), gomath.NewSquaredAABB2DXY[float32](3, 0, 2), true, 1, gomath.Vector2[float32]{1, 0}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := SquaredAABBSquaredAABB(test.bb1, test.bb2)
			if !r && test.isOverlapping {
				t.Errorf("Failed test %v, should have been contained", name)
			}
			if r && !test.isOverlapping {
				t.Errorf("Failed test %v, should NOT have been contained", name)
			}

			penetration, normal, overlapping := AABB2DAABB2DWithPenetration(test.bb1, test.bb2)
			if !overlapping && test.isOverlapping {
				t.Errorf("Failed test %v, should have been overlapping", name)
			}
			if overlapping && !test.isOverlapping {
				t.Errorf("Failed test %v, should NOT have been overlapping", name)
			}

			assert.Equal(t, test.normal, normal)

			if overlapping {
				assert.Equal(t, test.penetration, penetration)
			}
		})
	}
}
