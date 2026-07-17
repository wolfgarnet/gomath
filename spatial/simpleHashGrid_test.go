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
