package spatial

import (
	"iter"
	"math"

	"github.com/wolfgarnet/gomath"
	"github.com/wolfgarnet/gomath/collider"
)

// BoxedItem3 is one item held by a Grid3: an identifier and the world-space box
// that bounds it. It is the 3D sibling of BoxedItem.
type BoxedItem3[T gomath.Float, ID comparable] struct {
	ItemID ID
	Box    gomath.Box3[T]
}

// MaxGrid3Cells caps how many cells a Build will allocate.
//
// A pathological scene — one item a mile wide next to a thousand small ones —
// would otherwise ask for a grid with more cells than it has items to put in
// them, which costs more to clear each rebuild than the partitioning saves.
// Build doubles the cell size until the grid fits under this.
const MaxGrid3Cells = 1 << 20

// Grid3 is a uniform grid over box-bounded items in 3D. It answers the two
// questions a 3D game asks of a spatial structure: which items overlap this
// volume (the collision broad phase), and what does this ray meet first
// (picking, hitscan, line of sight) — the second with cells visited strictly
// near-to-far, so the caller can stop as soon as a confirmed hit is nearer than
// the next cell can possibly be.
//
// # Why a grid and not an octree
//
// The access pattern this is built for is clear-and-rebuild every tick from
// moving transforms. A grid rebuilds in one linear pass with no tree to
// rebalance, and its cells are visited in ray order by a digital differential
// analyser. An octree adapts better to wildly non-uniform density and costs
// more to rebuild, which is the wrong trade when the whole structure is thrown
// away next frame.
//
// # Rebuild, not incremental update
//
// Add appends; Build sizes the grid to the items it was given and files them.
// Clear truncates rather than discards, so a steady scene allocates nothing
// after the first rebuild. There is deliberately no Remove or Relocate: the
// consumer's pattern is clear-and-rebuild, and an incremental structure is a
// different design with different evidence behind it.
//
// Queries build the grid themselves if items were added since the last Build,
// so forgetting the call costs a rebuild rather than silently returning
// nothing. Call Build explicitly to pay that cost at a point you choose.
//
// # No wrapping
//
// Unlike SimpleHashGrid, this grid does not wrap toroidally. Wrapping is
// harmless for a proximity query and wrong for a ray, which would leave one
// edge of the world and collect candidates from the far side of it. Positions
// outside the grid clamp to the edge cells instead.
//
// # Flat axes
//
// A cell size of zero on an axis means that axis is not subdivided: the grid is
// one cell deep there and the ray traversal never crosses a boundary on it.
// NewGroundGrid3 is that case for height, matching the ground-plane assumption
// a world of things standing on terrain can make. It costs the narrow phase one
// test per item standing under the ray and saves maintaining a third axis.
// Whether it is worth it is a property of the scene — see Grid3's benchmarks.
//
// A Grid3 is not safe for concurrent use: Add, Build and Clear mutate it, and
// the queries may build. Concurrent queries against a grid that is already
// built and that no goroutine is modifying are fine.
type Grid3[T gomath.Float, ID comparable] struct {
	// cells holds indices into entries, so an item spanning several cells is
	// stored once and referenced from each. The slices are truncated rather
	// than discarded between rebuilds.
	cells [][]int32

	// occupied lists the cells that hold anything, so a rebuild resets those
	// rather than sweeping the whole grid.
	//
	// A grid whose cells are about the size of the items in it is mostly empty
	// in 3D — a scene that fills a fiftieth of its cells is ordinary — and
	// sweeping all of them was the largest single cost in a rebuild, larger
	// than filing the items.
	//
	// What makes the short sweep safe is that no cell outside this list ever
	// holds anything: the list is filled after the grid has been resized, so
	// every index in it is valid for the cells slice a later rebuild starts
	// with, however much the grid grew or shrank in between.
	occupied []int32

	entries []BoxedItem3[T, ID]

	// lowCells is the lowest cell each entry occupies, parallel to entries. It
	// is what lets a query decide in constant time, and without a visited set,
	// which of the several cells holding an item is the one allowed to yield
	// it — see queryRange. Only the low corner is needed for that: the walk
	// already knows the item is in the cell it is looking at, so the high
	// corner would never be read.
	lowCells [][3]int32

	// cellSize, origin, low and high are per-axis rather than Vector3 because
	// every loop that uses them is a loop over axes.
	cellSize [3]T
	origin   [3]T
	low      [3]T
	high     [3]T

	// inverseCellSize turns the division in cellOf, which every item pays
	// twice on every rebuild, into a multiplication. It was half the cost of a
	// rebuild. A flat axis holds zero here and is skipped rather than scaled.
	inverseCellSize [3]T

	counts [3]int

	built bool
}

// NewGrid3 returns an empty grid whose cells have the given extents. An extent
// of zero or less on an axis leaves that axis unsubdivided — see Grid3 on flat
// axes.
//
// The useful cell size is roughly the size of the things being indexed: much
// smaller and a single item spans many cells, much larger and each cell holds a
// crowd the query has to test one by one.
func NewGrid3[T gomath.Float, ID comparable](cellSize gomath.Vector3[T]) *Grid3[T, ID] {
	grid := &Grid3[T, ID]{cellSize: axisComponents(cellSize)}
	grid.scaleChanged()

	return grid
}

// scaleChanged recomputes what is derived from the cell size.
func (grid *Grid3[T, ID]) scaleChanged() {
	for axis := range 3 {
		grid.inverseCellSize[axis] = 0
		if grid.cellSize[axis] > 0 {
			grid.inverseCellSize[axis] = 1 / grid.cellSize[axis]
		}
	}
}

// NewUniformGrid3 returns a grid of cubic cells.
func NewUniformGrid3[T gomath.Float, ID comparable](cellSize T) *Grid3[T, ID] {
	return NewGrid3[T, ID](gomath.Vector3[T]{X: cellSize, Y: cellSize, Z: cellSize})
}

// NewGroundGrid3 returns a grid that subdivides X and Z and ignores height, for
// a world that is a ground plane with things standing on it.
func NewGroundGrid3[T gomath.Float, ID comparable](cellSize T) *Grid3[T, ID] {
	return NewGrid3[T, ID](gomath.Vector3[T]{X: cellSize, Z: cellSize})
}

// Add stages an item with its world-space bound. It is filed into cells by the
// next Build.
func (grid *Grid3[T, ID]) Add(id ID, box gomath.Box3[T]) {
	grid.entries = append(grid.entries, BoxedItem3[T, ID]{ItemID: id, Box: box})
	grid.built = false
}

// Clear empties the grid, keeping its backing storage for the next rebuild.
func (grid *Grid3[T, ID]) Clear() {
	grid.entries = grid.entries[:0]
	grid.lowCells = grid.lowCells[:0]
	grid.built = false
}

// Len is the number of items staged or held.
func (grid *Grid3[T, ID]) Len() int {
	return len(grid.entries)
}

// CellSize is the cell extent per axis, which Build may have grown to keep the
// cell count under MaxGrid3Cells.
func (grid *Grid3[T, ID]) CellSize() gomath.Vector3[T] {
	return gomath.Vector3[T]{X: grid.cellSize[0], Y: grid.cellSize[1], Z: grid.cellSize[2]}
}

// CellCounts is the number of cells along each axis after the last build. A
// flat axis reports 1.
func (grid *Grid3[T, ID]) CellCounts() (x int, y int, z int) {
	grid.ensureBuilt()

	return grid.counts[0], grid.counts[1], grid.counts[2]
}

// OccupiedCells is how many cells hold at least one item. A grid where most
// items share a cell is not partitioning anything, so this is the number to
// look at when judging a cell size.
func (grid *Grid3[T, ID]) OccupiedCells() int {
	grid.ensureBuilt()

	return len(grid.occupied)
}

// Bounds is the box the grid covers: the union of its items' bounds, rounded
// out to whole cells on the subdivided axes. It is empty when the grid is.
func (grid *Grid3[T, ID]) Bounds() gomath.Box3[T] {
	grid.ensureBuilt()

	if len(grid.entries) == 0 {
		return gomath.Box3[T]{}
	}

	return gomath.NewBox3FromMinMax(
		gomath.Vector3[T]{X: grid.low[0], Y: grid.low[1], Z: grid.low[2]},
		gomath.Vector3[T]{X: grid.high[0], Y: grid.high[1], Z: grid.high[2]})
}

// Build sizes the grid to the staged items and files each into every cell its
// bound touches. Calling it on an already-built grid does nothing.
//
// An item wider than a cell belongs to all the cells it covers, or a ray
// crossing only its edge would never consider it.
func (grid *Grid3[T, ID]) Build() {
	if grid.built {
		return
	}
	grid.built = true

	for _, at := range grid.occupied {
		grid.cells[at] = grid.cells[at][:0]
	}
	grid.occupied = grid.occupied[:0]

	if len(grid.entries) == 0 {
		grid.cells = grid.cells[:0]
		grid.counts = [3]int{}

		return
	}

	// The grid is sized to the items rather than to a world whose extent this
	// package does not know.
	minimum := grid.entries[0].Box.Min()
	maximum := grid.entries[0].Box.Max()
	for index := 1; index < len(grid.entries); index++ {
		box := grid.entries[index].Box
		minimum = minimum.Min(box.Min())
		maximum = maximum.Max(box.Max())
	}

	grid.low = axisComponents(minimum)
	grid.origin = grid.low
	itemHigh := axisComponents(maximum)

	for {
		total := int64(1)
		for axis := range 3 {
			grid.counts[axis] = countForAxis(itemHigh[axis]-grid.origin[axis], grid.cellSize[axis])
			total *= int64(grid.counts[axis])
		}

		if total <= MaxGrid3Cells {
			break
		}

		for axis := range 3 {
			if grid.cellSize[axis] > 0 {
				grid.cellSize[axis] *= 2
			}
		}
		grid.scaleChanged()
	}

	// On a flat axis the grid's extent is whatever the items span, since there
	// are no cell boundaries to round out to.
	for axis := range 3 {
		if grid.cellSize[axis] > 0 {
			grid.high[axis] = grid.origin[axis] + T(grid.counts[axis])*grid.cellSize[axis]
		} else {
			grid.high[axis] = itemHigh[axis]
		}
	}

	total := grid.counts[0] * grid.counts[1] * grid.counts[2]
	if cap(grid.cells) < total {
		// Carry the old cells across rather than starting from nil ones. They
		// have all just been emptied, so which cell each now stands for does
		// not matter; what is being kept is the capacity they accumulated, and
		// throwing it away would make every rebuild of a growing scene
		// reallocate all of it.
		grown := make([][]int32, total)
		copy(grown, grid.cells)
		grid.cells = grown
	} else {
		grid.cells = grid.cells[:total]
	}

	if cap(grid.lowCells) < len(grid.entries) {
		grid.lowCells = make([][3]int32, len(grid.entries))
	} else {
		grid.lowCells = grid.lowCells[:len(grid.entries)]
	}

	for index := range grid.entries {
		box := grid.entries[index].Box
		low := grid.cellOf(box.Min())
		high := grid.cellOf(box.Max())

		grid.lowCells[index] = [3]int32{int32(low[0]), int32(low[1]), int32(low[2])}

		for z := low[2]; z <= high[2]; z++ {
			for y := low[1]; y <= high[1]; y++ {
				for x := low[0]; x <= high[0]; x++ {
					at := grid.cellIndex(x, y, z)
					if len(grid.cells[at]) == 0 {
						grid.occupied = append(grid.occupied, int32(at))
					}
					grid.cells[at] = append(grid.cells[at], int32(index))
				}
			}
		}
	}
}

func (grid *Grid3[T, ID]) ensureBuilt() {
	if !grid.built {
		grid.Build()
	}
}

// QueryBox iterates the items whose bounds overlap the given box, each once.
// The overlap test is exact, not a cell approximation.
func (grid *Grid3[T, ID]) QueryBox(box gomath.Box3[T]) iter.Seq[ID] {
	return func(yield func(ID) bool) {
		grid.queryRange(box, func(entry *BoxedItem3[T, ID]) bool {
			return collider.Box3Box3(box, entry.Box)
		}, yield)
	}
}

// QuerySphere iterates the items whose bounds overlap the given sphere, each
// once. The cells are chosen by the sphere's bounding box; the test that
// decides what is yielded is the exact sphere-box one.
func (grid *Grid3[T, ID]) QuerySphere(sphere gomath.Sphere[T]) iter.Seq[ID] {
	return func(yield func(ID) bool) {
		grid.queryRange(sphere.BoundingBox(), func(entry *BoxedItem3[T, ID]) bool {
			return collider.SphereBox3(sphere, entry.Box)
		}, yield)
	}
}

// queryRange walks the cells covering a box and yields each item that passes
// the narrow test, once.
//
// The de-duplication needs no visited set. An item appears in every cell its
// span covers, and the cells being walked are a box, so the intersection of the
// two is itself a box whose lowest cell is the componentwise maximum of the two
// lows. Yielding only from that cell yields each item exactly once, in constant
// time per candidate and with no state carried between cells.
func (grid *Grid3[T, ID]) queryRange(
	box gomath.Box3[T],
	overlaps func(entry *BoxedItem3[T, ID]) bool,
	yield func(ID) bool,
) {
	grid.ensureBuilt()

	if len(grid.entries) == 0 || !collider.Box3Box3(box, grid.Bounds()) {
		return
	}

	low := grid.cellOf(box.Min())
	high := grid.cellOf(box.Max())

	for z := low[2]; z <= high[2]; z++ {
		for y := low[1]; y <= high[1]; y++ {
			for x := low[0]; x <= high[0]; x++ {
				for _, candidate := range grid.cells[grid.cellIndex(x, y, z)] {
					lowest := &grid.lowCells[candidate]
					if x != ownerCell(int(lowest[0]), low[0]) ||
						y != ownerCell(int(lowest[1]), low[1]) ||
						z != ownerCell(int(lowest[2]), low[2]) {
						continue
					}

					entry := &grid.entries[candidate]
					if !overlaps(entry) {
						continue
					}

					if !yield(entry.ItemID) {
						return
					}
				}
			}
		}
	}
}

// RayCandidates iterates the items in the cells a ray crosses, nearest cell
// first, yielding each with the distance at which the ray ENTERED the cell
// holding it.
//
// # The ordering is the feature
//
// A caller that has confirmed a hit stops when that hit is nearer than the cell
// now being offered:
//
//	for cellEntry, id := range grid.RayCandidates(ray) {
//		if found && nearest < cellEntry {
//			break
//		}
//		// narrow phase against id, updating found and nearest
//	}
//
// The distance is per CELL and not per candidate on purpose. Everything in this
// cell and every cell after it is either at least this far away or overlaps an
// earlier cell and has already been offered, so the test above is safe to make
// on any item of the cell. A distance measured per candidate would let the
// caller abandon the rest of a cell, one of which may be nearer than the hit
// just found — the items within a cell are in no particular order.
//
// Because the direction of a Ray3 is normalised, the distance is in world units
// and compares directly against what a narrow-phase intersection returns.
//
// # What it does not do
//
// It does not test the ray against anything; whether a candidate is a real hit
// is a narrow-phase question about the item's true shape. An item spanning
// several cells along the ray is offered once per cell — a caller for which the
// repeat test is dearer than a lookup should keep its own seen-set.
//
// A ray that starts inside the grid begins at distance zero; one that starts
// outside it enters where it crosses; one that misses the grid yields nothing.
// On a flat axis the grid extends forever, so a ray passing high over a ground
// grid still collects the candidates beneath it and rejects them in the narrow
// phase.
func (grid *Grid3[T, ID]) RayCandidates(ray gomath.Ray3[T]) iter.Seq2[T, ID] {
	return func(yield func(T, ID) bool) {
		grid.ensureBuilt()

		if len(grid.entries) == 0 || !ray.IsValid() {
			return
		}

		entry, hit := grid.rayEntry(ray)
		if !hit {
			return
		}

		position := axisComponents(ray.At(entry))
		direction := axisComponents(ray.Direction)
		cell := grid.cellOf(ray.At(entry))

		// A digital differential analyser: each axis knows the distance at
		// which the ray next crosses one of its cell boundaries, and the walk
		// always steps whichever axis crosses soonest.
		var (
			step  [3]int
			next  [3]T
			delta [3]T
		)
		for axis := range 3 {
			step[axis], next[axis], delta[axis] = grid.axisStep(axis, position[axis], direction[axis], cell[axis])

			// cellOf scales by a reciprocal, so it can disagree with the exact
			// boundary arithmetic above by an ulp and put the first crossing a
			// hair behind the ray. Clamping here, once per cast, is what makes
			// the yielded distances monotonic unconditionally: from now on the
			// walk only ever adds a positive delta.
			if next[axis] < 0 {
				next[axis] = 0
			}
			next[axis] += entry
		}

		cellEntry := entry
		for {
			for _, candidate := range grid.cells[grid.cellIndex(cell[0], cell[1], cell[2])] {
				if !yield(cellEntry, grid.entries[candidate].ItemID) {
					return
				}
			}

			soonest := 0
			if next[1] < next[soonest] {
				soonest = 1
			}
			if next[2] < next[soonest] {
				soonest = 2
			}

			// Every axis is either flat or parallel to the ray: the ray never
			// leaves this cell. A ray straight down a ground grid is that case.
			if math.IsInf(float64(next[soonest]), 1) {
				return
			}

			// The next cell begins where this one ends.
			cellEntry = next[soonest]

			cell[soonest] += step[soonest]
			if cell[soonest] < 0 || cell[soonest] >= grid.counts[soonest] {
				return
			}
			next[soonest] += delta[soonest]
		}
	}
}

// axisStep sets up one axis of the ray traversal: which way its cell index
// moves, how far along the ray it first crosses a boundary — measured from the
// point passed in, which the caller shifts to a distance from the ray's origin
// — and the distance between boundaries thereafter.
func (grid *Grid3[T, ID]) axisStep(
	axis int,
	position T,
	direction T,
	cell int,
) (step int, next T, delta T) {
	infinity := T(math.Inf(1))

	// A flat axis has no boundaries to cross, and a direction parallel to this
	// axis never reaches the ones it has.
	if grid.cellSize[axis] <= 0 || direction == 0 {
		return 0, infinity, infinity
	}

	delta = gomath.Abs(grid.cellSize[axis] / direction)

	if direction > 0 {
		boundary := grid.origin[axis] + T(cell+1)*grid.cellSize[axis]

		return 1, (boundary - position) / direction, delta
	}

	boundary := grid.origin[axis] + T(cell)*grid.cellSize[axis]

	return -1, (boundary - position) / direction, delta
}

// rayEntry returns how far along a ray the grid begins, and whether the ray
// meets it at all. A ray starting inside enters at zero; one that never crosses
// the grid — a camera looking away from the crowd — is rejected here, before a
// single cell is visited.
//
// Flat axes are skipped, because they do not partition anything and so must not
// exclude anything either.
func (grid *Grid3[T, ID]) rayEntry(ray gomath.Ray3[T]) (T, bool) {
	origin := axisComponents(ray.Origin)
	direction := axisComponents(ray.Direction)

	entry := T(0)
	exit := T(math.Inf(1))

	for axis := range 3 {
		if grid.cellSize[axis] <= 0 {
			continue
		}

		if direction[axis] == 0 {
			if origin[axis] < grid.low[axis] || origin[axis] > grid.high[axis] {
				return 0, false
			}

			continue
		}

		near := (grid.low[axis] - origin[axis]) / direction[axis]
		far := (grid.high[axis] - origin[axis]) / direction[axis]
		if near > far {
			near, far = far, near
		}

		if near > entry {
			entry = near
		}
		if far < exit {
			exit = far
		}

		if entry > exit {
			return 0, false
		}
	}

	if exit < 0 {
		return 0, false
	}

	return entry, true
}

// Pairs iterates every pair of items that share a cell, each pair once, with no
// pair reported as itself.
//
// This is the broad phase: two items whose bounds overlap necessarily share a
// cell, so no overlapping pair is missed. Pairs that share a cell without
// overlapping are candidates the narrow phase rejects, which is what a broad
// phase is for.
//
// # Why there is no neighbour sweep
//
// SimpleHashGrid files an item into the one cell holding its position, so it
// has to sweep the eight neighbours as well and de-duplicate by only ever
// looking at the "next" four. That rule does not translate to 26 neighbours and
// does not need to: this grid files an item into every cell its bound covers,
// so two overlapping bounds always meet inside a cell and the neighbours are
// already accounted for.
//
// What that costs instead is that a pair can share several cells. The pair is
// yielded from the lowest cell they share — the componentwise maximum of their
// two span lows, which is in both spans because the cell being walked already
// is — so no visited set is needed to yield it once.
func (grid *Grid3[T, ID]) Pairs() iter.Seq2[ID, ID] {
	return func(yield func(ID, ID) bool) {
		grid.ensureBuilt()

		// Walking the occupancy list rather than the grid: a pair needs two
		// items in a cell, and in 3D the great majority of cells hold none.
		for _, at := range grid.occupied {
			cell := grid.cells[at]
			x, y, z := grid.cellCoordinates(int(at))

			for first := 0; first < len(cell); first++ {
				firstLowest := &grid.lowCells[cell[first]]

				for second := first + 1; second < len(cell); second++ {
					secondLowest := &grid.lowCells[cell[second]]

					if x != ownerCell(int(firstLowest[0]), int(secondLowest[0])) ||
						y != ownerCell(int(firstLowest[1]), int(secondLowest[1])) ||
						z != ownerCell(int(firstLowest[2]), int(secondLowest[2])) {
						continue
					}

					if !yield(grid.entries[cell[first]].ItemID, grid.entries[cell[second]].ItemID) {
						return
					}
				}
			}
		}
	}
}

// cellOf returns the cell holding a world position, clamped to the grid.
func (grid *Grid3[T, ID]) cellOf(point gomath.Vector3[T]) [3]int {
	coordinates := axisComponents(point)

	var cell [3]int
	for axis := range 3 {
		if grid.cellSize[axis] <= 0 {
			continue
		}

		cell[axis] = clampCell(
			int((coordinates[axis]-grid.origin[axis])*grid.inverseCellSize[axis]),
			grid.counts[axis])
	}

	return cell
}

func (grid *Grid3[T, ID]) cellIndex(x int, y int, z int) int {
	return (z*grid.counts[1]+y)*grid.counts[0] + x
}

func (grid *Grid3[T, ID]) cellCoordinates(index int) (x int, y int, z int) {
	x = index % grid.counts[0]
	rest := index / grid.counts[0]

	return x, rest % grid.counts[1], rest / grid.counts[1]
}

// ownerCell is the lowest cell two inclusive ranges share on one axis, given
// that they do share at least one.
func ownerCell(first int, second int) int {
	if first > second {
		return first
	}

	return second
}

// clampCell holds a cell index inside a grid of the given width.
func clampCell(value int, count int) int {
	if value < 0 {
		return 0
	}
	if value >= count {
		return count - 1
	}

	return value
}

// countForAxis is how many cells of the given size an extent needs.
//
// It refuses to answer with a number that would overflow the cell budget, so
// that a NaN or an extent astronomically larger than the cell size becomes a
// grid Build shrinks rather than an allocation it cannot make.
func countForAxis[T gomath.Float](extent T, cellSize T) int {
	if cellSize <= 0 || extent <= 0 {
		return 1
	}

	cells := float64(extent) / float64(cellSize)
	if !(cells < MaxGrid3Cells) {
		return MaxGrid3Cells + 1
	}

	return int(cells) + 1
}

func axisComponents[T gomath.Float](vector gomath.Vector3[T]) [3]T {
	return [3]T{vector.X, vector.Y, vector.Z}
}
