package intersection

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
)

func TestLineAABB2(t *testing.T) {
	tests := map[string]struct {
		ls         gomath.LineSegment[float32]
		aabb       gomath.AABB[float32]
		intersects bool
	}{
		"actual": {
			ls: gomath.LineSegment[float32]{
				P1: gomath.Vector2[float32]{X: 161.87857, Y: 160.80266},
				P2: gomath.Vector2[float32]{X: 128, Y: 128},
			},
			aabb: gomath.AABB[float32]{
				Center:     gomath.Vector2[float32]{X: 176, Y: 176},
				HalfWidth:  16,
				HalfHeight: 16,
			},
			intersects: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.intersects, LineAABB(test.ls, test.aabb))
		})
	}
}

func TestLineSegmentLineSegment2D(t *testing.T) {
	tests := map[string]struct {
		p1, p2, q1, q2 gomath.Vector2[float32]
		intersecting   bool
		point          *gomath.Vector2[float32]
	}{
		"test 1":                         {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{1, 0}, gomath.Vector2[float32]{0.5, -0.5}, gomath.Vector2[float32]{0.5, 0.5}, true, &gomath.Vector2[float32]{0.5, 0}},
		"collinear 1 y axis":             {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 1}, gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 1}, true, nil},
		"collinear 2 y axis end points":  {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 1}, gomath.Vector2[float32]{0, 3}, true, nil},
		"test 4":                         {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{2, 0}, true, &gomath.Vector2[float32]{0, 0}},
		"collinear 3 y axis endpoints 2": {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 4}, true, nil},
		"collinear 4 x axis":             {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{2, 0}, gomath.Vector2[float32]{3, 0}, gomath.Vector2[float32]{5, 0}, false, nil},
		"test 6":                         {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{1, 0}, gomath.Vector2[float32]{1, 2}, false, nil},
		"test 7":                         {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 3}, gomath.Vector2[float32]{0, 5}, false, nil},
		"test 9":                         {gomath.Vector2[float32]{-2.2486298, -2.3608873}, gomath.Vector2[float32]{3.482002, 2.4513066}, gomath.Vector2[float32]{3.6403534, 0.059493814}, gomath.Vector2[float32]{-3.8130138, 0.6856461}, true, &gomath.Vector2[float32]{X: 0.90713596, Y: 0.2891097}},
		"test 10":                        {gomath.Vector2[float32]{-0.6, -2.4}, gomath.Vector2[float32]{0.46, 2.38}, gomath.Vector2[float32]{3.56, -0}, gomath.Vector2[float32]{-1.43, 0}, true, &gomath.Vector2[float32]{-0.0677824, 0}},
		"not touching":                   {gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{1, 0}, gomath.Vector2[float32]{1.0, 1.0}, gomath.Vector2[float32]{3.0, -1.0}, false, &gomath.Vector2[float32]{2.0, 0}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			point, intersecting := LineSegmentLineSegment2D(test.p1, test.p2, test.q1, test.q2)

			assert.Equalf(t, test.point, point, "Failed test %v", name)
			if intersecting != test.intersecting {
				if intersecting {
					t.Errorf("Test %v failed. Should NOT be intersecting(actual=%v/expected=%v)", name, intersecting, test.intersecting)
				} else {
					t.Errorf("Test %v failed. Should be intersecting(actual=%v/expected=%v)", name, intersecting, test.intersecting)
				}
			}
		})
	}
}

func TestLineSegmentLineSegment2D2(t *testing.T) {
	tests := []struct {
		p1, p2, q1, q2 gomath.Vector2[float32]
		expected       bool
	}{
		{gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{1, 0}, gomath.Vector2[float32]{0.5, -0.5}, gomath.Vector2[float32]{0.5, 0.5}, true},
		{gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 1}, gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 1}, false},
		{gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 1}, gomath.Vector2[float32]{0, 3}, false},
		{gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{2, 0}, true},
		{gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 4}, false},
		{gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{1, 0}, gomath.Vector2[float32]{1, 2}, false},
		{gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 3}, gomath.Vector2[float32]{0, 5}, false},
		{gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{1, 0}, gomath.Vector2[float32]{1, 2}, false},
		{gomath.Vector2[float32]{-2.2486298, -2.3608873}, gomath.Vector2[float32]{3.482002, 2.4513066}, gomath.Vector2[float32]{3.6403534, 0.059493814}, gomath.Vector2[float32]{-3.8130138, 0.6856461}, true},
		{gomath.Vector2[float32]{-0.6, -2.4}, gomath.Vector2[float32]{0.46, 2.38}, gomath.Vector2[float32]{3.56, -0}, gomath.Vector2[float32]{-1.43, 0}, true},
		// The following is collinear.
		//{gomath.Vector2[float32]{0, 0}, gomath.Vector2[float32]{0, 2}, gomath.Vector2[float32]{0, 1}, gomath.Vector2[float32]{0, 3}, true},
	}

	for i, test := range tests {
		point, intersecting := LineSegmentLineSegment2D2(test.p1, test.p2, test.q1, test.q2)
		fmt.Printf("INT PNT: %v \n", point)

		if intersecting != test.expected {
			if intersecting {
				t.Errorf("Test %v failed. Should NOT be intersecting(actual=%v/expected=%v)", i, intersecting, test.expected)
			} else {
				t.Errorf("Test %v failed. Should be intersecting(actual=%v/expected=%v)", i, intersecting, test.expected)
			}
		}
	}
}

func TestLineAABB(t *testing.T) {
	tests := map[string]struct {
		ls         gomath.LineSegment[float32]
		aabb       gomath.SquaredAABB2D[float32]
		intersects bool
	}{
		"straight line": {
			ls:         gomath.NewLineSegmentXY[float32](100, 0, -100, 0.0),
			aabb:       gomath.NewSquaredAABB2DXY[float32](0, 0, 10),
			intersects: true,
		},
		"straight line 2": {
			ls:         gomath.NewLineSegmentXY[float32](89, 0.0, -111, -0.0),
			aabb:       gomath.NewSquaredAABB2DXY[float32](64, 0, 10),
			intersects: true,
		},
		"straight line 3": {
			ls:         gomath.NewLineSegmentXY[float32](89, 0.00066808256, -111, -0.00084443326),
			aabb:       gomath.NewSquaredAABB2DXY[float32](64, 0, 10),
			intersects: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.intersects, LineSquaredBox(test.ls, test.aabb))
		})
	}
}
