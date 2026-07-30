package gomath

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBox3MinMax(t *testing.T) {
	box := NewBox3(Vector3[float64]{X: 1, Y: 2, Z: 3}, Vector3[float64]{X: 1, Y: 2, Z: 3})

	assert.Equal(t, Vector3[float64]{X: 0, Y: 0, Z: 0}, box.Min())
	assert.Equal(t, Vector3[float64]{X: 2, Y: 4, Z: 6}, box.Max())
	assert.Equal(t, Vector3[float64]{X: 2, Y: 4, Z: 6}, box.Size())
}

func TestBox3FromMinMaxSortsItsArguments(t *testing.T) {
	// The corners need not be given in order.
	wanted := NewBox3FromMinMax(
		Vector3[float64]{X: -1, Y: -2, Z: -3},
		Vector3[float64]{X: 1, Y: 2, Z: 3})

	swapped := NewBox3FromMinMax(
		Vector3[float64]{X: 1, Y: 2, Z: 3},
		Vector3[float64]{X: -1, Y: -2, Z: -3})

	mixed := NewBox3FromMinMax(
		Vector3[float64]{X: 1, Y: -2, Z: 3},
		Vector3[float64]{X: -1, Y: 2, Z: -3})

	assert.Equal(t, wanted, swapped)
	assert.Equal(t, wanted, mixed)
	assert.Equal(t, Vector3[float64]{}, wanted.Center)
	assert.Equal(t, Vector3[float64]{X: 1, Y: 2, Z: 3}, wanted.HalfSize)
}

func TestBox3FromPoints(t *testing.T) {
	_, ok := NewBox3FromPoints([]Vector3[float64]{})
	assert.False(t, ok, "no points means no box")

	box, ok := NewBox3FromPoints([]Vector3[float64]{
		{X: 1, Y: 1, Z: 1},
		{X: -3, Y: 2, Z: 0},
		{X: 0, Y: -1, Z: 5},
	})
	require.True(t, ok)

	assert.Equal(t, Vector3[float64]{X: -3, Y: -1, Z: 0}, box.Min())
	assert.Equal(t, Vector3[float64]{X: 1, Y: 2, Z: 5}, box.Max())
}

func TestBox3RoundTripsThroughItsOwnCorners(t *testing.T) {
	// The box built from a box's corners must be the box itself, which pins the
	// corner enumeration and the min/max reconstruction against each other.
	box := NewBox3(Vector3[float64]{X: 2, Y: -1, Z: 4}, Vector3[float64]{X: 1, Y: 3, Z: 2})

	corners := box.Corners()
	rebuilt, ok := NewBox3FromPoints(corners[:])
	require.True(t, ok)

	assert.True(t, box.ApproxEquals(rebuilt, 1e-12), "got %v want %v", rebuilt, box)
}

func TestBox3CornerOrdering(t *testing.T) {
	// Bit 0 selects +X, bit 1 +Y, bit 2 +Z.
	box := NewBox3(Vector3[float64]{}, Vector3[float64]{X: 1, Y: 1, Z: 1})
	corners := box.Corners()

	assert.Equal(t, Vector3[float64]{X: -1, Y: -1, Z: -1}, corners[0])
	assert.Equal(t, Vector3[float64]{X: 1, Y: -1, Z: -1}, corners[1])
	assert.Equal(t, Vector3[float64]{X: -1, Y: 1, Z: -1}, corners[2])
	assert.Equal(t, Vector3[float64]{X: 1, Y: 1, Z: 1}, corners[7])
}

func TestBox3ContainsPoint(t *testing.T) {
	box := NewBox3(Vector3[float64]{}, Vector3[float64]{X: 1, Y: 2, Z: 3})

	assert.True(t, box.ContainsPoint(Vector3[float64]{}))
	assert.True(t, box.ContainsPoint(Vector3[float64]{X: 1, Y: 2, Z: 3}), "the boundary counts as inside")
	assert.True(t, box.ContainsPoint(Vector3[float64]{X: -1, Y: -2, Z: -3}))
	assert.False(t, box.ContainsPoint(Vector3[float64]{X: 1.1}))
	assert.False(t, box.ContainsPoint(Vector3[float64]{Z: 3.1}))
}

func TestBox3ContainsBox(t *testing.T) {
	outer := NewBox3(Vector3[float64]{}, Vector3[float64]{X: 3, Y: 3, Z: 3})

	assert.True(t, outer.ContainsBox(NewBox3(Vector3[float64]{}, Vector3[float64]{X: 1, Y: 1, Z: 1})))
	assert.True(t, outer.ContainsBox(outer), "a box contains itself")
	assert.True(t, outer.ContainsBox(NewBox3(Vector3[float64]{X: 2}, Vector3[float64]{X: 1, Y: 1, Z: 1})),
		"touching the inside of a face still counts as contained")
	assert.False(t, outer.ContainsBox(NewBox3(Vector3[float64]{X: 2.5}, Vector3[float64]{X: 1, Y: 1, Z: 1})))
	assert.False(t, outer.ContainsBox(NewBox3(Vector3[float64]{}, Vector3[float64]{X: 4, Y: 1, Z: 1})))
}

func TestBox3ClosestPoint(t *testing.T) {
	box := NewBox3(Vector3[float64]{}, Vector3[float64]{X: 1, Y: 1, Z: 1})

	// A point inside is its own closest point.
	inside := Vector3[float64]{X: 0.5, Y: -0.25}
	assert.Equal(t, inside, box.ClosestPoint(inside))

	// Outside on one axis clamps that axis only.
	assert.Equal(t, Vector3[float64]{X: 1, Y: 0.5}, box.ClosestPoint(Vector3[float64]{X: 5, Y: 0.5}))

	// Outside on every axis clamps to a corner.
	assert.Equal(t, Vector3[float64]{X: 1, Y: 1, Z: 1},
		box.ClosestPoint(Vector3[float64]{X: 9, Y: 9, Z: 9}))
}

func TestBox3UnionAndExpansion(t *testing.T) {
	first := NewBox3(Vector3[float64]{}, Vector3[float64]{X: 1, Y: 1, Z: 1})
	second := NewBox3(Vector3[float64]{X: 4}, Vector3[float64]{X: 1, Y: 1, Z: 1})

	union := first.Union(second)
	assert.Equal(t, Vector3[float64]{X: -1, Y: -1, Z: -1}, union.Min())
	assert.Equal(t, Vector3[float64]{X: 5, Y: 1, Z: 1}, union.Max())
	assert.True(t, union.ContainsBox(first))
	assert.True(t, union.ContainsBox(second))

	expanded := first.Expanded(0.5)
	assert.Equal(t, Vector3[float64]{X: 1.5, Y: 1.5, Z: 1.5}, expanded.HalfSize)
	assert.Equal(t, first.Center, expanded.Center, "expanding keeps the centre")

	toPoint := first.ExpandedToPoint(Vector3[float64]{X: 10})
	assert.True(t, toPoint.ContainsPoint(Vector3[float64]{X: 10}))
	assert.True(t, toPoint.ContainsBox(first))
}

func TestBox3VolumeAndSurfaceArea(t *testing.T) {
	box := NewBox3(Vector3[float64]{}, Vector3[float64]{X: 1, Y: 2, Z: 3})

	// Full extents 2 x 4 x 6.
	assert.Equal(t, 48.0, box.Volume())
	assert.Equal(t, 2*(2*4+4*6+6*2), int(box.SurfaceArea()))
}

func TestBox3IsEmpty(t *testing.T) {
	assert.True(t, Box3[float64]{}.IsEmpty(), "a zero box encloses nothing")
	assert.True(t, NewBox3(Vector3[float64]{}, Vector3[float64]{X: 1, Y: 0, Z: 1}).IsEmpty())
	assert.False(t, NewBox3(Vector3[float64]{}, Vector3[float64]{X: 1, Y: 1, Z: 1}).IsEmpty())
}

func TestBox3Transforms(t *testing.T) {
	box := NewBox3(Vector3[float64]{}, Vector3[float64]{X: 1, Y: 1, Z: 1})

	assert.Equal(t, Vector3[float64]{X: 5, Y: 6, Z: 7},
		box.Translate(Vector3[float64]{X: 5, Y: 6, Z: 7}).Center)
	assert.Equal(t, Vector3[float64]{X: 9}, box.WithCenter(Vector3[float64]{X: 9}).Center)
	assert.Equal(t, Vector3[float64]{X: 2, Y: 2, Z: 2}, box.Scaled(2).HalfSize)

	// Non-mutating, like every other primitive here.
	assert.Equal(t, Vector3[float64]{}, box.Center)
	assert.Equal(t, Vector3[float64]{X: 1, Y: 1, Z: 1}, box.HalfSize)
}

func TestBox3EqualsAndString(t *testing.T) {
	first := NewBox3(Vector3[float64]{X: 1}, Vector3[float64]{X: 2, Y: 2, Z: 2})
	same := NewBox3(Vector3[float64]{X: 1}, Vector3[float64]{X: 2, Y: 2, Z: 2})
	different := NewBox3(Vector3[float64]{X: 1.5}, Vector3[float64]{X: 2, Y: 2, Z: 2})

	assert.True(t, first.Equals(same))
	assert.False(t, first.Equals(different))
	assert.True(t, first.ApproxEquals(different, 1))
	assert.False(t, first.ApproxEquals(different, 0.1))

	assert.Contains(t, first.String(), "Box3")
}
