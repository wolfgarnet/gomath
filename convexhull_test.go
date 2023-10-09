package gomath

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvexHull2D(t *testing.T) {
	tests := []struct {
		points     []Vector2[float32]
		convexHull []Vector2[float32]
	}{
		{[]Vector2[float32]{{1, 1}, {0, 0}, {1, 0}, {0, 1}}, []Vector2[float32]{{1, 0}, {1, 1}, {0, 1}, {0, 0}}},
		{[]Vector2[float32]{{0, 0}, {1, 0}, {1, 1}, {0, 1}}, []Vector2[float32]{{1, 0}, {1, 1}, {0, 1}, {0, 0}}},
		{[]Vector2[float32]{{0, 0}, {1, 1}, {0, 1}, {1, 0}}, []Vector2[float32]{{1, 0}, {1, 1}, {0, 1}, {0, 0}}},
		{[]Vector2[float32]{{0, 0}, {1, 0}, {1, 1}, {0.5, 0.5}, {0, 1}}, []Vector2[float32]{{1, 0}, {1, 1}, {0, 1}, {0, 0}}},
		{[]Vector2[float32]{{0, 0}, {1, 1}, {0, 2}, {0.5, 0.5}, {2, 5}, {-1, 1}}, []Vector2[float32]{{0, 0}, {1, 1}, {2, 5}, {-1, 1}}},
		//{[]Vector2[float32{{1.2469797, 0}, {0.62348986, 10}, {1.2469794, 0}, {-0.44504166, 0}, {-1.8019377, 0}, {-0.44504294, 0}, {2, 0}, {1, 10}, {-0.22252083, 10}, {-0.90096897, 10}, {0.6234897, 10}, {-0.90096885, 10}, {-1.8019379, 0}, {-0.22252147, 10}}, []Vector2[float32{{0, 0}, {1, 1}, {2, 5}, {-1, 1}}},
		//{[]Vector2[float32{{1, 1}, {0, 0}, {0, 0}, {0.1, 0}, {0.6, 0}, {0.3, 0}, {0, 1}}, []Vector2[float32{{1, 0}, {1, 1}, {0, 1}, {0, 0}}},
	}

	for i, test := range tests {
		ch := ConvexHull(test.points)
		require.Equalf(t, test.convexHull, ch, "Test %v failed", i)
	}
}

func TestExtremePoints(t *testing.T) {
	tests := map[string]struct {
		center  Vector2[float32]
		polygon []Vector2[float32]
		left    Vector2[float32]
		right   Vector2[float32]
	}{
		"1": {
			center: Vector2[float32]{X: 0, Y: 0},
			polygon: []Vector2[float32]{
				{X: 1, Y: 1}, {X: 1, Y: 2}, {X: -1, Y: 2}, {X: -1, Y: 1},
			},
			left:  Vector2[float32]{X: -1, Y: 1},
			right: Vector2[float32]{X: 1, Y: 1},
		},
		"2": {
			center: Vector2[float32]{X: -2, Y: 1.5},
			polygon: []Vector2[float32]{
				{X: 1, Y: 1}, {X: 1, Y: 2}, {X: -1, Y: 2}, {X: -1, Y: 1},
			},
			left:  Vector2[float32]{X: -1, Y: 2},
			right: Vector2[float32]{X: -1, Y: 1},
		},
		"parallel up": {
			center: Vector2[float32]{X: 0, Y: 0},
			polygon: []Vector2[float32]{
				{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 2}, {X: 0, Y: 2},
			},
			left:  Vector2[float32]{X: 0, Y: 1},
			right: Vector2[float32]{X: 1, Y: 1},
		},
		"parallel left 1": {
			center: Vector2[float32]{X: -1, Y: 1},
			polygon: []Vector2[float32]{
				{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 2}, {X: 0, Y: 2},
			},
			left:  Vector2[float32]{X: 0, Y: 2},
			right: Vector2[float32]{X: 0, Y: 1},
		},
		"parallel left 2": {
			center: Vector2[float32]{X: -1, Y: 2},
			polygon: []Vector2[float32]{
				{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 2}, {X: 0, Y: 2},
			},
			left:  Vector2[float32]{X: 1, Y: 2},
			right: Vector2[float32]{X: 0, Y: 1},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			left, right := ExtremePoints(test.center, test.polygon)
			assert.Equal(t, test.left, left, "Left")
			assert.Equal(t, test.right, right, "Right")
		})
	}
}
