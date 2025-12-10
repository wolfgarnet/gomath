package intersection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
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

			penetration, normal, overlapping := SquaredAABB2DSquaredAABB2DWithPenetration[float32](test.bb1, test.bb2)
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

func TestPointSideOfBox(t *testing.T) {
	tests := map[string]struct {
		point     gomath.Vector2[float32]
		ls1       gomath.LineSegment[float32]
		ls2       gomath.LineSegment[float32]
		contained bool
	}{ // positive y is down...?
		"left": {
			point: gomath.Vector2[float32]{Y: 1},
			ls1: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 1, Y: 1},
				P2: gomath.Vector2[float32]{X: 1, Y: 3},
			},
			ls2: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 3, Y: 3},
				P2: gomath.Vector2[float32]{X: 3, Y: 1},
			},
			contained: true,
		},
		"left upper": {
			point: gomath.Vector2[float32]{Y: 4},
			ls1: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 1, Y: 1},
				P2: gomath.Vector2[float32]{X: 1, Y: 3},
			},
			ls2: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 3, Y: 3},
				P2: gomath.Vector2[float32]{X: 3, Y: 1},
			},
			contained: false,
		},
		"test": {
			point: gomath.Vector2[float32]{X: 64, Y: 64},
			ls1: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 0, Y: 0},
				P2: gomath.Vector2[float32]{X: 0, Y: 96},
			},
			ls2: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 0, Y: 96},
				P2: gomath.Vector2[float32]{X: 1024, Y: 96},
			},
			contained: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			contained := PointSideOfBox(test.point, test.ls1, test.ls2)
			assert.Equal(t, test.contained, contained)
		})
	}
}

func TestPointInsideBox(t *testing.T) {
	tests := map[string]struct {
		point       gomath.Vector2[float32]
		box         gomath.AABB[float32]
		penetration float32
		normal      gomath.Vector2[float32]
		contained   bool
	}{
		"simple": {
			point: gomath.Vector2[float32]{X: 1, Y: 1},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 1, Y: 1},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained:   true,
			penetration: 1,
			normal:      gomath.Vector2[float32]{X: 0, Y: 1},
		},
		"simple2": {
			point: gomath.Vector2[float32]{X: 1, Y: 0.5},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 1, Y: 1},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained:   true,
			penetration: 0.5,
			normal:      gomath.Vector2[float32]{X: 0, Y: -1},
		},
		"simple3": {
			point: gomath.Vector2[float32]{X: 0.5, Y: 0.5},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 1, Y: 1},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained:   true,
			penetration: 0.5,
			normal:      gomath.Vector2[float32]{X: 0, Y: -1},
		},
		"simple4": {
			point: gomath.Vector2[float32]{X: 0.2, Y: 0.5},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 1, Y: 1},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained:   true,
			penetration: 0.2,
			normal:      gomath.Vector2[float32]{X: -1, Y: 0},
		},
		"outside up": {
			point: gomath.Vector2[float32]{X: 1.0, Y: 0.0},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 2, Y: 2},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained: false,
		},
		"outside up right": {
			point: gomath.Vector2[float32]{X: 1.1, Y: 0.9},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 2, Y: 2},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained: false,
		},
		"outside down": {
			point: gomath.Vector2[float32]{X: 1.0, Y: 3.01},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 2, Y: 2},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained: false,
		},
		"outside down right": {
			point: gomath.Vector2[float32]{X: 3.0, Y: 3.01},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 2, Y: 2},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained: false,
		},
		"on down": {
			point: gomath.Vector2[float32]{X: 1.0, Y: 3.0},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 2, Y: 2},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained:   true,
			penetration: 0.0,
			normal:      gomath.Vector2[float32]{X: 0, Y: 1},
		},
		"on down right": {
			point: gomath.Vector2[float32]{X: 3.0, Y: 3.0},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 2, Y: 2},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained:   true,
			penetration: 0.0,
			normal:      gomath.Vector2[float32]{X: 0, Y: 1},
		},
		"on down right 2": {
			point: gomath.Vector2[float32]{X: 3.0, Y: 2.9},
			box: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 2, Y: 2},
				HalfWidth:  1,
				HalfHeight: 1,
			},
			contained:   true,
			penetration: 0.0,
			normal:      gomath.Vector2[float32]{X: 1, Y: 0},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			penetration, normal, contained := PointAABB2DPenetration(test.point, test.box)
			assert.Equal(t, test.contained, contained)
			if !contained {
				return
			}
			assert.InDelta(t, test.penetration, penetration, 0.00001)
			assert.Equal(t, test.normal, normal)
		})
	}
}
