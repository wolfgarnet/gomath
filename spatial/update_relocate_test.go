package spatial

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSimpleHashGrid_UpdateItem_Relocates pins that moving an item across a cell
// boundary leaves exactly one copy, in the new cell. The old implementation
// hashed only the new position and appended when it wasn't found there, leaving
// a stale ghost in the previous cell.
func TestSimpleHashGrid_UpdateItem_RelocatesGhost(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](8, 8, 10)

	oldPos := vec(5, 5)   // cell (0,0)
	newPos := vec(75, 75) // cell (7,7)
	grid.AddItem(1, oldPos)

	grid.UpdateItem(1, oldPos, newPos)

	assert.Equal(t, uint64(1), grid.NumberOfElements(), "no ghost copy may remain")

	oldHash := grid.HashPositionIndexVector(oldPos)
	newHash := grid.HashPositionIndexVector(newPos)
	assert.Equal(t, 0, grid.GetCell(oldHash).Size(), "old cell must be empty")
	assert.Equal(t, 1, grid.GetCell(newHash).Size(), "item must be in the new cell")

	// Idempotent: replaying the same move must not add a second copy.
	grid.UpdateItem(1, oldPos, newPos)
	assert.Equal(t, uint64(1), grid.NumberOfElements(), "replay must not duplicate")
}

// TestSimpleHashGrid_UpdateItem_SameCell covers the within-cell case: the item
// stays put and is not duplicated, and is inserted if it wasn't present yet.
func TestSimpleHashGrid_UpdateItem_SameCell(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](8, 8, 10)

	a := vec(5, 5) // cell (0,0)
	b := vec(8, 8) // same cell (0,0)
	grid.AddItem(1, a)

	grid.UpdateItem(1, a, b)
	assert.Equal(t, uint64(1), grid.NumberOfElements())

	// Missing item gets inserted at the destination cell.
	grid.UpdateItem(2, a, b)
	assert.Equal(t, uint64(2), grid.NumberOfElements())
	assert.Equal(t, 2, grid.GetCell(grid.HashPositionIndexVector(b)).Size())
}

// TestHashGrid_UpdateItem_Relocates mirrors the relocation guarantee on the
// PositionedItem-based HashGrid, where the new position comes from Position().
func TestHashGrid_UpdateItem_Relocates(t *testing.T) {
	grid := NewHashGrid[float32, *testItem, uint64](8, 8, 10)

	oldPos := vec(5, 5)
	item := &testItem{IDValue: 1, PositionValue: oldPos}
	grid.AddItem(item)

	item.PositionValue = vec(75, 75) // item moved
	grid.UpdateItem(item, oldPos)

	assert.Equal(t, uint64(1), grid.NumberOfElements(), "no ghost copy may remain")
	assert.Equal(t, 0, grid.GetCell(grid.HashPositionIndexVector(oldPos)).Size())
	assert.Equal(t, 1, grid.GetCell(grid.HashPositionIndexVector(item.PositionValue)).Size())
}
