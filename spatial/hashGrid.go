package spatial

import (
	"fmt"
	"iter"

	"github.com/wolfgarnet/gomath"
)

const (
	upNeighbour        = 0
	upRightNeighbour   = 1
	rightNeighbour     = 2
	downRightNeighbour = 3
	downNeighbour      = 4
	downLeftNeighbour  = 5
	leftNeighbour      = 6
	upLeftNeighbour    = 7
)

type GridCell[T gomath.Number, I PositionedItem[T, ID], ID comparable] struct {
	neighbours      [8]*GridCell[T, I, ID]
	neighbourHashes [8]uint64
	hash            uint64
	items           []I
}

type HashGrid[T gomath.Number, I PositionedItem[T, ID], ID comparable] struct {
	grid map[uint64]*GridCell[T, I, ID]

	// size is the total number of cells in the grid
	size uint64

	// width is the width of the grid, number of cells in x direction
	width uint64

	// height is the height of the grid, number of cells in y direction
	height uint64

	// cellSize is the size of a cell in the grid
	cellSize T

	// inverseCellSize is the inverse of cellSize for easier computations
	inverseCellSize T

	// xHashMask and yHashMask are used to mask the x and y hash values to the width and height of the grid
	xHashMask uint64
	yHashMask uint64
}

func NewHashGridFromWorldSize[T gomath.Number, I PositionedItem[T, ID], ID comparable](worldWidth, worldHeight, cellSize T) *HashGrid[T, I, ID] {
	width := gomath.NextPowerOfTwo(uint64(worldWidth / cellSize))
	height := gomath.NextPowerOfTwo(uint64(worldHeight / cellSize))
	grid := NewHashGrid[T, I, ID](width, height, cellSize)

	return grid
}

func NewHashGrid[T gomath.Number, I PositionedItem[T, ID], ID comparable](width, height uint64, cellSize T) *HashGrid[T, I, ID] {
	width = gomath.NextPowerOfTwo(width)
	height = gomath.NextPowerOfTwo(height)

	grid := &HashGrid[T, I, ID]{
		width:           width,
		height:          height,
		size:            width * height,
		cellSize:        cellSize,
		inverseCellSize: 1 / cellSize,
		xHashMask:       width - 1,
		yHashMask:       height - 1,
	}

	grid.grid = make(map[uint64]*GridCell[T, I, ID])
	grid.initialize()

	return grid
}

func (grid *HashGrid[T, I, ID]) Size() uint64 {
	return grid.size
}

func (grid *HashGrid[T, I, ID]) Width() uint64 {
	return grid.width
}

func (grid *HashGrid[T, I, ID]) Height() uint64 {
	return grid.height
}

func (grid *HashGrid[T, I, ID]) CellSize() T {
	return grid.cellSize
}

func (grid *HashGrid[T, I, ID]) NumberOfElements() uint64 {
	var count uint64
	for _, cell := range grid.grid {
		count += uint64(len(cell.items))
	}
	return count
}

func (grid *HashGrid[T, I, ID]) Display() string {
	return fmt.Sprintf("HashGrid: %d x %d, %d cells, %d items", grid.width, grid.height, grid.size, grid.NumberOfElements())
}

//go:inline
func (grid *HashGrid[T, I, ID]) localToHash(x, y uint64) uint64 {
	x = x & grid.xHashMask
	y = y & grid.yHashMask

	return x + y*grid.width
}

func (grid *HashGrid[T, I, ID]) initialize() {
	for i := uint64(0); i < grid.size; i++ {
		grid.grid[i] = &GridCell[T, I, ID]{
			neighbours:      [8]*GridCell[T, I, ID]{},
			neighbourHashes: [8]uint64{},
			items:           make([]I, 0, 8),
			hash:            i,
		}
	}

	for i := uint64(0); i < grid.size; i++ {
		cell := grid.grid[i]
		x, y := grid.HashToCell(i)

		// Top 0
		cell.neighbours[upNeighbour] = grid.grid[grid.localToHash(x, y-1)]
		cell.neighbourHashes[upNeighbour] = grid.localToHash(x, y-1)
		// Top-right 1
		cell.neighbours[upRightNeighbour] = grid.grid[grid.localToHash(x+1, y-1)]
		cell.neighbourHashes[upRightNeighbour] = grid.localToHash(x+1, y-1)
		// Right 2
		cell.neighbours[rightNeighbour] = grid.grid[grid.localToHash(x+1, y)]
		cell.neighbourHashes[rightNeighbour] = grid.localToHash(x+1, y)
		// Bottom-right 3
		cell.neighbours[downRightNeighbour] = grid.grid[grid.localToHash(x+1, y+1)]
		cell.neighbourHashes[downRightNeighbour] = grid.localToHash(x+1, y+1)
		// Bottom 4
		cell.neighbours[downNeighbour] = grid.grid[grid.localToHash(x, y+1)]
		cell.neighbourHashes[downNeighbour] = grid.localToHash(x, y+1)
		// Bottom-left 5
		cell.neighbours[downLeftNeighbour] = grid.grid[grid.localToHash(x-1, y+1)]
		cell.neighbourHashes[downLeftNeighbour] = grid.localToHash(x-1, y+1)
		// Left 6
		cell.neighbours[leftNeighbour] = grid.grid[grid.localToHash(x-1, y)]
		cell.neighbourHashes[leftNeighbour] = grid.localToHash(x-1, y)
		// Top-left 7
		cell.neighbours[upLeftNeighbour] = grid.grid[grid.localToHash(x-1, y-1)]
		cell.neighbourHashes[upLeftNeighbour] = grid.localToHash(x-1, y-1)
	}
}

func (grid *HashGrid[T, I, ID]) IteratorForHash(hash uint64) iter.Seq[I] {
	return func(yield func(I) bool) {
		for _, item := range grid.grid[hash].items {
			if !yield(item) {
				return
			}
		}
	}
}

func (grid *HashGrid[T, I, ID]) Iterator(hash uint64) iter.Seq[I] {
	return func(yield func(I) bool) {
		for _, item := range grid.grid[hash].items {
			if !yield(item) {
				return
			}
		}
		for _, neighbour := range grid.grid[hash].neighbours {
			for _, item := range neighbour.items {
				if !yield(item) {
					return
				}
			}
		}
	}
}

func (grid *HashGrid[T, I, ID]) IteratorWithYield(hash uint64, yield func(I) bool) {
	for _, item := range grid.grid[hash].items {
		if !yield(item) {
			return
		}
	}
	for _, neighbour := range grid.grid[hash].neighbours {
		for _, item := range neighbour.items {
			if !yield(item) {
				return
			}
		}
	}
}

func (grid *HashGrid[T, I, ID]) HashValuesForSquared(position gomath.Vector2[T], radius T) (startHash uint64, offsetX uint64, offsetY uint64) {
	topX := position.X - radius
	topY := position.Y - radius

	startX, startY := grid.HashPosition(topX, topY)
	startHash = startX + startY*grid.width

	offsetX = (2 * uint64(radius*grid.inverseCellSize)) + 1
	offsetY = offsetX

	return startHash, offsetX, offsetY
}

func (grid *HashGrid[T, I, ID]) HashValuesForRectangle(position gomath.Vector2[T], halfWidth T, halfHeight T) (startHash uint64, offsetX uint64, offsetY uint64) {
	topX := position.X - halfWidth
	topY := position.Y - halfHeight

	startX, startY := grid.HashPosition(topX, topY)
	startHash = startX + startY*grid.width

	offsetX = (2 * uint64(halfWidth*grid.inverseCellSize)) + 1
	offsetY = (2 * uint64(halfHeight*grid.inverseCellSize)) + 1

	return startHash, offsetX, offsetY
}

//go:inline
func (grid *HashGrid[T, I, ID]) rectangleIterator(hash uint64, offsetX uint64, offsetY uint64) iter.Seq[I] {
	return func(yield func(I) bool) {
		cell := grid.grid[hash]
		for y := uint64(0); y < offsetY; y++ {
			xCell := cell
			for x := uint64(0); x < offsetX; x++ {
				for _, item := range xCell.items {
					if !yield(item) {
						return
					}
				}
				// Move the cell pointer one to the right
				xCell = xCell.neighbours[rightNeighbour]
			}

			// Move the cell pointer one down
			cell = cell.neighbours[downNeighbour]
		}
	}
}

func (grid *HashGrid[T, I, ID]) SquaredIterator(position gomath.Vector2[T], radius T) iter.Seq[I] {
	hash, offsetX, offsetY := grid.HashValuesForSquared(position, radius)
	return grid.rectangleIterator(hash, offsetX, offsetY)
}

func (grid *HashGrid[T, I, ID]) RectangleIterator(position gomath.Vector2[T], halfWidth T, halfHeight T) iter.Seq[I] {
	hash, offsetX, offsetY := grid.HashValuesForRectangle(position, halfWidth, halfHeight)
	return grid.rectangleIterator(hash, offsetX, offsetY)
}

func (grid *HashGrid[T, I, ID]) SquaredAABBIteratorWithYield(position gomath.Vector2[T], radius T, yield func(I) bool) {
	hash, offsetX, offsetY := grid.HashValuesForSquared(position, radius)
	for y := uint64(0); y < offsetY; y++ {
		hashOffset := hash + y*grid.width
		for x := uint64(0); x < offsetX; x++ {
			currentHash := hashOffset + x
			for _, item := range grid.grid[currentHash].items {
				if !yield(item) {
					return
				}
			}
		}
	}
}

func (grid *HashGrid[T, I, ID]) IterateAABB(aabb gomath.AABB[T]) iter.Seq[I] {
	hash, offsetX, offsetY := grid.HashValuesForRectangle(aabb.Center, aabb.HalfWidth, aabb.HalfHeight)
	return func(yield func(I) bool) {
		for y := uint64(0); y < offsetY; y++ {
			hashOffset := hash + y*grid.width
			for x := uint64(0); x < offsetX; x++ {
				currentHash := hashOffset + x
				for _, item := range grid.grid[currentHash].items {
					if !yield(item) {
						return
					}
				}
			}
		}
	}
}

// AddItem adds an item to the grid and returns the hash of the cell it was added to
func (grid *HashGrid[T, I, ID]) AddItem(item I) (hash uint64) {
	p := item.Position()
	hash = grid.HashPositionIndex(p.X, p.Y)
	grid.grid[hash].items = append(grid.grid[hash].items, item)

	return hash
}

func (grid *HashGrid[T, I, ID]) AddItemAtCellIndex(item I, hash uint64) {
	grid.grid[hash].items = append(grid.grid[hash].items, item)
}

// RemoveItemAtCellIndex removes an item from the grid at the given cell index
func (grid *HashGrid[T, I, ID]) RemoveItemAtCellIndex(item I, hash uint64) bool {
	for i, pp := range grid.grid[hash].items {
		if pp.ID() == item.ID() {
			grid.grid[hash].items = RemoveUnordered(grid.grid[hash].items, i)
			return true
		}
	}

	return false
}

func (grid *HashGrid[T, I, ID]) RemoveItem(item I) bool {
	p := item.Position()
	hash := grid.HashPositionIndex(p.X, p.Y)
	for i, pp := range grid.grid[hash].items {
		if pp.ID() == item.ID() {
			grid.grid[hash].items = RemoveUnordered(grid.grid[hash].items, i)
			return true
		}
	}

	return false
}

func (grid *HashGrid[T, I, ID]) UpdateItem(item I) {
	p := item.Position()
	hash := grid.HashPositionIndex(p.X, p.Y)
	cell, _ := grid.grid[hash]
	for i, pp := range cell.items {
		if pp.ID() == item.ID() {
			cell.items[i] = item
			return
		}
	}
	cell.items = append(cell.items, item)
}

func (grid *HashGrid[T, I, ID]) NeighbourHashes(index uint64) (hashes []uint64) {
	x, y := grid.HashToCell(index)
	hashes = make([]uint64, 0, 8)

	// Top
	hashes = append(hashes, grid.localToHash(x, y-1))
	// Top-right
	hashes = append(hashes, grid.localToHash(x+1, y-1))
	// Right
	hashes = append(hashes, grid.localToHash(x+1, y))
	// Bottom-right
	hashes = append(hashes, grid.localToHash(x+1, y+1))
	// Bottom
	hashes = append(hashes, grid.localToHash(x, y+1))
	// Bottom-left
	hashes = append(hashes, grid.localToHash(x-1, y+1))
	// Left
	hashes = append(hashes, grid.localToHash(x-1, y))
	// Top-left
	hashes = append(hashes, grid.localToHash(x-1, y-1))

	return hashes
}

func (grid *HashGrid[T, I, ID]) NeighbourHashesPair(index uint64) (pairs [][2]uint64) {
	x, y := grid.HashToCell(index)
	pairs = make([][2]uint64, 0, 8)

	pair := func(x, y uint64) [2]uint64 {
		return [2]uint64{x & grid.xHashMask, y & grid.yHashMask}
	}

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

func (grid *HashGrid[T, I, ID]) Neighbours(index uint64) [8]*GridCell[T, I, ID] {
	cell, _ := grid.grid[index]
	return cell.neighbours
}

func (grid *HashGrid[T, I, ID]) GetCell(index uint64) *GridCell[T, I, ID] {
	cell, _ := grid.grid[index]
	return cell
}

func (grid *HashGrid[T, I, ID]) HashPositionIndexVector(position gomath.Vector2[T]) uint64 {
	return grid.HashPositionIndex(position.X, position.Y)
}

func (grid *HashGrid[T, I, ID]) HashPositionIndex(worldPositionX T, worldPositionY T) uint64 {
	var yHash, xHash uint64
	if worldPositionX < 0 {
		xIndex := -worldPositionX * grid.inverseCellSize
		xHash = grid.width - 1 - uint64(xIndex)&grid.xHashMask
	} else {
		xIndex := worldPositionX * grid.inverseCellSize
		xHash = uint64(xIndex) & grid.xHashMask
	}

	if worldPositionY < 0 {
		yIndex := -worldPositionY * grid.inverseCellSize
		masked := uint64(yIndex) & grid.yHashMask
		yHash = grid.yHashMask - masked
	} else {
		yIndex := worldPositionY * grid.inverseCellSize
		yHash = uint64(yIndex) & grid.yHashMask
	}

	hash := xHash + yHash*grid.width
	return hash
}

func (grid *HashGrid[T, I, ID]) HashPosition(worldPositionX T, worldPositionY T) (uint64, uint64) {
	var yHash, xHash uint64
	if worldPositionX < 0 {
		xIndex := -worldPositionX * grid.inverseCellSize
		xHash = grid.width - 1 - uint64(xIndex)&grid.xHashMask
	} else {
		xIndex := worldPositionX * grid.inverseCellSize
		xHash = uint64(xIndex) & grid.xHashMask
	}

	if worldPositionY < 0 {
		yIndex := -worldPositionY * grid.inverseCellSize
		masked := uint64(yIndex) & grid.yHashMask
		yHash = grid.yHashMask - masked
	} else {
		yIndex := worldPositionY * grid.inverseCellSize
		yHash = uint64(yIndex) & grid.yHashMask
	}

	return xHash, yHash
}

func (grid *HashGrid[T, I, ID]) HashToCell(hash uint64) (uint64, uint64) {
	return hash % grid.width, hash / grid.width
}

func (grid *HashGrid[T, I, ID]) HashToWorldCoordinate(hash uint64) gomath.Vector2[T] {
	x, y := grid.HashToCell(hash)
	return gomath.Vector2[T]{
		X: T(x) * grid.cellSize, // + grid.cellSize*0.5,
		Y: T(y) * grid.cellSize, // + grid.cellSize*0.5,
	}
}
