package spatial

import (
	"fmt"

	"github.com/wolfgarnet/gomath"
)

type GridCell[I BoundedItem[ID], ID comparable] struct {
	neighbours []*GridCell[I, ID]
	items      []I
}
type HashGrid[T gomath.Number, I BoundedItem[ID], ID comparable] struct {
	grid            map[uint64]*GridCell[I, ID]
	size            uint64
	width           uint64
	height          uint64
	cellSize        T
	inverseCellSize T
	xHashMask       uint64
	yHashMask       uint64
}

func NewHashGrid[T gomath.Number, I BoundedItem[ID], ID comparable](width, height uint64, cellSize T) *HashGrid[T, I, ID] {
	grid := &HashGrid[T, I, ID]{
		width:           width,
		height:          height,
		size:            width * height,
		cellSize:        cellSize,
		inverseCellSize: 1 / cellSize,
		xHashMask:       width - 1,
		yHashMask:       height - 1,
	}

	grid.grid = make(map[uint64]*GridCell[I, ID])
	grid.initialize()

	return grid
}

func (grid *HashGrid[T, I, ID]) initialize() {
	for i := uint64(0); i < grid.size; i++ {
		grid.grid[i] = &GridCell[I, ID]{
			neighbours: make([]*GridCell[I, ID], 0, 8),
			items:      make([]I, 0, 8),
		}
	}

	idx := func(x, y uint64) uint64 {
		x = grid.width - 1 - x&grid.xHashMask
		y = grid.height - 1 - y&grid.yHashMask
		return x + y*grid.width
	}

	for i := uint64(0); i < grid.size; i++ {
		cell := grid.grid[i]
		x, y := grid.HashToCell(i)

		// Top
		cell.neighbours = append(cell.neighbours, grid.grid[idx(x, y-1)])
		// Top-right
		cell.neighbours = append(cell.neighbours, grid.grid[idx(x+1, y-1)])
		// Right
		cell.neighbours = append(cell.neighbours, grid.grid[idx(x+1, y)])
		// Bottom-right
		cell.neighbours = append(cell.neighbours, grid.grid[idx(x+1, y+1)])
		// Bottom
		cell.neighbours = append(cell.neighbours, grid.grid[idx(x, y+1)])
		// Bottom-left
		cell.neighbours = append(cell.neighbours, grid.grid[idx(x-1, y+1)])
		// Left
		cell.neighbours = append(cell.neighbours, grid.grid[idx(x-1, y)])
		// Top-left
		cell.neighbours = append(cell.neighbours, grid.grid[idx(x-1, y-1)])
	}
}

func (grid *HashGrid[T, I, ID]) NeighbourHashes(index uint64) (hashes []uint64) {
	idx := func(x, y uint64) uint64 {
		x = grid.width - 1 - x&grid.xHashMask
		y = grid.height - 1 - y&grid.yHashMask
		return x + y*grid.width
	}
	x, y := grid.HashToCell(index)
	hashes = make([]uint64, 0, 8)

	fmt.Printf("x: %d, y: %d\n", x, y)

	// Top
	hashes = append(hashes, idx(x, y-1))
	// Top-right
	hashes = append(hashes, idx(x+1, y-1))
	// Right
	hashes = append(hashes, idx(x+1, y))
	// Bottom-right
	hashes = append(hashes, idx(x+1, y+1))
	// Bottom
	hashes = append(hashes, idx(x, y+1))
	// Bottom-left
	hashes = append(hashes, idx(x-1, y+1))
	// Left
	hashes = append(hashes, idx(x-1, y))
	// Top-left
	hashes = append(hashes, idx(x-1, y-1))

	return hashes
}

func (grid *HashGrid[T, I, ID]) NeighbourHashesPair(index uint64) (pairs [][2]uint64) {
	x, y := grid.HashToCell(index)
	pairs = make([][2]uint64, 0, 8)

	pair := func(x, y uint64) [2]uint64 {
		fmt.Printf("CONVERTING x: %d, y: %d -> ", x, y)
		x = grid.width - 1 - x&grid.xHashMask
		ymasked := y & grid.yHashMask
		y = grid.height - 1 - ymasked
		fmt.Printf("x: %d, y: %d\n", x, y)
		return [2]uint64{x, y}
	}
	fmt.Printf("x: %d, y: %d\n", x, y)

	// Top
	pairs = append(pairs, pair(x, y-1))
	// Top-right
	pairs = append(pairs, pair(x+1, y-1))
	// Right
	pairs = append(pairs, pair(x+1, y))
	// Bottom-right
	pairs = append(pairs, pair(x+1, y+1))
	// Bottom
	pairs = append(pairs, pair(x, y+1))
	// Bottom-left
	pairs = append(pairs, pair(x-1, y+1))
	// Left
	pairs = append(pairs, pair(x-1, y))
	// Top-left
	pairs = append(pairs, pair(x-1, y-1))

	return pairs
}

func (grid *HashGrid[T, I, ID]) Neighbours(index uint64) []*GridCell[I, ID] {
	cell, _ := grid.grid[index]
	return cell.neighbours
}

func (grid *HashGrid[T, I, ID]) GetCell(index uint64) *GridCell[I, ID] {
	cell, _ := grid.grid[index]
	return cell
}

func (grid *HashGrid[T, I, ID]) HashPositionIndex(x T, y T) uint64 {
	var yHash, xHash uint64
	if x < 0 {
		xIndex := -x * grid.inverseCellSize
		xHash = grid.width - 1 - uint64(xIndex)&grid.xHashMask
	} else {
		xIndex := x * grid.inverseCellSize
		xHash = uint64(xIndex) & grid.xHashMask
	}

	if y < 0 {
		yIndex := -y * grid.inverseCellSize
		masked := uint64(yIndex) & grid.yHashMask
		yHash = grid.height - 1 - masked
		//yHash = grid.height - 1 - int(yIndex)&grid.yHashMask
	} else {
		yIndex := y * grid.inverseCellSize
		yHash = uint64(yIndex) & grid.yHashMask
	}

	hash := xHash + yHash*grid.width
	return hash
}

func (grid *HashGrid[T, I, ID]) HashPosition(x T, y T) (uint64, uint64) {
	var yHash, xHash uint64
	if x < 0 {
		xIndex := -x * grid.inverseCellSize
		xHash = grid.width - 1 - uint64(xIndex)&grid.xHashMask
	} else {
		xIndex := x * grid.inverseCellSize
		xHash = uint64(xIndex) & grid.xHashMask
	}
	if y < 0 {
		yIndex := -y * grid.inverseCellSize
		masked := uint64(yIndex) & grid.yHashMask
		yHash = grid.height - 1 - masked
	} else {
		yIndex := y * grid.inverseCellSize
		yHash = uint64(yIndex) & grid.yHashMask
	}

	return xHash, yHash
}

func (grid *HashGrid[T, I, ID]) HashToCell(hash uint64) (uint64, uint64) {
	return hash % grid.width, hash / grid.width
}
