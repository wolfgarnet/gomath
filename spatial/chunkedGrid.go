package spatial

import (
	"iter"
	"math/bits"

	"github.com/wolfgarnet/gomath"
	"golang.org/x/exp/constraints"
)

// ChunkedGrid is a uniform grid of square cells for many moving items, rebuilt
// as often as they move, over a world of any size: a broad phase.
//
// Cells are grouped into square chunks, and a chunk exists only once an item has
// landed in it; chunks are found through a dense directory of chunk pointers
// that grows, in any direction and into negative coordinates, when an item lands
// outside it. Inside a chunk the cells are a plain array, so finding a cell, or
// a cell's neighbour, is arithmetic — no hashing — and an occupancy bitset per
// chunk lets a walk skip its empty cells. So the grid costs what its items
// occupy: the chunks they are in, never the world's area. That is the difference
// from SimpleHashGrid, which masks coordinates onto a fixed toroidal grid: a
// world wider than that grid folds onto itself, and items far apart share cells.
//
// Rebuilding is Clear and then AddItem for everything. Clear keeps every chunk
// and every cell's item list, so a grid rebuilt every tick allocates nothing
// once its items have visited the chunks they move through. (Chunks are kept
// once made; a world crossed by a crowd keeps the chunks it crossed.)
//
// Order is deterministic and needs no sorting: IterateCells walks occupied cells
// in ascending (y, x) order and a cell yields its items in the order they were
// added. That is also the order SimpleHashGrid walks a world that fits inside it
// starting at the origin, which lets one replace the other without changing what
// a simulation does there.
//
// It is not safe for concurrent modification. Use a floating-point T.
type ChunkedGrid[T constraints.Float, ID any] struct {
	cellSize        T
	inverseCellSize T

	// A chunk is 1<<chunkShift cells a side.
	chunkShift int
	chunkMask  int

	directory       []*gridChunk[ID]
	directoryWidth  int
	directoryHeight int
	originX         int // chunk coordinates of directory[0]
	originY         int

	// touched lists the chunks an item has landed in since the last Clear.
	touched []*gridChunk[ID]
	count   int
}

type gridChunk[ID any] struct {
	cells    [][]ID
	occupied []uint64
	touched  bool
}

// ChunkedGridCell is an occupied cell of a ChunkedGrid, as its walks yield it.
type ChunkedGridCell[T constraints.Float, ID any] struct {
	grid  *ChunkedGrid[T, ID]
	x, y  int
	items []ID
}

// NewChunkedGrid makes an empty grid of cellSize cells, chunks of 1<<chunkShift
// cells a side. A shift of 5 (32 × 32 cells) keeps a chunk's cell array small
// while the directory stays a pointer per chunk.
func NewChunkedGrid[T constraints.Float, ID any](cellSize T, chunkShift int) *ChunkedGrid[T, ID] {
	if cellSize <= 0 {
		panic("spatial: a chunked grid needs a positive cell size")
	}
	if chunkShift < 3 || chunkShift > 10 {
		panic("spatial: a chunked grid's chunk shift must be between 3 and 10")
	}
	return &ChunkedGrid[T, ID]{
		cellSize:        cellSize,
		inverseCellSize: 1 / cellSize,
		chunkShift:      chunkShift,
		chunkMask:       1<<chunkShift - 1,
	}
}

// CellSize is the side of a cell, in world units.
func (grid *ChunkedGrid[T, ID]) CellSize() T {
	return grid.cellSize
}

// CellCoordinate is the index of the cell a world coordinate falls in, on
// either axis: ⌊value / cellSize⌋, for negative values too.
func (grid *ChunkedGrid[T, ID]) CellCoordinate(value T) int {
	return floorToInt(value * grid.inverseCellSize)
}

// NumberOfElements is how many items the grid holds.
func (grid *ChunkedGrid[T, ID]) NumberOfElements() int {
	return grid.count
}

// Chunks is how many chunks the grid has made — the memory it holds.
func (grid *ChunkedGrid[T, ID]) Chunks() int {
	made := 0
	for _, chunk := range grid.directory {
		if chunk != nil {
			made++
		}
	}
	return made
}

// Clear empties the grid, keeping its chunks and their memory.
func (grid *ChunkedGrid[T, ID]) Clear() {
	for _, chunk := range grid.touched {
		for word, bitsSet := range chunk.occupied {
			for bitsSet != 0 {
				bit := bits.TrailingZeros64(bitsSet)
				bitsSet &= bitsSet - 1
				cell := &chunk.cells[word<<6|bit]
				*cell = (*cell)[:0]
			}
			chunk.occupied[word] = 0
		}
		chunk.touched = false
	}
	grid.touched = grid.touched[:0]
	grid.count = 0
}

// AddItem puts item in the cell containing position.
func (grid *ChunkedGrid[T, ID]) AddItem(item ID, position gomath.Vector2[T]) {
	x, y := grid.CellCoordinate(position.X), grid.CellCoordinate(position.Y)
	chunkX, chunkY := x>>grid.chunkShift, y>>grid.chunkShift
	chunk := grid.chunk(chunkX, chunkY)
	if chunk == nil {
		grid.ensureChunk(chunkX, chunkY)
		chunk = grid.chunk(chunkX, chunkY)
	}
	if !chunk.touched {
		chunk.touched = true
		grid.touched = append(grid.touched, chunk)
	}
	local := (y&grid.chunkMask)<<grid.chunkShift | x&grid.chunkMask
	chunk.cells[local] = append(chunk.cells[local], item)
	chunk.occupied[local>>6] |= 1 << (local & 63)
	grid.count++
}

// Cell is the cell at cell coordinates (x, y); an empty one holds no items.
func (grid *ChunkedGrid[T, ID]) Cell(x int, y int) ChunkedGridCell[T, ID] {
	return ChunkedGridCell[T, ID]{grid: grid, x: x, y: y, items: grid.items(x, y)}
}

// IterateCells yields every occupied cell in ascending (y, x) order.
func (grid *ChunkedGrid[T, ID]) IterateCells() iter.Seq[ChunkedGridCell[T, ID]] {
	return func(yield func(ChunkedGridCell[T, ID]) bool) {
		side := 1 << grid.chunkShift
		for directoryY := range grid.directoryHeight {
			for row := range side {
				for directoryX := range grid.directoryWidth {
					chunk := grid.directory[directoryY*grid.directoryWidth+directoryX]
					if chunk == nil || !chunk.touched {
						continue
					}
					// A chunk row is side cells, inside one occupancy word or
					// spanning several.
					start := row << grid.chunkShift
					for local := start; local < start+side; {
						word := chunk.occupied[local>>6] >> (local & 63)
						if span := start + side - local; span < 64 {
							word &= 1<<span - 1
						}
						if word == 0 {
							local = (local | 63) + 1
							continue
						}
						local += bits.TrailingZeros64(word)
						if local >= start+side {
							break
						}
						cell := ChunkedGridCell[T, ID]{
							grid:  grid,
							x:     (grid.originX+directoryX)<<grid.chunkShift | local&grid.chunkMask,
							y:     (grid.originY+directoryY)<<grid.chunkShift | row,
							items: chunk.cells[local],
						}
						if !yield(cell) {
							return
						}
						local++
					}
				}
			}
		}
	}
}

// SweepPairs calls visit for every pair of items in the same or adjacent cells,
// once each: walking occupied cells in ascending (y, x) order, each item is
// paired with the items after it in its cell, then with every item of the
// cell's right, down-left, down and down-right neighbours, in that order — the
// order a sweep built from IterateCells, IterateContent and IterateNextNeighbors
// meets them, and the order SimpleHashGrid's walk meets them on a world that
// fits it. visit gets pointers into the grid's item lists, good until the next
// change to the grid.
//
// It is that sweep written as plain loops, for the broad phase that runs it on
// every tick: no nested iterators, nothing copied, and a cell's four neighbours
// looked up once for the cell rather than once for each item in it.
func (grid *ChunkedGrid[T, ID]) SweepPairs(visit func(first *ID, second *ID)) {
	side := 1 << grid.chunkShift
	var neighbours [4][]ID
	for directoryY := range grid.directoryHeight {
		for row := range side {
			for directoryX := range grid.directoryWidth {
				chunk := grid.directory[directoryY*grid.directoryWidth+directoryX]
				if chunk == nil || !chunk.touched {
					continue
				}
				start := row << grid.chunkShift
				for local := start; local < start+side; {
					word := chunk.occupied[local>>6] >> (local & 63)
					if span := start + side - local; span < 64 {
						word &= 1<<span - 1
					}
					if word == 0 {
						local = (local | 63) + 1
						continue
					}
					local += bits.TrailingZeros64(word)
					if local >= start+side {
						break
					}

					items := chunk.cells[local]
					column := local & grid.chunkMask
					if column > 0 && column < side-1 && row < side-1 {
						// All four neighbours are in this chunk: index them.
						neighbours[0] = chunk.cells[local+1]
						neighbours[1] = chunk.cells[local+side-1]
						neighbours[2] = chunk.cells[local+side]
						neighbours[3] = chunk.cells[local+side+1]
					} else {
						x := (grid.originX+directoryX)<<grid.chunkShift | column
						y := (grid.originY+directoryY)<<grid.chunkShift | row
						for position, offset := range nextNeighbourOffsets {
							neighbours[position] = grid.items(x+offset[0], y+offset[1])
						}
					}
					for first := range items {
						for second := first + 1; second < len(items); second++ {
							visit(&items[first], &items[second])
						}
						for _, neighbour := range neighbours {
							for other := range neighbour {
								visit(&items[first], &neighbour[other])
							}
						}
					}
					local++
				}
			}
		}
	}
}

// CellRangeIterator yields the items of every cell in the inclusive cell range,
// row by row (y, then x), each cell's items in the order they were added.
func (grid *ChunkedGrid[T, ID]) CellRangeIterator(minX int, minY int, maxX int, maxY int) iter.Seq[ID] {
	return func(yield func(ID) bool) {
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				for _, item := range grid.items(x, y) {
					if !yield(item) {
						return
					}
				}
			}
		}
	}
}

// AABBIterator yields the items of every cell aabb touches: from the cell its
// minimum corner is in to the cell its maximum corner is in. Items are binned by
// a single position, so this answers "items whose position is in these cells";
// a caller after items whose extent overlaps aabb widens it by that extent.
func (grid *ChunkedGrid[T, ID]) AABBIterator(aabb gomath.AABB2[T]) iter.Seq[ID] {
	return grid.CellRangeIterator(
		grid.CellCoordinate(aabb.MinX), grid.CellCoordinate(aabb.MinY),
		grid.CellCoordinate(aabb.MaxX), grid.CellCoordinate(aabb.MaxY))
}

// Size is how many items the cell holds.
func (cell ChunkedGridCell[T, ID]) Size() int {
	return len(cell.items)
}

// Coordinates are the cell's cell coordinates.
func (cell ChunkedGridCell[T, ID]) Coordinates() (x int, y int) {
	return cell.x, cell.y
}

// IterateContent yields the cell's items with their positions in the cell.
func (cell ChunkedGridCell[T, ID]) IterateContent() iter.Seq2[int, ID] {
	return func(yield func(int, ID) bool) {
		for index, item := range cell.items {
			if !yield(index, item) {
				return
			}
		}
	}
}

// IterateContentFrom yields the cell's items from position from on.
func (cell ChunkedGridCell[T, ID]) IterateContentFrom(from int) iter.Seq[ID] {
	return func(yield func(ID) bool) {
		for index := from; index < len(cell.items); index++ {
			if !yield(cell.items[index]) {
				return
			}
		}
	}
}

// IterateNextNeighbors yields the occupied ones of the four neighbours a pair
// sweep visits from a cell — right, down-left, down, down-right, in that order —
// so that walking every cell in IterateCells order and pairing it with itself
// and these meets every pair of adjacent cells exactly once. The same four, in
// the same order, as SimpleGridCell's.
func (cell ChunkedGridCell[T, ID]) IterateNextNeighbors() iter.Seq[ChunkedGridCell[T, ID]] {
	return func(yield func(ChunkedGridCell[T, ID]) bool) {
		for _, offset := range nextNeighbourOffsets {
			x, y := cell.x+offset[0], cell.y+offset[1]
			items := cell.grid.items(x, y)
			if len(items) == 0 {
				continue
			}
			if !yield(ChunkedGridCell[T, ID]{grid: cell.grid, x: x, y: y, items: items}) {
				return
			}
		}
	}
}

// nextNeighbourOffsets are the four neighbours a pair sweep visits from a cell,
// in the order it visits them.
var nextNeighbourOffsets = [4][2]int{{1, 0}, {-1, 1}, {0, 1}, {1, 1}}

// items is the item list of the cell at (x, y), nil if it holds none.
func (grid *ChunkedGrid[T, ID]) items(x int, y int) []ID {
	chunk := grid.chunk(x>>grid.chunkShift, y>>grid.chunkShift)
	if chunk == nil || !chunk.touched {
		return nil
	}
	return chunk.cells[(y&grid.chunkMask)<<grid.chunkShift|x&grid.chunkMask]
}

// chunk is the chunk at chunk coordinates, nil outside the directory or where
// none was made. Arithmetic right shifts round negative cell coordinates down,
// so they land in the right chunk.
func (grid *ChunkedGrid[T, ID]) chunk(chunkX int, chunkY int) *gridChunk[ID] {
	directoryX, directoryY := chunkX-grid.originX, chunkY-grid.originY
	if directoryX < 0 || directoryY < 0 || directoryX >= grid.directoryWidth || directoryY >= grid.directoryHeight {
		return nil
	}
	return grid.directory[directoryY*grid.directoryWidth+directoryX]
}

// ensureChunk makes the chunk at chunk coordinates, growing the directory to
// cover it first if need be. Growth keeps every chunk already made.
func (grid *ChunkedGrid[T, ID]) ensureChunk(chunkX int, chunkY int) {
	if grid.directory == nil {
		grid.directory = make([]*gridChunk[ID], 1)
		grid.directoryWidth, grid.directoryHeight = 1, 1
		grid.originX, grid.originY = chunkX, chunkY
	}
	if grid.chunk(chunkX, chunkY) == nil && !grid.covers(chunkX, chunkY) {
		lowX, lowY := min(chunkX, grid.originX), min(chunkY, grid.originY)
		highX := max(chunkX, grid.originX+grid.directoryWidth-1)
		highY := max(chunkY, grid.originY+grid.directoryHeight-1)
		width, height := highX-lowX+1, highY-lowY+1
		directory := make([]*gridChunk[ID], width*height)
		for y := range grid.directoryHeight {
			for x := range grid.directoryWidth {
				directory[(grid.originY+y-lowY)*width+grid.originX+x-lowX] = grid.directory[y*grid.directoryWidth+x]
			}
		}
		grid.directory, grid.directoryWidth, grid.directoryHeight = directory, width, height
		grid.originX, grid.originY = lowX, lowY
	}
	cells := 1 << (2 * grid.chunkShift)
	grid.directory[(chunkY-grid.originY)*grid.directoryWidth+chunkX-grid.originX] = &gridChunk[ID]{
		cells:    make([][]ID, cells),
		occupied: make([]uint64, (cells+63)/64),
	}
}

func (grid *ChunkedGrid[T, ID]) covers(chunkX int, chunkY int) bool {
	directoryX, directoryY := chunkX-grid.originX, chunkY-grid.originY
	return directoryX >= 0 && directoryY >= 0 && directoryX < grid.directoryWidth && directoryY < grid.directoryHeight
}
