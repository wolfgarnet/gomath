package spatial

import (
	"fmt"
	"iter"

	"github.com/wolfgarnet/gomath"
)

type GridCell[T gomath.Number, I PositionedItem[T, ID], ID comparable] struct {
	neighbours []*GridCell[T, I, ID]
	items      []I
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

	offsetX T
	offsetY T

	// xHashMask and yHashMask are used to mask the x and y hash values to the width and height of the grid
	xHashMask uint64
	yHashMask uint64
}

func NewHashGridFromWorldSize[T gomath.Number, I PositionedItem[T, ID], ID comparable](worldWidth, worldHeight, cellSize T) *HashGrid[T, I, ID] {
	width := gomath.NextPowerOfTwo(uint64(worldWidth / cellSize))
	height := gomath.NextPowerOfTwo(uint64(worldHeight / cellSize))
	grid := NewHashGrid[T, I, ID](width, height, cellSize)
	grid.offsetX = worldWidth / 2
	grid.offsetY = worldHeight / 2

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

func (grid *HashGrid[T, I, ID]) initialize() {
	for i := uint64(0); i < grid.size; i++ {
		grid.grid[i] = &GridCell[T, I, ID]{
			neighbours: make([]*GridCell[T, I, ID], 0, 8),
			items:      make([]I, 0, 8),
		}
	}

	idx := func(x, y uint64) uint64 {
		x = grid.xHashMask - x&grid.xHashMask
		y = grid.yHashMask - y&grid.yHashMask
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

func (grid *HashGrid[T, I, ID]) IteratorForHash(hash uint64) iter.Seq[I] {
	return func(yield func(I) bool) {
		for _, item := range grid.grid[hash].items {
			if yield(item) {
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
	bottomX := position.X + radius
	bottomY := position.Y + radius

	startX, startY := grid.HashPosition(topX, topY)
	endX, endY := grid.HashPosition(bottomX, bottomY)

	offsetX = endX - startX + 1
	offsetY = endY - startY + 1

	startHash = startX + startY*grid.width

	return startHash, offsetX, offsetY
}

func (grid *HashGrid[T, I, ID]) HashValuesForRectangle(position gomath.Vector2[T], halfWidth T, halfHeight T) (startHash uint64, offsetX uint64, offsetY uint64) {
	topX := position.X - halfWidth
	topY := position.Y - halfHeight
	bottomX := position.X + halfWidth
	bottomY := position.Y + halfHeight

	startX, startY := grid.HashPosition(topX, topY)
	endX, endY := grid.HashPosition(bottomX, bottomY)

	offsetX = endX - startX + 1
	offsetY = endY - startY + 1

	startHash = startX + startY*grid.width

	return startHash, offsetX, offsetY
}

func (grid *HashGrid[T, I, ID]) SquaredIterator(position gomath.Vector2[T], radius T) iter.Seq[I] {
	hash, offsetX, offsetY := grid.HashValuesForSquared(position, radius)
	return func(yield func(I) bool) {
		for y := uint64(0); y < offsetY; y++ {
			hashOffset := hash + y*grid.width
			for x := uint64(0); x < offsetX; x++ {
				currentHash := hashOffset + x
				for _, item := range grid.grid[currentHash].items {
					if yield(item) {
						return
					}
				}
			}
		}
	}
}

func (grid *HashGrid[T, I, ID]) RectangleIterator(position gomath.Vector2[T], halfWidth T, halfHeight T) iter.Seq[I] {
	hash, offsetX, offsetY := grid.HashValuesForRectangle(position, halfWidth, halfHeight)
	return func(yield func(I) bool) {
		for y := uint64(0); y < offsetY; y++ {
			hashOffset := hash + y*grid.width
			for x := uint64(0); x < offsetX; x++ {
				currentHash := hashOffset + x
				for _, item := range grid.grid[currentHash].items {
					if yield(item) {
						return
					}
				}
			}
		}
	}
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
	idx := func(x, y uint64) uint64 {
		x = grid.xHashMask - x&grid.xHashMask
		y = grid.yHashMask - y&grid.yHashMask
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
		x = grid.xHashMask - x&grid.xHashMask
		y = grid.yHashMask - y&grid.yHashMask
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

func (grid *HashGrid[T, I, ID]) Neighbours(index uint64) []*GridCell[T, I, ID] {
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

	worldPositionX += grid.offsetX
	worldPositionY += grid.offsetY

	if worldPositionY < 0 {
		yIndex := -worldPositionY * grid.inverseCellSize
		masked := uint64(yIndex) & grid.yHashMask
		yHash = grid.xHashMask - masked
		//yHash = grid.height - 1 - int(yIndex)&grid.yHashMask
	} else {
		yIndex := worldPositionY * grid.inverseCellSize
		yHash = uint64(yIndex) & grid.yHashMask
	}

	hash := xHash + yHash*grid.width
	return hash
}

func (grid *HashGrid[T, I, ID]) HashPosition(worldPositionX T, worldPositionY T) (uint64, uint64) {
	worldPositionX += grid.offsetX
	worldPositionY += grid.offsetY

	var yHash, xHash uint64
	if worldPositionX < 0 {
		xIndex := -worldPositionX * grid.inverseCellSize
		xHash = grid.xHashMask - uint64(xIndex)&grid.xHashMask
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
