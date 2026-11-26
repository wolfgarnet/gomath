package spatial

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wolfgarnet/gomath"
)

type sweepCell interface {
	Size() int
}

// closePairsSimple and closePairsChunked run the broad-phase sweep a collision
// system runs — every cell in order, each item against the rest of its cell and
// then the cell's four next neighbours — and list, in the order met, the pairs
// within reach. That ordered list is what decides the order contacts are
// appended in, and so what a simulation does next.
func closePairsSimple(grid *SimpleHashGrid[float32, int], positions []gomath.Vector2[float32], reach float32) [][2]int {
	var pairs [][2]int
	test := func(a int, b int) {
		if positions[a].DistanceToSquared(positions[b]) <= reach*reach {
			pairs = append(pairs, [2]int{a, b})
		}
	}
	for _, cell := range grid.IterateCells() {
		if cell.Size() == 0 {
			continue
		}
		for index, item := range cell.IterateContent() {
			for other := range cell.IterateContentFrom(index + 1) {
				test(item, other)
			}
			for neighbour := range cell.IterateNextNeighbors() {
				for _, other := range neighbour.IterateContent() {
					test(item, other)
				}
			}
		}
	}
	return pairs
}

func closePairsChunked(grid *ChunkedGrid[float32, int], positions []gomath.Vector2[float32], reach float32) [][2]int {
	var pairs [][2]int
	test := func(a int, b int) {
		if positions[a].DistanceToSquared(positions[b]) <= reach*reach {
			pairs = append(pairs, [2]int{a, b})
		}
	}
	for cell := range grid.IterateCells() {
		for index, item := range cell.IterateContent() {
			for other := range cell.IterateContentFrom(index + 1) {
				test(item, other)
			}
			for neighbour := range cell.IterateNextNeighbors() {
				for _, other := range neighbour.IterateContent() {
					test(item, other)
				}
			}
		}
	}
	return pairs
}

func closePairsSwept(grid *ChunkedGrid[float32, int], positions []gomath.Vector2[float32], reach float32) [][2]int {
	var pairs [][2]int
	grid.SweepPairs(func(first *int, second *int) {
		if positions[*first].DistanceToSquared(positions[*second]) <= reach*reach {
			pairs = append(pairs, [2]int{*first, *second})
		}
	})
	return pairs
}

func randomPositions(random *rand.Rand, count int, minimum float32, maximum float32) []gomath.Vector2[float32] {
	positions := make([]gomath.Vector2[float32], count)
	for index := range positions {
		positions[index] = gomath.Vector2[float32]{
			X: minimum + random.Float32()*(maximum-minimum),
			Y: minimum + random.Float32()*(maximum-minimum),
		}
	}
	return positions
}

// TestChunkedGridSweepsAsTheHashGridDoesOnASmallWorld pins what lets the chunked
// grid replace the 64 × 64 hash grid without changing a simulation: on a world
// inside the hash grid's span, the sweep meets the same close pairs in the same
// order. Clustered crowds and an even spread, so both crowded and thinly held cells
// are walked.
func TestChunkedGridSweepsAsTheHashGridDoesOnASmallWorld(t *testing.T) {
	const cellSize, reach = 24, 16
	span := float32(64 * cellSize) // the 64 × 64 grid pixelmaster's collisions use

	tests := []struct {
		name      string
		positions func(random *rand.Rand) []gomath.Vector2[float32]
	}{
		{name: "an even spread", positions: func(random *rand.Rand) []gomath.Vector2[float32] {
			return randomPositions(random, 3000, 0, span-1)
		}},
		{name: "a dense crowd", positions: func(random *rand.Rand) []gomath.Vector2[float32] {
			return randomPositions(random, 3000, 600, 700)
		}},
		{name: "crowds on cell boundaries", positions: func(random *rand.Rand) []gomath.Vector2[float32] {
			positions := randomPositions(random, 2000, 0, span-1)
			for index := range positions {
				if index%2 == 0 {
					positions[index].X = float32(int(positions[index].X/cellSize)) * cellSize
				}
			}
			return positions
		}},
	}

	for seed, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			positions := tc.positions(rand.New(rand.NewSource(int64(seed + 1))))
			simple := NewSimpleHashGrid[float32, int](40, 40, cellSize)
			chunked := NewChunkedGrid[float32, int](cellSize, 5)
			for index, position := range positions {
				simple.AddItem(index, position)
				chunked.AddItem(index, position)
			}

			want := closePairsSimple(simple, positions, reach)
			require.NotEmpty(t, want, "no close pairs; the comparison would be vacuous")
			assert.Equal(t, want, closePairsChunked(chunked, positions, reach), "the iterator sweep")
			assert.Equal(t, want, closePairsSwept(chunked, positions, reach), "SweepPairs")
		})
	}
}

// TestChunkedGridFindsEveryClosePairAnywhere: on a world far larger than any
// fixed grid, negative coordinates included, the sweep meets every pair within
// reach exactly once — checked against all pairs by brute force.
func TestChunkedGridFindsEveryClosePairAnywhere(t *testing.T) {
	const cellSize, reach = 24, 16
	random := rand.New(rand.NewSource(7))
	positions := randomPositions(random, 1500, -20000, 20000)
	// A few tight clusters, so there are pairs to find in a world this chunked.
	for cluster := range 20 {
		centre := randomPositions(random, 1, -20000, 20000)[0]
		for range 30 {
			positions = append(positions, centre.Add(gomath.Vector2[float32]{X: random.Float32() * 40, Y: random.Float32() * 40}))
		}
		_ = cluster
	}

	grid := NewChunkedGrid[float32, int](cellSize, 5)
	for index, position := range positions {
		grid.AddItem(index, position)
	}

	var want [][2]int
	for a := range positions {
		for b := a + 1; b < len(positions); b++ {
			if positions[a].DistanceToSquared(positions[b]) <= reach*reach {
				want = append(want, [2]int{a, b})
			}
		}
	}
	got := closePairsChunked(grid, positions, reach)
	for index, pair := range got {
		if pair[0] > pair[1] {
			got[index] = [2]int{pair[1], pair[0]}
		}
	}
	slices.SortFunc(got, func(x [2]int, y [2]int) int {
		if x[0] != y[0] {
			return x[0] - y[0]
		}
		return x[1] - y[1]
	})
	require.NotEmpty(t, want)
	assert.Equal(t, want, got)
}

// TestChunkedGridQueries covers the two rectangle queries: a cell range yields
// exactly what the hash grid's rectangle walk over the same cells does, and an
// AABB query finds every item whose position is inside the box — including
// across a cell boundary, which the hash grid's rectangle query misses.
func TestChunkedGridQueries(t *testing.T) {
	random := rand.New(rand.NewSource(3))
	positions := randomPositions(random, 2000, 0, 1500)
	simple := NewSimpleHashGrid[float32, int](40, 40, 24)
	chunked := NewChunkedGrid[float32, int](24, 5)
	for index, position := range positions {
		simple.AddItem(index, position)
		chunked.AddItem(index, position)
	}

	t.Run("a cell range walks as the hash grid's rectangle does", func(t *testing.T) {
		for range 500 {
			box := gomath.AABB2[float32]{MinX: random.Float32() * 1400, MinY: random.Float32() * 1400}
			box.MaxX = box.MinX + random.Float32()*80
			box.MaxY = box.MinY + random.Float32()*80
			var want []int
			for item := range simple.AABBIterator(box) {
				want = append(want, item)
			}
			// The cells the hash grid's rectangle covers: from the minimum
			// corner's cell, as many as the box's extent in whole cells, plus one.
			minX, minY := chunked.CellCoordinate(box.MinX), chunked.CellCoordinate(box.MinY)
			countX := int((box.MaxX-box.MinX)/24) + 1
			countY := int((box.MaxY-box.MinY)/24) + 1
			var got []int
			for item := range chunked.CellRangeIterator(minX, minY, minX+countX-1, minY+countY-1) {
				got = append(got, item)
			}
			require.Equal(t, want, got, "box %v", box)
		}
	})

	t.Run("an AABB query finds every item positioned inside the box", func(t *testing.T) {
		missedByHashGrid := 0
		for range 500 {
			box := gomath.AABB2[float32]{MinX: random.Float32() * 1400, MinY: random.Float32() * 1400}
			box.MaxX = box.MinX + random.Float32()*60
			box.MaxY = box.MinY + random.Float32()*60
			found := map[int]bool{}
			for item := range chunked.AABBIterator(box) {
				found[item] = true
			}
			foundByHashGrid := map[int]bool{}
			for item := range simple.AABBIterator(box) {
				foundByHashGrid[item] = true
			}
			for index, position := range positions {
				inside := position.X >= box.MinX && position.X <= box.MaxX && position.Y >= box.MinY && position.Y <= box.MaxY
				if inside {
					assert.True(t, found[index], "item %d at %v inside %v not found", index, position, box)
					if !foundByHashGrid[index] {
						missedByHashGrid++
					}
				}
			}
		}
		t.Logf("items inside a query box the hash grid's rectangle query missed: %d", missedByHashGrid)
	})
}

// TestChunkedGridRebuild: clearing and refilling keeps nothing from the last
// fill, and once the grid has held its largest population a rebuild allocates
// nothing.
func TestChunkedGridRebuild(t *testing.T) {
	random := rand.New(rand.NewSource(5))
	grid := NewChunkedGrid[float32, int](24, 5)
	first := randomPositions(random, 2000, 0, 3000)
	second := randomPositions(random, 2000, 5000, 9000)

	for index, position := range first {
		grid.AddItem(index, position)
	}
	grid.Clear()
	for index, position := range second {
		grid.AddItem(index, position)
	}
	assert.Equal(t, len(second), grid.NumberOfElements())
	for cell := range grid.IterateCells() {
		for _, item := range cell.IterateContent() {
			cellX, cellY := cell.Coordinates()
			assert.Equal(t, grid.CellCoordinate(second[item].X), cellX, "item %d in the wrong cell", item)
			assert.Equal(t, grid.CellCoordinate(second[item].Y), cellY)
		}
	}
	assert.Zero(t, grid.Cell(grid.CellCoordinate(first[0].X), grid.CellCoordinate(first[0].Y)).Size(), "a cell from the first fill survived Clear")

	fill := 0
	allocations := testing.AllocsPerRun(50, func() {
		grid.Clear()
		positions := first
		if fill%2 == 1 {
			positions = second
		}
		for index, position := range positions {
			grid.AddItem(index, position)
		}
		for range grid.IterateCells() {
		}
		fill++
	})
	assert.Zero(t, allocations)
}

// BenchmarkBroadPhaseRebuild compares the rebuilt broad phase on the hash grid
// and the chunked grid: on a world inside the hash grid's span, where only the
// structures differ, and on a wider one, where the hash grid folds.
func BenchmarkBroadPhaseRebuild(b *testing.B) {
	for _, world := range []struct {
		name string
		size float32
	}{{"world 1536", 1536}, {"world 2560", 2560}} {
		for _, items := range []int{1000, 4000} {
			random := rand.New(rand.NewSource(1))
			positions := randomPositions(random, items, 0, world.size-1)
			b.Run(fmt.Sprintf("%s/items=%d/hash grid", world.name, items), func(b *testing.B) {
				grid := NewSimpleHashGrid[float32, int](40, 40, 24)
				for b.Loop() {
					grid.Clear()
					for index, position := range positions {
						grid.AddItem(index, position)
					}
					_ = closePairsSimple(grid, positions, 16)
				}
			})
			b.Run(fmt.Sprintf("%s/items=%d/chunked grid", world.name, items), func(b *testing.B) {
				grid := NewChunkedGrid[float32, int](24, 5)
				for b.Loop() {
					grid.Clear()
					for index, position := range positions {
						grid.AddItem(index, position)
					}
					_ = closePairsChunked(grid, positions, 16)
				}
			})
			b.Run(fmt.Sprintf("%s/items=%d/chunked grid SweepPairs", world.name, items), func(b *testing.B) {
				grid := NewChunkedGrid[float32, int](24, 5)
				for b.Loop() {
					grid.Clear()
					for index, position := range positions {
						grid.AddItem(index, position)
					}
					_ = closePairsSwept(grid, positions, 16)
				}
			})
		}
	}
}
