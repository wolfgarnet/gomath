package gomath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector3_Length(t *testing.T) {
	tests := map[string]struct {
		vector   Vector3[float64]
		expected float64
	}{
		"zero":            {vector: Vector3[float64]{0, 0, 0}, expected: 0},
		"unit x":          {vector: Vector3[float64]{1, 0, 0}, expected: 1},
		"unit y":          {vector: Vector3[float64]{0, 1, 0}, expected: 1},
		"unit z":          {vector: Vector3[float64]{0, 0, 1}, expected: 1},
		"z only":          {vector: Vector3[float64]{0, 0, 3}, expected: 3},
		"3-4-0 triangle":  {vector: Vector3[float64]{3, 4, 0}, expected: 5},
		"1-2-2 pythagean": {vector: Vector3[float64]{1, 2, 2}, expected: 3},
		"2-3-6":           {vector: Vector3[float64]{2, 3, 6}, expected: 7},
		"negative":        {vector: Vector3[float64]{-2, -3, -6}, expected: 7},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, test.expected, test.vector.Length(), 0.0001)
			assert.InDelta(t, test.expected, test.vector.Length64(), 0.0001)
			assert.InDelta(t, test.expected, float64(test.vector.Length32()), 0.001)
			assert.InDelta(t, test.expected*test.expected, test.vector.LengthSq(), 0.0001)
		})
	}
}

func TestVector3_DistanceToSquared(t *testing.T) {
	tests := map[string]struct {
		v        Vector3[float64]
		v2       Vector3[float64]
		expected float64
	}{
		"same point":     {v: Vector3[float64]{1, 2, 3}, v2: Vector3[float64]{1, 2, 3}, expected: 0},
		"along x":        {v: Vector3[float64]{0, 0, 0}, v2: Vector3[float64]{10, 0, 0}, expected: 100},
		"along y":        {v: Vector3[float64]{0, 0, 0}, v2: Vector3[float64]{0, 10, 0}, expected: 100},
		"along z":        {v: Vector3[float64]{0, 0, 0}, v2: Vector3[float64]{0, 0, 10}, expected: 100},
		"3-4-0 triangle": {v: Vector3[float64]{0, 0, 0}, v2: Vector3[float64]{3, 4, 0}, expected: 25},
		"1-2-2":          {v: Vector3[float64]{0, 0, 0}, v2: Vector3[float64]{1, 2, 2}, expected: 9},
		"offset":         {v: Vector3[float64]{1, 1, 1}, v2: Vector3[float64]{2, 3, 3}, expected: 9},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, test.expected, test.v.DistanceToSquared(test.v2), 0.0001)
			assert.InDelta(t, math.Sqrt(test.expected), test.v.DistanceTo(test.v2), 0.0001)
			assert.InDelta(t, math.Sqrt(test.expected), test.v.Distance(test.v2), 0.0001)
		})
	}
}

func TestVector3_AngleTo(t *testing.T) {
	tests := map[string]struct {
		v        Vector3[float64]
		v2       Vector3[float64]
		expected float64
	}{
		"same direction unit":     {v: Vector3[float64]{1, 0, 0}, v2: Vector3[float64]{1, 0, 0}, expected: 0},
		"orthogonal unit":         {v: Vector3[float64]{1, 0, 0}, v2: Vector3[float64]{0, 1, 0}, expected: math.Pi / 2},
		"orthogonal non-unit":     {v: Vector3[float64]{2, 0, 0}, v2: Vector3[float64]{0, 5, 0}, expected: math.Pi / 2},
		"opposite non-unit":       {v: Vector3[float64]{3, 0, 0}, v2: Vector3[float64]{-7, 0, 0}, expected: math.Pi},
		"45 degrees non-unit":     {v: Vector3[float64]{2, 0, 0}, v2: Vector3[float64]{2, 2, 0}, expected: math.Pi / 4},
		"45 degrees in z":         {v: Vector3[float64]{0, 0, 3}, v2: Vector3[float64]{0, 3, 3}, expected: math.Pi / 4},
		"same direction non-unit": {v: Vector3[float64]{2, 2, 2}, v2: Vector3[float64]{5, 5, 5}, expected: 0},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, test.expected, test.v.AngleTo(test.v2), 0.0001)
		})
	}
}

func TestVector3_SetLength(t *testing.T) {
	tests := map[string]struct {
		vector   Vector3[float64]
		length   float64
		expected Vector3[float64]
	}{
		"scale up":    {vector: Vector3[float64]{3, 4, 0}, length: 10, expected: Vector3[float64]{6, 8, 0}},
		"scale down":  {vector: Vector3[float64]{0, 0, 8}, length: 2, expected: Vector3[float64]{0, 0, 2}},
		"unchanged":   {vector: Vector3[float64]{3, 4, 0}, length: 5, expected: Vector3[float64]{3, 4, 0}},
		"zero vector": {vector: Vector3[float64]{0, 0, 0}, length: 5, expected: Vector3[float64]{0, 0, 0}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.vector.SetLength(test.length)
			assert.InDelta(t, test.expected.X, got.X, 0.0001)
			assert.InDelta(t, test.expected.Y, got.Y, 0.0001)
			assert.InDelta(t, test.expected.Z, got.Z, 0.0001)
		})
	}
}

func TestVector3_Reflect(t *testing.T) {
	tests := map[string]struct {
		vector   Vector3[float64]
		normal   Vector3[float64]
		expected Vector3[float64]
	}{
		"off floor":        {vector: Vector3[float64]{1, -1, 0}, normal: Vector3[float64]{0, 1, 0}, expected: Vector3[float64]{1, 1, 0}},
		"off wall":         {vector: Vector3[float64]{-1, -1, 0}, normal: Vector3[float64]{1, 0, 0}, expected: Vector3[float64]{1, -1, 0}},
		"head on":          {vector: Vector3[float64]{0, -1, 0}, normal: Vector3[float64]{0, 1, 0}, expected: Vector3[float64]{0, 1, 0}},
		"parallel to wall": {vector: Vector3[float64]{1, 0, 0}, normal: Vector3[float64]{0, 1, 0}, expected: Vector3[float64]{1, 0, 0}},
		"off z plane":      {vector: Vector3[float64]{0, 0, -5}, normal: Vector3[float64]{0, 0, 1}, expected: Vector3[float64]{0, 0, 5}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.vector.Reflect(test.normal)
			assert.InDelta(t, test.expected.X, got.X, 0.0001)
			assert.InDelta(t, test.expected.Y, got.Y, 0.0001)
			assert.InDelta(t, test.expected.Z, got.Z, 0.0001)
		})
	}
}

// TestVector3_ReflectPreservesLength asserts the defining property of a reflection
// about a unit normal: the reflected vector has the same magnitude as the original.
func TestVector3_ReflectPreservesLength(t *testing.T) {
	vectors := []Vector3[float64]{
		{1, -1, 0}, {3, -4, 0}, {-2, -7, 1}, {0.5, -0.25, -0.75},
	}
	normals := []Vector3[float64]{
		{0, 1, 0}, {1, 0, 0}, {0, 0, 1},
	}

	for _, v := range vectors {
		for _, n := range normals {
			reflected := v.Reflect(n)
			assert.InDeltaf(t, v.Length(), reflected.Length(), 0.0001,
				"reflecting %v about %v gave %v", v, n, reflected)
		}
	}
}
