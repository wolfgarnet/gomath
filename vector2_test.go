// vector_test.go
package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
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

		// Non-unit inputs: the angle must not depend on the magnitude of either
		// operand, only on the directions.
		"orthogonal vectors, non-unit": {
			v:        Vector2[float64]{X: 3, Y: 0},
			v2:       Vector2[float64]{X: 0, Y: 5},
			expected: math.Pi / 2,
		},
		"45 degrees, non-unit": {
			v:        Vector2[float64]{X: 2, Y: 0},
			v2:       Vector2[float64]{X: 2, Y: 2},
			expected: math.Pi / 4,
		},
		"45 degrees, both scaled": {
			v:        Vector2[float64]{X: 7, Y: 0},
			v2:       Vector2[float64]{X: 3, Y: 3},
			expected: math.Pi / 4,
		},
		"120 degrees, non-unit": {
			v:        Vector2[float64]{X: 2, Y: 0},
			v2:       Vector2[float64]{X: -1, Y: 1.7320508},
			expected: 2 * math.Pi / 3,
		},
		"opposite vectors, non-unit": {
			v:        Vector2[float64]{X: 3, Y: 0},
			v2:       Vector2[float64]{X: -7, Y: 0},
			expected: math.Pi,
		},
		"same vectors, non-unit": {
			v:        Vector2[float64]{X: 2, Y: 2},
			v2:       Vector2[float64]{X: 5, Y: 5},
			expected: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			angle := tc.v.AngleTo(tc.v2)
			if math.IsNaN(tc.expected) {
				assert.True(t, math.IsNaN(angle))
			} else {
				assert.InDelta(t, tc.expected, angle, 0.0001)
			}

			// AngleTo is unsigned, so it must be symmetric.
			reversed := tc.v2.AngleTo(tc.v)
			if math.IsNaN(tc.expected) {
				assert.True(t, math.IsNaN(reversed))
			} else {
				assert.InDelta(t, tc.expected, reversed, 0.0001)
			}

			// The pointer variant must agree with the value variant.
			v, v2 := tc.v, tc.v2
			pointerAngle := v.AngleTop(&v2)
			if math.IsNaN(tc.expected) {
				assert.True(t, math.IsNaN(pointerAngle))
			} else {
				assert.InDelta(t, tc.expected, pointerAngle, 0.0001)
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
		// AngleToWithSign is Atan2-based, and Atan2(0, 0) is defined as 0.
		// (The Acos-based implementations return NaN here instead.)
		"default vectors": {
			v:        Vector2[float64]{X: 0, Y: 0},
			v2:       Vector2[float64]{X: 0, Y: 0},
			expected: 0,
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

func BenchmarkVectorAngle(b *testing.B) {
	b.Run("Angle atan2", func(b *testing.B) {
		v1 := Vector2[float64]{X: 1, Y: 0}
		v2 := Vector2[float64]{X: 0, Y: 1}
		for i := 0; i < b.N; i++ {
			a := v1.AngleToWithSign(v2)
			_ = a
		}
	})

	b.Run("Angle acos", func(b *testing.B) {
		v1 := Vector2[float64]{X: 1, Y: 0}
		v2 := Vector2[float64]{X: 0, Y: 1}
		for i := 0; i < b.N; i++ {
			a := v1.AngleToWithSignAcos(v2)
			_ = a
		}
	})

}

func BenchmarkVector2Float32Operations(b *testing.B) {

	b.Run("by value", func(b *testing.B) {
		v1 := Vector2[float32]{X: 1, Y: 0}
		v2 := Vector2[float32]{X: 0, Y: 1}
		v3 := Vector2[float32]{X: 0, Y: -1}
		for i := 0; i < b.N; i++ {
			v4 := v1.Add(v2)
			v5 := v4.Mul(v3)
			_ = v5.Add(v4)
		}
	})

	b.Run("by reference", func(b *testing.B) {
		v1 := &Vector2p[float32]{X: 1, Y: 0}
		v2 := &Vector2p[float32]{X: 0, Y: 1}
		v3 := &Vector2p[float32]{X: 0, Y: -1}
		for i := 0; i < b.N; i++ {
			v4 := v1.Add2(v2)
			v5 := v4.Mul2(v3)
			_ = v5.Add2(v4)
		}
	})
}
