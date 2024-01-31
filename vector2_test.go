// vector_test.go
package gomath

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestVector2_ScaleToLength(t *testing.T) {
	tests := map[string]struct {
		v      Vector2[float64]
		length float64
		want   Vector2[float64]
	}{
		"case 1: Scaling Non-Zero Vector": {
			v: Vector2[float64]{
				X: math.Sqrt(2),
				Y: math.Sqrt(2),
			},
			length: 5.0,
			want:   Vector2[float64]{X: 3.536, Y: 3.536}, // Expected lengths are rounded to 3 decimal places
		},
		"case 2: Scaling Zero Vector": {
			v: Vector2[float64]{
				X: 0,
				Y: 0,
			},
			length: 4.0,
			want:   Vector2[float64]{X: 0, Y: 0}, // Zero vector scaled to any length remains a zero-vector
		},
		"case 3: Scaling With Zero Length": {
			v: Vector2[float64]{
				X: 2,
				Y: 3,
			},
			length: 0,
			want:   Vector2[float64]{X: 2, Y: 3}, // A vector scaled to a length of 0 remains unchanged
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.v.ScaleToLength(tt.length)
			assert.Equal(t, tt.want.X, Round2(got.X, 3), "They should be equal")
			assert.Equal(t, tt.want.Y, Round2(got.Y, 3), "They should be equal")
		})
	}
}

func TestVector2_AngleTo(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		v        Vector2[float64]
		v2       Vector2[float64]
		expected float64
	}{
		"default vectors": {
			v:        Vector2[float64]{X: 0, Y: 0},
			v2:       Vector2[float64]{X: 0, Y: 0},
			expected: math.NaN(),
		},
		"same vectors": {
			v:        Vector2[float64]{X: 1, Y: 1},
			v2:       Vector2[float64]{X: 1, Y: 1},
			expected: 0,
		},
		"orthogonal vectors": {
			v:        Vector2[float64]{X: 1, Y: 0},
			v2:       Vector2[float64]{X: 0, Y: 1},
			expected: math.Pi / 2,
		},
		"orthogonal vectors 2": {
			v:        Vector2[float64]{X: 1, Y: 0},
			v2:       Vector2[float64]{X: 0, Y: -1},
			expected: math.Pi / 2,
		},
		"opposite vectors": {
			v:        Vector2[float64]{X: 1, Y: 0},
			v2:       Vector2[float64]{X: -1, Y: 0},
			expected: math.Pi,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			angle := tc.v.AngleTo(tc.v2)
			if math.IsNaN(tc.expected) {
				assert.True(t, math.IsNaN(angle))
			} else {
				assert.Equal(t, tc.expected, angle)
			}
		})
	}
}

func TestVector2_AngleToWithSign(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		v        Vector2[float64]
		v2       Vector2[float64]
		expected float64
	}{
		"default vectors": {
			v:        Vector2[float64]{X: 0, Y: 0},
			v2:       Vector2[float64]{X: 0, Y: 0},
			expected: math.NaN(),
		},
		"same vectors": {
			v:        Vector2[float64]{X: 1, Y: 1},
			v2:       Vector2[float64]{X: 1, Y: 1},
			expected: 0,
		},
		"orthogonal vectors": {
			v:        Vector2[float64]{X: 1, Y: 0},
			v2:       Vector2[float64]{X: 0, Y: 1},
			expected: math.Pi / 2,
		},
		"orthogonal vectors 2": {
			v:        Vector2[float64]{X: 1, Y: 0},
			v2:       Vector2[float64]{X: 0, Y: -1},
			expected: -math.Pi / 2,
		},
		"opposite vectors": {
			v:        Vector2[float64]{X: 1, Y: 0},
			v2:       Vector2[float64]{X: -1, Y: 0},
			expected: math.Pi,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			angle := tc.v.AngleToWithSign(tc.v2)
			if math.IsNaN(tc.expected) {
				assert.True(t, math.IsNaN(angle))
			} else {
				assert.Equal(t, tc.expected, angle)
			}
		})
	}
}
