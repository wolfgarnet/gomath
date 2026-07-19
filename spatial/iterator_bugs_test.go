package spatial

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
)

// --- C1: inverted yield in SimpleHashGrid.IterateContent / IteratorForHash ---

// TestSimpleHashGrid_IterateContent_VisitsAllCells pins the iter.Seq contract:
// a consumer that never breaks must see every cell. The bug returned after the
// first cell because it used `if yield(cell) { return }` (stop when the consumer
// wants MORE) instead of `if !yield(cell) { return }`.
func TestSimpleHashGrid_IterateContent_VisitsAllCells(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](8, 8, 10)
	grid.AddItem(1, vec(5, 5))
	grid.AddItem(2, vec(15, 5))
	grid.AddItem(3, vec(25, 25))

	cells := 0
	items := 0
	for cell := range grid.IterateContent() {
		cells++
		items += cell.Size()
	}

	assert.Equal(t, int(grid.Size()), cells, "must visit every cell")
	assert.Equal(t, 3, items, "must observe every inserted item")
}

// TestSimpleHashGrid_IterateContent_EarlyBreak asserts the loop still honours an
// early break after the fix (breaking on the first cell visits exactly one).
func TestSimpleHashGrid_IterateContent_EarlyBreak(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](8, 8, 10)

	cells := 0
	for range grid.IterateContent() {
		cells++
		break
	}

	assert.Equal(t, 1, cells, "break must stop after one cell")
}

// TestSimpleHashGrid_IteratorForHash_ReturnsAllItems pins that IteratorForHash
// yields every item in the cell, not just the first.
func TestSimpleHashGrid_IteratorForHash_ReturnsAllItems(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](8, 8, 10)
	hash := grid.AddItem(1, vec(15, 15))
	grid.AddItem(2, vec(15, 15))
	grid.AddItem(3, vec(15, 15))

	got := make([]uint64, 0, 3)
	for id := range grid.IteratorForHash(hash) {
		got = append(got, id)
	}

	assert.ElementsMatch(t, []uint64{1, 2, 3}, got)
}

// --- C2: unmasked linear indexing in the *WithYield / IterateAABB variants ---

// TestSimpleHashGrid_SquaredAABBIteratorWithYield_EdgeMatches pins two things at
// the grid edge, where hash+y*width+x runs past the linear end of the map:
//  1. it must not panic on a nil cell, and
//  2. it must return the same items as the pointer-walking SquaredIterator.
func TestSimpleHashGrid_SquaredAABBIteratorWithYield_EdgeMatches(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](8, 8, 10)
	// (75,75) hashes to the bottom-right cell (7,7) = linear index 63.
	grid.AddItem(1, vec(75, 75))

	position := vec(75, 75)
	var radius float32 = 15

	want := collectSquared(grid, position, radius)

	var got []uint64
	assert.NotPanics(t, func() {
		grid.SquaredAABBIteratorWithYield(position, radius, func(id uint64) bool {
			got = append(got, id)
			return true
		})
	})

	assert.Contains(t, got, uint64(1), "the item at the query centre must be found")
	assert.ElementsMatch(t, want, got, "WithYield must agree with SquaredIterator")
}

func collectSquared(grid *SimpleHashGrid[float32, uint64], position gomath.Vector2[float32], radius float32) []uint64 {
	var out []uint64
	for id := range grid.SquaredIterator(position, radius) {
		out = append(out, id)
	}
	return out
}

// TestHashGrid_IterateAABB_EdgeNoPanic pins the same edge/panic hazard on the
// HashGrid.IterateAABB path.
func TestHashGrid_IterateAABB_EdgeNoPanic(t *testing.T) {
	grid := NewHashGrid[float32, *testItem, uint64](8, 8, 10)
	item := &testItem{IDValue: 1, PositionValue: vec(75, 75)}
	grid.AddItem(item)

	aabb := gomath.AABB[float32]{Center: vec(75, 75), HalfWidth: 15, HalfHeight: 15}

	var found bool
	assert.NotPanics(t, func() {
		for it := range grid.IterateAABB(aabb) {
			if it.ID() == 1 {
				found = true
			}
		}
	})
	assert.True(t, found, "item at the query centre must be found")
}

// TestHashGrid_SquaredAABBIteratorWithYield_EdgeNoPanic mirrors the SimpleHashGrid
// edge test on the HashGrid variant.
func TestHashGrid_SquaredAABBIteratorWithYield_EdgeNoPanic(t *testing.T) {
	grid := NewHashGrid[float32, *testItem, uint64](8, 8, 10)
	item := &testItem{IDValue: 1, PositionValue: vec(75, 75)}
	grid.AddItem(item)

	var found bool
	assert.NotPanics(t, func() {
		grid.SquaredAABBIteratorWithYield(vec(75, 75), 15, func(it *testItem) bool {
			if it.ID() == 1 {
				found = true
			}
			return true
		})
	})
	assert.True(t, found, "item at the query centre must be found")
}
