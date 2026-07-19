package spatial

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The raw-hash accessors take a uint64 that a caller might compute or hold onto.
// An out-of-range hash has no cell, so the map lookup returns nil; dereferencing
// it used to panic. These tests pin that such a hash yields nothing / a zero
// value instead of crashing.

func TestSimpleHashGrid_RawHash_OutOfRangeNoPanic(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](8, 8, 10)
	grid.AddItem(1, vec(5, 5))

	bad := grid.Size() // one past the last valid hash

	assert.NotPanics(t, func() {
		count := 0
		for range grid.IteratorForHash(bad) {
			count++
		}
		assert.Equal(t, 0, count)

		count = 0
		for range grid.Iterator(bad) {
			count++
		}
		assert.Equal(t, 0, count)

		grid.IteratorWithYield(bad, func(uint64) bool { return true })

		assert.Equal(t, [8]*SimpleGridCell[float32, uint64]{}, grid.Neighbours(bad))
		assert.Nil(t, grid.GetCell(bad))
	})
}

func TestHashGrid_RawHash_OutOfRangeNoPanic(t *testing.T) {
	grid := NewHashGrid[float32, *testItem, uint64](8, 8, 10)
	grid.AddItem(&testItem{IDValue: 1, PositionValue: vec(5, 5)})

	bad := grid.Size()

	assert.NotPanics(t, func() {
		count := 0
		for range grid.IteratorForHash(bad) {
			count++
		}
		assert.Equal(t, 0, count)

		count = 0
		for range grid.Iterator(bad) {
			count++
		}
		assert.Equal(t, 0, count)

		grid.IteratorWithYield(bad, func(*testItem) bool { return true })

		assert.Equal(t, [8]*GridCell[float32, *testItem, uint64]{}, grid.Neighbours(bad))
		assert.Nil(t, grid.GetCell(bad))
	})
}
