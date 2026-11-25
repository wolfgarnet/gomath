package spatial

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wolfgarnet/gomath"
)

func vec(x, y float32) gomath.Vector2[float32] {
	return gomath.Vector2[float32]{X: x, Y: y}
}

// TestSimpleHashGrid_NonSquareNegativeY pins the y hash on a non-square grid.
// The negative-Y branch masked with yHashMask but then subtracted from xHashMask,
// so on a grid whose width and height differ the wrapped row was wrong.
func TestSimpleHashGrid_NonSquareNegativeY(t *testing.T) {
	// 8 wide, 16 tall: xHashMask = 7, yHashMask = 15.
	grid := NewSimpleHashGrid[float32, uint64](8, 16, 10)

	// y = -15 -> floor(15/10) = 1 masked against height. The wrapped row must be
	// derived from the height (16), not the width (8).
	_, y := grid.HashPosition(0, -15)
	assert.Equal(t, uint64(15)-uint64(1), y, "row must wrap within the grid height")

	// A symmetric positive case is unaffected and stays within height.
	_, y2 := grid.HashPosition(0, 155) // floor(155/10)=15
	assert.Equal(t, uint64(15), y2)
}

// TestSimpleHashGrid_NeighbourHashesPairWrap asserts NeighbourHashesPair wraps
// its cell coordinates at the grid edge instead of underflowing. For the origin
// cell the "up" and "left" neighbours wrap to the far edge.
func TestSimpleHashGrid_NeighbourHashesPairWrap(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](8, 8, 10)

	pairs := grid.NeighbourHashesPair(0) // cell (0,0)

	// Every coordinate must be a valid in-range cell index (0..7), never a
	// wrapped-around huge uint64.
	for _, p := range pairs {
		assert.Lessf(t, p[0], grid.Width(), "x %d out of range", p[0])
		assert.Lessf(t, p[1], grid.Height(), "y %d out of range", p[1])
	}

	// The eight neighbours of (0,0) on an 8x8 torus, in the code's order:
	// up, up-right, right, down-right, down, down-left, left, up-left.
	expected := [][2]uint64{
		{0, 7}, {1, 7}, {1, 0}, {1, 1}, {0, 1}, {7, 1}, {7, 0}, {7, 7},
	}
	assert.Equal(t, expected, pairs)
}

// TestSimpleHashGrid_RemoveItem covers the load-bearing add/remove round trip.
func TestSimpleHashGrid_RemoveItem(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](8, 8, 10)

	grid.AddItem(1, vec(15, 15))
	grid.AddItem(2, vec(15, 15))
	assert.Equal(t, uint64(2), grid.NumberOfElements())

	assert.True(t, grid.RemoveItem(2, vec(15, 15)))
	assert.Equal(t, uint64(1), grid.NumberOfElements())
	assert.False(t, grid.RemoveItem(99, vec(15, 15)))
}

// TestSimpleHashGrid_IterateCells pins the contract the broad phase depends
// on: IterateCells yields exactly the cells that hold an item, each once, in
// ascending hash order — whatever order the items went in, and however they
// were moved or removed since the last Clear.
func TestSimpleHashGrid_IterateCells(t *testing.T) {
	type step func(grid *SimpleHashGrid[float32, uint64])

	tests := []struct {
		name  string
		steps []step
	}{
		{
			name: "empty grid yields nothing",
		},
		{
			name: "items added out of hash order come back sorted",
			steps: []step{
				func(grid *SimpleHashGrid[float32, uint64]) {
					grid.AddItem(1, vec(500, 900))
					grid.AddItem(2, vec(10, 10))
					grid.AddItem(3, vec(300, 40))
					grid.AddItem(4, vec(11, 12)) // same cell as 2
				},
			},
		},
		{
			name: "negative positions wrap and still sort",
			steps: []step{
				func(grid *SimpleHashGrid[float32, uint64]) {
					grid.AddItem(1, vec(-5, -5))
					grid.AddItem(2, vec(5, 5))
				},
			},
		},
		{
			name: "a cell emptied by removal is skipped",
			steps: []step{
				func(grid *SimpleHashGrid[float32, uint64]) {
					grid.AddItem(1, vec(10, 10))
					grid.AddItem(2, vec(300, 300))
					grid.RemoveItem(1, vec(10, 10))
				},
			},
		},
		{
			name: "a cell emptied by a move is skipped, the new cell is yielded",
			steps: []step{
				func(grid *SimpleHashGrid[float32, uint64]) {
					grid.AddItem(1, vec(10, 10))
					grid.UpdateItem(1, vec(10, 10), vec(700, 700))
				},
			},
		},
		{
			name: "items added by cell index are yielded",
			steps: []step{
				func(grid *SimpleHashGrid[float32, uint64]) {
					grid.AddItemAtCellIndex(1, grid.Size()-1)
					grid.AddItemAtCellIndex(2, 0)
				},
			},
		},
		{
			name: "clear forgets everything added before it",
			steps: []step{
				func(grid *SimpleHashGrid[float32, uint64]) {
					grid.AddItem(1, vec(10, 10))
					grid.AddItem(2, vec(300, 300))
					grid.Clear()
					grid.AddItem(3, vec(600, 20))
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			grid := NewSimpleHashGrid[float32, uint64](40, 40, 24)
			for _, apply := range test.steps {
				apply(grid)
			}

			// The reference answer: every cell that holds anything, by brute
			// force over the whole grid.
			var expected []uint64
			for hash := uint64(0); hash < grid.Size(); hash++ {
				if grid.GetCell(hash).Size() > 0 {
					expected = append(expected, hash)
				}
			}

			var visited []uint64
			for hash, cell := range grid.IterateCells() {
				assert.Same(t, grid.GetCell(hash), cell, "cell yielded under the wrong hash")
				visited = append(visited, hash)
			}

			assert.Equal(t, expected, visited)
		})
	}
}

// TestSimpleHashGrid_ClearEmptiesEveryCell checks Clear reaches every cell
// that was written, including ones reached through moves and cell-index adds,
// which are the paths that have to remember occupancy themselves.
func TestSimpleHashGrid_ClearEmptiesEveryCell(t *testing.T) {
	grid := NewSimpleHashGrid[float32, uint64](40, 40, 24)
	extent := float32(grid.Width()) * grid.CellSize()
	for index := uint64(0); index < 500; index++ {
		position := vec(float32(index*37%997)/997*extent, float32(index*61%991)/991*extent)
		grid.AddItem(index, position)
		if index%3 == 0 {
			grid.UpdateItem(index, position, vec(position.Y, position.X))
		}
	}
	grid.AddItemAtCellIndex(9999, grid.Size()-1)
	assert.Positive(t, grid.NumberOfElements())

	grid.Clear()

	assert.Zero(t, grid.NumberOfElements())
	for range grid.IterateCells() {
		assert.Fail(t, "a cleared grid yielded a cell")
	}
}
