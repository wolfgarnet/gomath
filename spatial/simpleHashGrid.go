package spatial

import (
	"fmt"
	"iter"
	"slices"

	"github.com/wolfgarnet/gomath"
)

type SimpleGridCell[T gomath.Number, ID comparable] struct {
	neighbours      [8]*SimpleGridCell[T, ID]
	neighbourHashes [8]uint64
	hash            uint64
	items           []ID
}

func (cell *SimpleGridCell[T, ID]) Size() int {
	return len(cell.items)
}

func (cell *SimpleGridCell[T, ID]) Hash() uint64 {
	return cell.hash
}

func (cell *SimpleGridCell[T, ID]) IterateNextNeighbors() iter.Seq[*SimpleGridCell[T, ID]] {
	return func(yield func(*SimpleGridCell[T, ID]) bool) {
		for _, neighbour := range []int8{rightNeighbour, downLeftNeighbour, downNeighbour, downRightNeighbour} {
			if !yield(cell.neighbours[neighbour]) {
				return
			}
		}
	}
}

func (cell *SimpleGridCell[T, ID]) IterateContent() iter.Seq2[int, ID] {
	return func(yield func(int, ID) bool) {
		for i, item := range cell.items {
			if !yield(i, item) {
				return
			}
		}
	}
}

func (cell *SimpleGridCell[T, ID]) IterateContentFrom(id int) iter.Seq[ID] {
	return func(yield func(ID) bool) {
		for i := id; i < len(cell.items); i++ {
			item := cell.items[i]
			if !yield(item) {
				return
			}
		}
	}
}

// SimpleHashGrid is a fixed-size, toroidal spatial hash grid for broad-phase
// proximity queries: it cheaply answers "which items share or border this
// cell/rectangle" so a caller can run a precise narrow-phase test (distance,
// intersection, ...) on the much smaller candidate set.
//
// It is deliberately an over-approximation, and two properties follow from the
// toroidal wrap (cell neighbours are masked to width/height, so the right edge
// borders the left and the bottom borders the top):
//
//   - Edge wrap: a query near a world edge may return items from the opposite
//     edge. They are false positives that the narrow-phase test discards. The
//     wrap only ever adds candidate cells, so it never causes a missed pair.
//   - Duplicates: if a query spans as many cells as the grid is wide/tall, the
//     same cell — and therefore the same item — can be yielded more than once.
//
// Both are accepted on purpose: clamping or de-duplicating would add branches
// to the hot path for no correctness benefit, since results always feed a
// narrow-phase filter. Grid dimensions are rounded up to a power of two (see
// NewSimpleHashGrid), and the grid is not safe for concurrent modification.
//
// Use a floating-point T. Although Number admits integers, the grid precomputes
// inverseCellSize = 1/cellSize, which truncates to 0 for any integer cellSize
// greater than one and collapses every hash to cell 0.
type SimpleHashGrid[T gomath.Number, ID comparable] struct {
	grid map[uint64]*SimpleGridCell[T, ID]

	// order is IterateCells' reusable buffer of sorted cell hashes.
	order []uint64

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

// NewSimpleHashGrid builds a grid of width x height cells of the given cellSize.
// width and height are each rounded UP to the next power of two (so the hash can
// mask instead of divide), which means the resulting grid can be larger than
// requested — check Width/Height for the actual dimensions.
func NewSimpleHashGrid[T gomath.Number, ID comparable](width, height uint64, cellSize T) *SimpleHashGrid[T, ID] {
	width = gomath.NextPowerOfTwo(width)
	height = gomath.NextPowerOfTwo(height)

	grid := &SimpleHashGrid[T, ID]{
		width:           width,
		height:          height,
		size:            width * height,
		cellSize:        cellSize,
		inverseCellSize: 1 / cellSize,
		xHashMask:       width - 1,
		yHashMask:       height - 1,
	}

	grid.grid = make(map[uint64]*SimpleGridCell[T, ID])
	grid.initialize()

	return grid
}

func (grid *SimpleHashGrid[T, ID]) Clear() {
	for _, cell := range grid.grid {
		cell.items = cell.items[:0]
	}
}

func (grid *SimpleHashGrid[T, ID]) Size() uint64 {
	return grid.size
}

func (grid *SimpleHashGrid[T, ID]) Width() uint64 {
	return grid.width
}

func (grid *SimpleHashGrid[T, ID]) Height() uint64 {
	return grid.height
}

func (grid *SimpleHashGrid[T, ID]) CellSize() T {
	return grid.cellSize
}

func (grid *SimpleHashGrid[T, ID]) NumberOfElements() uint64 {
	var count uint64
	for _, cell := range grid.grid {
		count += uint64(len(cell.items))
	}
	return count
}

func (grid *SimpleHashGrid[T, ID]) Display() string {
	return fmt.Sprintf("HashGrid: %d x %d, %d cells, %d items", grid.width, grid.height, grid.size, grid.NumberOfElements())
}

//go:inline
func (grid *SimpleHashGrid[T, ID]) localToHash(x, y uint64) uint64 {
	x = x & grid.xHashMask
	y = y & grid.yHashMask

	return x + y*grid.width
}

func (grid *SimpleHashGrid[T, ID]) initialize() {
	for i := uint64(0); i < grid.size; i++ {
		grid.grid[i] = &SimpleGridCell[T, ID]{
			neighbours:      [8]*SimpleGridCell[T, ID]{},
			neighbourHashes: [8]uint64{},
			items:           make([]ID, 0, 8),
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

func (grid *SimpleHashGrid[T, ID]) IterateContent() iter.Seq[*SimpleGridCell[T, ID]] {
	return func(yield func(hashGrid *SimpleGridCell[T, ID]) bool) {
		for _, cell := range grid.grid {
			if !yield(cell) {
				return
			}
		}
	}
}

func (grid *SimpleHashGrid[T, ID]) IteratorForHash(hash uint64) iter.Seq[ID] {
	return func(yield func(ID) bool) {
		cell, ok := grid.grid[hash]
		if !ok {
			return
		}
		for _, item := range cell.items {
			if !yield(item) {
				return
			}
		}
	}
}

func (grid *SimpleHashGrid[T, ID]) Iterator(hash uint64) iter.Seq[ID] {
	return func(yield func(ID) bool) {
		cell, ok := grid.grid[hash]
		if !ok {
			return
		}
		for _, item := range cell.items {
			if !yield(item) {
				return
			}
		}
		for _, neighbour := range cell.neighbours {
			for _, item := range neighbour.items {
				if !yield(item) {
					return
				}
			}
		}
	}
}

func (grid *SimpleHashGrid[T, ID]) IteratorWithYield(hash uint64, yield func(ID) bool) {
	cell, ok := grid.grid[hash]
	if !ok {
		return
	}
	for _, item := range cell.items {
		if !yield(item) {
			return
		}
	}
	for _, neighbour := range cell.neighbours {
		for _, item := range neighbour.items {
			if !yield(item) {
				return
			}
		}
	}
}

func (grid *SimpleHashGrid[T, ID]) HashValuesForSquared(position gomath.Vector2[T], radius T) (startHash uint64, offsetX uint64, offsetY uint64) {
	topX := position.X - radius
	topY := position.Y - radius

	startX, startY := grid.HashPosition(topX, topY)
	startHash = startX + startY*grid.width

	offsetX = (2 * uint64(radius*grid.inverseCellSize)) + 1
	offsetY = offsetX

	return startHash, offsetX, offsetY
}

func (grid *SimpleHashGrid[T, ID]) HashValuesForRectangle(position gomath.Vector2[T], halfWidth T, halfHeight T) (startHash uint64, offsetX uint64, offsetY uint64) {
	topX := position.X - halfWidth
	topY := position.Y - halfHeight

	startX, startY := grid.HashPosition(topX, topY)
	startHash = startX + startY*grid.width

	offsetX = (2 * uint64(halfWidth*grid.inverseCellSize)) + 1
	offsetY = (2 * uint64(halfHeight*grid.inverseCellSize)) + 1

	return startHash, offsetX, offsetY
}

func (grid *SimpleHashGrid[T, ID]) HashValuesForRectangleMinMax(aabb gomath.AABB2[T]) (startHash uint64, offsetX uint64, offsetY uint64) {

	startX, startY := grid.HashPosition(aabb.MinX, aabb.MinY)
	startHash = startX + startY*grid.width

	//offsetX = (2 * uint64(halfWidth*grid.inverseCellSize)) + 1
	offsetX = uint64((aabb.MaxX-aabb.MinX)*grid.inverseCellSize) + 1
	//offsetY = (2 * uint64(halfHeight*grid.inverseCellSize)) + 1
	offsetY = uint64((aabb.MaxY-aabb.MinY)*grid.inverseCellSize) + 1

	return startHash, offsetX, offsetY
}

//go:inline
func (grid *SimpleHashGrid[T, ID]) rectangleIterator(hash uint64, offsetX uint64, offsetY uint64) iter.Seq[ID] {
	return func(yield func(ID) bool) {
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

// SquaredIterator yields the broad-phase candidates in the square of the given
// radius around position. Results may include edge-wrapped false positives and
// duplicates; see the SimpleHashGrid type doc.
func (grid *SimpleHashGrid[T, ID]) SquaredIterator(position gomath.Vector2[T], radius T) iter.Seq[ID] {
	hash, offsetX, offsetY := grid.HashValuesForSquared(position, radius)
	return grid.rectangleIterator(hash, offsetX, offsetY)
}

// RectangleIterator yields the broad-phase candidates in the half-width x
// half-height rectangle around position. Results may include edge-wrapped false
// positives and duplicates; see the SimpleHashGrid type doc.
func (grid *SimpleHashGrid[T, ID]) RectangleIterator(position gomath.Vector2[T], halfWidth T, halfHeight T) iter.Seq[ID] {
	hash, offsetX, offsetY := grid.HashValuesForRectangle(position, halfWidth, halfHeight)
	return grid.rectangleIterator(hash, offsetX, offsetY)
}

// AABBIterator yields the broad-phase candidates overlapping aabb. Results may
// include edge-wrapped false positives and duplicates; see the SimpleHashGrid
// type doc.
func (grid *SimpleHashGrid[T, ID]) AABBIterator(aabb gomath.AABB2[T]) iter.Seq[ID] {
	hash, offsetX, offsetY := grid.HashValuesForRectangleMinMax(aabb)
	return grid.rectangleIterator(hash, offsetX, offsetY)
}

// SquaredAABBIteratorWithYield is the callback form of SquaredIterator. Results
// may include edge-wrapped false positives and duplicates; see the
// SimpleHashGrid type doc.
func (grid *SimpleHashGrid[T, ID]) SquaredAABBIteratorWithYield(position gomath.Vector2[T], radius T, yield func(ID) bool) {
	hash, offsetX, offsetY := grid.HashValuesForSquared(position, radius)
	// Follow neighbour pointers (which wrap toroidally) rather than adding the
	// raw linear offset: hash + y*width + x escapes the row near the right edge
	// and can index past the last cell, where the map returns a nil cell and
	// .items panics. This mirrors rectangleIterator exactly.
	cell := grid.grid[hash]
	for y := uint64(0); y < offsetY; y++ {
		xCell := cell
		for x := uint64(0); x < offsetX; x++ {
			for _, item := range xCell.items {
				if !yield(item) {
					return
				}
			}
			xCell = xCell.neighbours[rightNeighbour]
		}
		cell = cell.neighbours[downNeighbour]
	}
}

// IterateCells yields every occupied cell, in ascending hash order.
//
// The order is sorted rather than the map's own because callers use this to
// generate work whose *result* depends on the order they see it in — a physics
// broad phase pairs items cell by cell, and the contacts it produces are then
// resolved in the order they were appended. Go randomises map iteration per run,
// so ranging the map directly makes a simulation that cannot replay itself:
// entities with two or more contacts settle a fraction of a unit differently
// every time, and over a few thousand ticks that compounds into a different
// world.
//
// Sorting costs O(cells log cells) per call against a buffer reused between
// calls, which is small beside the pair checks the caller then does. The buffer
// means one iteration at a time per grid: do not nest IterateCells over the same
// grid.
func (grid *SimpleHashGrid[T, ID]) IterateCells() iter.Seq2[uint64, *SimpleGridCell[T, ID]] {
	return func(yield func(hash uint64, cell *SimpleGridCell[T, ID]) bool) {
		grid.order = grid.order[:0]
		for hash := range grid.grid {
			grid.order = append(grid.order, hash)
		}
		slices.Sort(grid.order)

		for _, hash := range grid.order {
			if !yield(hash, grid.grid[hash]) {
				return
			}
		}
	}
}

// AddItem adds an item to the grid and returns the hash of the cell it was added to
func (grid *SimpleHashGrid[T, ID]) AddItem(item ID, position gomath.Vector2[T]) (hash uint64) {
	hash = grid.HashPositionIndex(position.X, position.Y)
	grid.grid[hash].items = append(grid.grid[hash].items, item)

	return hash
}

func (grid *SimpleHashGrid[T, ID]) AddItemAtCellIndex(item ID, hash uint64) {
	grid.grid[hash].items = append(grid.grid[hash].items, item)
}

// RemoveItemAtCellIndex removes an item from the grid at the given cell index
func (grid *SimpleHashGrid[T, ID]) RemoveItemAtCellIndex(item ID, hash uint64) bool {
	for i, pp := range grid.grid[hash].items {
		if pp == item {
			grid.grid[hash].items = RemoveUnordered(grid.grid[hash].items, i)
			return true
		}
	}

	return false
}

func (grid *SimpleHashGrid[T, ID]) RemoveItem(item ID, position gomath.Vector2[T]) bool {
	hash := grid.HashPositionIndex(position.X, position.Y)
	for i, pp := range grid.grid[hash].items {
		if pp == item {
			grid.grid[hash].items = RemoveUnordered(grid.grid[hash].items, i)
			return true
		}
	}

	return false
}

// UpdateItem moves item from the cell containing oldPosition to the cell
// containing newPosition. When the two positions map to the same cell the item
// is left in place (added if it was missing). Removing it from the old cell is
// what prevents a stale ghost copy when the item crosses a cell boundary. The
// operation is idempotent: the item ends up present exactly once, in the
// newPosition cell, regardless of how many times it is called with the same
// arguments.
func (grid *SimpleHashGrid[T, ID]) UpdateItem(item ID, oldPosition gomath.Vector2[T], newPosition gomath.Vector2[T]) {
	oldHash := grid.HashPositionIndex(oldPosition.X, oldPosition.Y)
	newHash := grid.HashPositionIndex(newPosition.X, newPosition.Y)

	if oldHash != newHash {
		oldCell := grid.grid[oldHash]
		for i, existing := range oldCell.items {
			if existing == item {
				oldCell.items = RemoveUnordered(oldCell.items, i)
				break
			}
		}
	}

	newCell := grid.grid[newHash]
	for _, existing := range newCell.items {
		if existing == item {
			return
		}
	}
	newCell.items = append(newCell.items, item)
}

func (grid *SimpleHashGrid[T, ID]) NeighbourHashes(index uint64) (hashes []uint64) {
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

func (grid *SimpleHashGrid[T, ID]) NeighbourHashesPair(index uint64) (pairs [][2]uint64) {
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

func (grid *SimpleHashGrid[T, ID]) Neighbours(index uint64) [8]*SimpleGridCell[T, ID] {
	cell, ok := grid.grid[index]
	if !ok {
		return [8]*SimpleGridCell[T, ID]{}
	}
	return cell.neighbours
}

// GetCell returns the cell for index, or nil if index is out of range.
func (grid *SimpleHashGrid[T, ID]) GetCell(index uint64) *SimpleGridCell[T, ID] {
	return grid.grid[index]
}

func (grid *SimpleHashGrid[T, ID]) HashPositionIndexVector(position gomath.Vector2[T]) uint64 {
	return grid.HashPositionIndex(position.X, position.Y)
}

func (grid *SimpleHashGrid[T, ID]) HashPositionIndex(worldPositionX T, worldPositionY T) uint64 {
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

func (grid *SimpleHashGrid[T, ID]) HashPosition(worldPositionX T, worldPositionY T) (uint64, uint64) {
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

func (grid *SimpleHashGrid[T, ID]) HashToCell(hash uint64) (uint64, uint64) {
	return hash % grid.width, hash / grid.width
}

func (grid *SimpleHashGrid[T, ID]) HashToWorldCoordinate(hash uint64) gomath.Vector2[T] {
	x, y := grid.HashToCell(hash)
	return gomath.Vector2[T]{
		X: T(x) * grid.cellSize, // + grid.cellSize*0.5,
		Y: T(y) * grid.cellSize, // + grid.cellSize*0.5,
	}
}
