package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVector2_Rotatep asserts that the in-place Rotatep agrees with the
// value-returning Rotate for the same input.
func TestVector2_Rotatep(t *testing.T) {
	tests := map[string]struct {
		vector Vector2[float64]
		angle  float64
	}{
		"right 90":  {vector: Vector2[float64]{1, 0}, angle: math.Pi / 2},
		"right -90": {vector: Vector2[float64]{1, 0}, angle: -math.Pi / 2},
		"up 90":     {vector: Vector2[float64]{0, 1}, angle: math.Pi / 2},
		"up 45":     {vector: Vector2[float64]{0, 1}, angle: math.Pi / 4},
		"up 180":    {vector: Vector2[float64]{0, 1}, angle: math.Pi},
		"diagonal":  {vector: Vector2[float64]{3, 4}, angle: math.Pi / 3},
		"zero":      {vector: Vector2[float64]{0, 0}, angle: math.Pi / 3},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			want := test.vector.Rotate(test.angle)

			got := test.vector
			got.Rotatep(test.angle)

			assert.InDelta(t, want.X, got.X, 0.0001)
			assert.InDelta(t, want.Y, got.Y, 0.0001)
		})
	}
}

// TestVector2_RotatepPreservesLength asserts a defining property of rotation:
// it never changes the magnitude of the vector.
func TestVector2_RotatepPreservesLength(t *testing.T) {
	angles := []float64{0, math.Pi / 6, math.Pi / 4, math.Pi / 2, math.Pi, -math.Pi / 3}

	for _, angle := range angles {
		v := Vector2[float64]{3, 4}
		v.Rotatep(angle)
		assert.InDeltaf(t, 5.0, v.Length(), 0.0001, "rotating (3,4) by %v gave %v", angle, v)
	}
}

// TestVector2_AngleToWithSignAcosMatchesAtan2 asserts the two signed-angle
// implementations agree. AngleToWithSign (Atan2-based) is the reference.
func TestVector2_AngleToWithSignAcosMatchesAtan2(t *testing.T) {
	tests := map[string]struct {
		v  Vector2[float64]
		v2 Vector2[float64]
	}{
		"orthogonal ccw":  {v: Vector2[float64]{1, 0}, v2: Vector2[float64]{0, 1}},
		"orthogonal cw":   {v: Vector2[float64]{1, 0}, v2: Vector2[float64]{0, -1}},
		"45 ccw non-unit": {v: Vector2[float64]{2, 0}, v2: Vector2[float64]{2, 2}},
		"45 cw non-unit":  {v: Vector2[float64]{2, 0}, v2: Vector2[float64]{2, -2}},
		"same":            {v: Vector2[float64]{1, 1}, v2: Vector2[float64]{1, 1}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, test.v.AngleToWithSign(test.v2), test.v.AngleToWithSignAcos(test.v2), 0.0001)
		})
	}
}

func TestVector2_SetLength(t *testing.T) {
	tests := map[string]struct {
		vector   Vector2[float64]
		length   float64
		expected Vector2[float64]
	}{
		"scale up":    {vector: Vector2[float64]{3, 4}, length: 10, expected: Vector2[float64]{6, 8}},
		"scale down":  {vector: Vector2[float64]{3, 4}, length: 2.5, expected: Vector2[float64]{1.5, 2}},
		"unchanged":   {vector: Vector2[float64]{3, 4}, length: 5, expected: Vector2[float64]{3, 4}},
		"to zero":     {vector: Vector2[float64]{3, 4}, length: 0, expected: Vector2[float64]{0, 0}},
		"axis":        {vector: Vector2[float64]{0, 8}, length: 2, expected: Vector2[float64]{0, 2}},
		"zero vector": {vector: Vector2[float64]{0, 0}, length: 5, expected: Vector2[float64]{0, 0}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.vector.SetLength(test.length)
			assert.InDelta(t, test.expected.X, got.X, 0.0001)
			assert.InDelta(t, test.expected.Y, got.Y, 0.0001)

			// The pointer variant must agree with the value variant.
			p := test.vector
			p.SetLengthp(test.length)
			assert.InDelta(t, test.expected.X, p.X, 0.0001)
			assert.InDelta(t, test.expected.Y, p.Y, 0.0001)
		})
	}
}

func TestVector2_Reflect(t *testing.T) {
	tests := map[string]struct {
		vector   Vector2[float64]
		normal   Vector2[float64]
		expected Vector2[float64]
	}{
		"off floor":        {vector: Vector2[float64]{1, -1}, normal: Vector2[float64]{0, 1}, expected: Vector2[float64]{1, 1}},
		"off ceiling":      {vector: Vector2[float64]{1, 1}, normal: Vector2[float64]{0, -1}, expected: Vector2[float64]{1, -1}},
		"off left wall":    {vector: Vector2[float64]{-1, -1}, normal: Vector2[float64]{1, 0}, expected: Vector2[float64]{1, -1}},
		"head on":          {vector: Vector2[float64]{0, -1}, normal: Vector2[float64]{0, 1}, expected: Vector2[float64]{0, 1}},
		"parallel to wall": {vector: Vector2[float64]{1, 0}, normal: Vector2[float64]{0, 1}, expected: Vector2[float64]{1, 0}},
		"shallow angle":    {vector: Vector2[float64]{3, -1}, normal: Vector2[float64]{0, 1}, expected: Vector2[float64]{3, 1}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.vector.Reflect(test.normal)
			assert.InDelta(t, test.expected.X, got.X, 0.0001)
			assert.InDelta(t, test.expected.Y, got.Y, 0.0001)

			// The pointer variant must agree with the value variant.
			v, n := test.vector, test.normal
			v.Reflectp(&n)
			assert.InDelta(t, test.expected.X, v.X, 0.0001)
			assert.InDelta(t, test.expected.Y, v.Y, 0.0001)
		})
	}
}

// TestVector2_ReflectPreservesLength asserts the defining property of a reflection
// about a unit normal: the reflected vector has the same magnitude as the original.
func TestVector2_ReflectPreservesLength(t *testing.T) {
	vectors := []Vector2[float64]{{1, -1}, {3, -4}, {-2, -7}, {0.5, -0.25}}
	normals := []Vector2[float64]{{0, 1}, {1, 0}, {0.70710678, 0.70710678}}

	for _, v := range vectors {
		for _, n := range normals {
			reflected := v.Reflect(n)
			assert.InDeltaf(t, v.Length(), reflected.Length(), 0.0001,
				"reflecting %v about %v gave %v", v, n, reflected)
		}
	}
}

// TestVector2_SetCenter pins that SetCenter actually mutates the receiver.
// Vector2 carries the volume methods to satisfy the CenteredVolume/SimpleVolume
// interfaces, where a no-op setter would silently drop writes.
func TestVector2_SetCenter(t *testing.T) {
	v := &Vector2[float64]{1, 1}
	v.SetCenter(Vector2[float64]{9, -3})

	assert.Equal(t, Vector2[float64]{9, -3}, *v)
	assert.Equal(t, Vector2[float64]{9, -3}, v.GetCenter())
}
