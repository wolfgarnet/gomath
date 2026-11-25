package spatial

import (
	"iter"

	"github.com/wolfgarnet/gomath"
	"golang.org/x/exp/constraints"
)

// ChunkedIndex answers "which items overlap this box" for items that rarely
// move — map regions, walls, buildings — over a world of any size, including
// one that grows while it is in use.
//
// Space is cut into square cells, and cells into square chunks. Each cell lists
// the items covering it; a query reads the cells its box covers and keeps the
// items whose box overlaps. So a query costs what it covers, an insert or a
// remove costs the cells that item covers, and nothing costs the world's size:
//
//   - Chunks are made only where an item lands, and found through a dense
//     directory of chunk pointers that grows, in any direction and into negative
//     coordinates, when an item lands outside it. The directory is one pointer a
//     chunk, so even a very large world keeps it small.
//   - There is no boundary to insert outside of, and no rebuild: removing an item
//     clears only the cells it was written to.
//
// It is the structure to reach for over a BoxedQuadtree when the world is large
// or changes. A boxed quadtree keeps an item that straddles a split at the higher
// node, so long items (wall runs) pile up near the root, where every query tests
// them; and its boundary is fixed when it is made.
//
// Its bad case is a huge item: an item is written into every cell it covers, so
// one spanning the whole world costs a cell entry everywhere.
//
// Results come in a deterministic order — the cells in row order (y, then x),
// within a cell in the order items were added to it, each item once, at its
// first cell — but not in any order a quadtree would give. Each item is yielded
// once per query, which is tracked by stamping items as they are seen. That
// makes a query a write: queries on one index must not be nested, and the index
// is not safe for concurrent use, even by readers.
//
// Use a floating-point T, as for the other grids.
type ChunkedIndex[T constraints.Float, ID comparable] struct {
	cellSize        T
	inverseCellSize T
	chunkCells      int

	// directory holds the chunks, row-major, directoryWidth chunks wide; entry
	// (x, y) is the chunk at chunk coordinate (originX+x, originY+y). Nil until
	// the first item lands.
	directory       []*indexChunk
	directoryWidth  int
	directoryHeight int
	originX         int
	originY         int

	// items is indexed by handle; a removed item's slot is reused through free.
	items   []indexedItem[T, ID]
	free    []int32
	handles map[ID]int32
	count   int

	// stamps marks, per handle, the query that last yielded it.
	stamps []uint32
	stamp  uint32
}

type indexedItem[T constraints.Float, ID comparable] struct {
	item BoxedItem[T, ID]
	live bool
	// The cell range the item was written to, inclusive.
	minX, minY, maxX, maxY int
}

// indexChunk is chunkCells × chunkCells cells, row-major, each a list of item
// handles.
type indexChunk struct {
	cells [][]int32
}

// NewChunkedIndex makes an index of cellSize cells, chunkCells cells to a chunk
// side. A cell the size of a map tile suits map regions; a chunk of 16 or 32
// cells keeps a chunk small while the directory stays a pointer per chunk.
func NewChunkedIndex[T constraints.Float, ID comparable](cellSize T, chunkCells int) *ChunkedIndex[T, ID] {
	if cellSize <= 0 {
		panic("spatial: a chunked index needs a positive cell size")
	}
	if chunkCells <= 0 {
		panic("spatial: a chunked index needs a positive chunk size")
	}
	return &ChunkedIndex[T, ID]{
		cellSize:        cellSize,
		inverseCellSize: 1 / cellSize,
		chunkCells:      chunkCells,
		handles:         make(map[ID]int32),
	}
}

// Size is how many items the index holds.
func (index *ChunkedIndex[T, ID]) Size() int {
	return index.count
}

// CellSize is the side of a cell, in world units.
func (index *ChunkedIndex[T, ID]) CellSize() T {
	return index.cellSize
}

// Clear removes every item and every chunk.
func (index *ChunkedIndex[T, ID]) Clear() {
	index.directory = nil
	index.directoryWidth = 0
	index.directoryHeight = 0
	index.originX = 0
	index.originY = 0
	index.items = index.items[:0]
	index.free = index.free[:0]
	clear(index.handles)
	index.count = 0
	index.stamps = index.stamps[:0]
	index.stamp = 0
}

// Insert adds item, or replaces the item already held under its ID. It never
// fails: the index grows to wherever the item is.
func (index *ChunkedIndex[T, ID]) Insert(item BoxedItem[T, ID]) {
	if _, exists := index.handles[item.ItemID]; exists {
		index.Remove(item.ItemID)
	}

	handle := index.newHandle()
	minX, minY, maxX, maxY := index.cellRange(item.AABB)
	index.items[handle] = indexedItem[T, ID]{item: item, live: true, minX: minX, minY: minY, maxX: maxX, maxY: maxY}
	index.handles[item.ItemID] = handle
	index.count++

	index.ensureChunks(minX, minY, maxX, maxY)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			cell := index.cell(x, y, true)
			*cell = append(*cell, handle)
		}
	}
}

// Remove takes the item with id out of the index, clearing only the cells it was
// written to, and reports whether it was there.
func (index *ChunkedIndex[T, ID]) Remove(id ID) bool {
	handle, exists := index.handles[id]
	if !exists {
		return false
	}
	entry := &index.items[handle]
	for y := entry.minY; y <= entry.maxY; y++ {
		for x := entry.minX; x <= entry.maxX; x++ {
			cell := index.cell(x, y, false)
			if cell == nil {
				continue
			}
			for position, held := range *cell {
				if held == handle {
					// Keep the rest in order, so the order a cell yields in
					// stays the order its items were added.
					*cell = append((*cell)[:position], (*cell)[position+1:]...)
					break
				}
			}
		}
	}
	*entry = indexedItem[T, ID]{}
	delete(index.handles, id)
	index.free = append(index.free, handle)
	index.count--
	return true
}

// Iterator yields every item whose box overlaps bounds — the same overlap test
// a BoxedQuadtree's Iterator applies, so the two answer the same set — each item
// once. See the type's doc for the order and for why queries must not nest.
func (index *ChunkedIndex[T, ID]) Iterator(bounds gomath.AABB[T]) iter.Seq[BoxedItem[T, ID]] {
	return func(yield func(BoxedItem[T, ID]) bool) {
		if index.count == 0 {
			return
		}
		stamp := index.nextStamp()
		minX, minY, maxX, maxY := index.cellRange(bounds)
		// Only the part of the query the directory covers can hold anything.
		minX = max(minX, index.originX*index.chunkCells)
		minY = max(minY, index.originY*index.chunkCells)
		maxX = min(maxX, (index.originX+index.directoryWidth)*index.chunkCells-1)
		maxY = min(maxY, (index.originY+index.directoryHeight)*index.chunkCells-1)

		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				cell := index.cell(x, y, false)
				if cell == nil {
					continue
				}
				for _, handle := range *cell {
					if index.stamps[handle] == stamp {
						continue
					}
					index.stamps[handle] = stamp
					item := index.items[handle].item
					if !AABBAABB3(item.AABB, bounds) {
						continue
					}
					if !yield(item) {
						return
					}
				}
			}
		}
	}
}

func (index *ChunkedIndex[T, ID]) newHandle() int32 {
	if last := len(index.free) - 1; last >= 0 {
		handle := index.free[last]
		index.free = index.free[:last]
		return handle
	}
	index.items = append(index.items, indexedItem[T, ID]{})
	index.stamps = append(index.stamps, 0)
	return int32(len(index.items) - 1)
}

// nextStamp starts a query. On the rare wrap of the counter every mark is reset,
// so an item last seen 2³² queries ago is not taken for one seen in this one.
func (index *ChunkedIndex[T, ID]) nextStamp() uint32 {
	index.stamp++
	if index.stamp == 0 {
		clear(index.stamps)
		index.stamp = 1
	}
	return index.stamp
}

// cellRange is the inclusive range of cells box covers. A box edge exactly on a
// cell boundary counts the cell beyond it too; the overlap test sorts that out.
func (index *ChunkedIndex[T, ID]) cellRange(box gomath.AABB[T]) (minX int, minY int, maxX int, maxY int) {
	minX = floorToInt((box.Center.X - box.HalfWidth) * index.inverseCellSize)
	minY = floorToInt((box.Center.Y - box.HalfHeight) * index.inverseCellSize)
	maxX = floorToInt((box.Center.X + box.HalfWidth) * index.inverseCellSize)
	maxY = floorToInt((box.Center.Y + box.HalfHeight) * index.inverseCellSize)
	return minX, minY, maxX, maxY
}

// cell is the list of the cell at (x, y). With create it makes the chunk if
// there is none (the directory must already cover it); without, a missing chunk
// or a cell outside the directory gives nil.
func (index *ChunkedIndex[T, ID]) cell(x int, y int, create bool) *[]int32 {
	chunkX, cellX := floorDivide(x, index.chunkCells)
	chunkY, cellY := floorDivide(y, index.chunkCells)
	directoryX := chunkX - index.originX
	directoryY := chunkY - index.originY
	if directoryX < 0 || directoryY < 0 || directoryX >= index.directoryWidth || directoryY >= index.directoryHeight {
		return nil
	}
	slot := &index.directory[directoryY*index.directoryWidth+directoryX]
	if *slot == nil {
		if !create {
			return nil
		}
		*slot = &indexChunk{cells: make([][]int32, index.chunkCells*index.chunkCells)}
	}
	return &(*slot).cells[cellY*index.chunkCells+cellX]
}

// ensureChunks grows the directory to cover the cell range, keeping every chunk
// it already has. Growth in any direction, including below the origin, is a
// reallocation of pointers only.
func (index *ChunkedIndex[T, ID]) ensureChunks(minX int, minY int, maxX int, maxY int) {
	lowX, _ := floorDivide(minX, index.chunkCells)
	lowY, _ := floorDivide(minY, index.chunkCells)
	highX, _ := floorDivide(maxX, index.chunkCells)
	highY, _ := floorDivide(maxY, index.chunkCells)

	if index.directory != nil {
		if lowX >= index.originX && lowY >= index.originY &&
			highX < index.originX+index.directoryWidth && highY < index.originY+index.directoryHeight {
			return
		}
		lowX = min(lowX, index.originX)
		lowY = min(lowY, index.originY)
		highX = max(highX, index.originX+index.directoryWidth-1)
		highY = max(highY, index.originY+index.directoryHeight-1)
	}

	width := highX - lowX + 1
	height := highY - lowY + 1
	directory := make([]*indexChunk, width*height)
	for y := 0; y < index.directoryHeight; y++ {
		for x := 0; x < index.directoryWidth; x++ {
			newX := index.originX + x - lowX
			newY := index.originY + y - lowY
			directory[newY*width+newX] = index.directory[y*index.directoryWidth+x]
		}
	}
	index.directory = directory
	index.directoryWidth = width
	index.directoryHeight = height
	index.originX = lowX
	index.originY = lowY
}

// floorToInt is ⌊value⌋, for negative values too.
func floorToInt[T constraints.Float](value T) int {
	truncated := int(value)
	if T(truncated) > value {
		truncated--
	}
	return truncated
}

// floorDivide splits value into a quotient rounded down and a remainder in
// [0, divisor), for negative values too.
func floorDivide(value int, divisor int) (quotient int, remainder int) {
	quotient = value / divisor
	remainder = value % divisor
	if remainder < 0 {
		quotient--
		remainder += divisor
	}
	return quotient, remainder
}
