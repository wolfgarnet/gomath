package spatial

import (
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wolfgarnet/gomath"
)

func box(centerX float32, centerY float32, halfWidth float32, halfHeight float32) gomath.AABB[float32] {
	return gomath.AABB[float32]{Center: gomath.Vector2[float32]{X: centerX, Y: centerY}, HalfWidth: halfWidth, HalfHeight: halfHeight}
}

func collect(index *ChunkedIndex[float32, int], bounds gomath.AABB[float32]) []int {
	var ids []int
	for item := range index.Iterator(bounds) {
		ids = append(ids, item.ItemID)
	}
	return ids
}

// TestChunkedIndexAnswersWhatTheQuadtreeAnswers holds the index to the
// structure it replaces: over random items and random queries — big and small,
// thin wall-like runs, boxes straddling chunk and cell boundaries, negative
// coordinates — it must yield exactly the set a BoxedQuadtree yields, each item
// once. Order may differ; the set may not.
func TestChunkedIndexAnswersWhatTheQuadtreeAnswers(t *testing.T) {
	tests := []struct {
		name       string
		items      int
		cellSize   float32
		chunkCells int
		maxHalf    float32
	}{
		{name: "tile-sized cells, small regions", items: 300, cellSize: 32, chunkCells: 16, maxHalf: 48},
		{name: "small chunks, long wall runs", items: 200, cellSize: 32, chunkCells: 2, maxHalf: 400},
		{name: "big cells", items: 300, cellSize: 200, chunkCells: 4, maxHalf: 64},
		{name: "one-cell chunks", items: 100, cellSize: 16, chunkCells: 1, maxHalf: 40},
	}

	for seed, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			random := rand.New(rand.NewSource(int64(seed + 1)))
			const extent = 2000 // items land in [-extent, extent]
			quadtree := NewBoxedQuadtree[float32, int](0, 0, 4*extent)
			index := NewChunkedIndex[float32, int](tc.cellSize, tc.chunkCells)

			for id := range tc.items {
				halfWidth := 1 + random.Float32()*tc.maxHalf
				halfHeight := 1 + random.Float32()*tc.maxHalf
				if random.Intn(3) == 0 {
					halfHeight = tc.cellSize / 2 // a one-cell-thick run, like a wall
				}
				// Snap some to the cell grid, so edges sit exactly on boundaries.
				centerX := (random.Float32()*2 - 1) * extent
				centerY := (random.Float32()*2 - 1) * extent
				if random.Intn(2) == 0 {
					centerX = float32(math.Round(float64(centerX/tc.cellSize))) * tc.cellSize
					centerY = float32(math.Round(float64(centerY/tc.cellSize))) * tc.cellSize
				}
				item := BoxedItem[float32, int]{ItemID: id, AABB: box(centerX, centerY, halfWidth, halfHeight)}
				_, err := quadtree.Insert(item)
				require.NoError(t, err)
				index.Insert(item)
			}
			require.Equal(t, tc.items, index.Size())

			found := 0
			for range 2000 {
				halfWidth := random.Float32() * 300
				halfHeight := random.Float32() * 300
				if random.Intn(4) == 0 {
					halfWidth, halfHeight = 8, 8 // a creep's broad phase
				}
				bounds := box((random.Float32()*2-1)*extent*1.2, (random.Float32()*2-1)*extent*1.2, halfWidth, halfHeight)

				var want []int
				for item := range quadtree.Iterator(bounds) {
					want = append(want, item.ItemID)
				}
				got := collect(index, bounds)

				assert.Len(t, got, len(slices.Compact(slices.Sorted(slices.Values(got)))), "an item yielded twice for %v", bounds)
				slices.Sort(want)
				slices.Sort(got)
				require.Equal(t, want, got, "query %v", bounds)
				found += len(got)
			}
			require.Positive(t, found, "the queries found nothing; the comparison would be vacuous")
		})
	}
}

// TestChunkedIndexChanges covers what a static index exists to do without a
// rebuild: grow in every direction, take items out without leaving them behind
// in any cell, and replace an item re-inserted under its ID.
func TestChunkedIndexChanges(t *testing.T) {
	everywhere := box(0, 0, 1e6, 1e6)

	tests := []struct {
		name  string
		build func(index *ChunkedIndex[float32, int])
		query gomath.AABB[float32]
		want  []int
	}{
		{
			name: "grows below and left of the first item",
			build: func(index *ChunkedIndex[float32, int]) {
				index.Insert(BoxedItem[float32, int]{ItemID: 1, AABB: box(10, 10, 4, 4)})
				index.Insert(BoxedItem[float32, int]{ItemID: 2, AABB: box(-5000, -7000, 4, 4)})
				index.Insert(BoxedItem[float32, int]{ItemID: 3, AABB: box(9000, -100, 4, 4)})
			},
			query: everywhere,
			want:  []int{1, 2, 3},
		},
		{
			name: "the first item is still found after growth",
			build: func(index *ChunkedIndex[float32, int]) {
				index.Insert(BoxedItem[float32, int]{ItemID: 1, AABB: box(10, 10, 4, 4)})
				index.Insert(BoxedItem[float32, int]{ItemID: 2, AABB: box(-5000, -7000, 4, 4)})
			},
			query: box(10, 10, 1, 1),
			want:  []int{1},
		},
		{
			name: "a removed item is in no cell it covered",
			build: func(index *ChunkedIndex[float32, int]) {
				index.Insert(BoxedItem[float32, int]{ItemID: 1, AABB: box(0, 0, 500, 16)}) // spans many cells and chunks
				index.Insert(BoxedItem[float32, int]{ItemID: 2, AABB: box(0, 0, 4, 4)})
				index.Remove(1)
			},
			query: everywhere,
			want:  []int{2},
		},
		{
			name: "a freed slot is reused without reviving its old item",
			build: func(index *ChunkedIndex[float32, int]) {
				index.Insert(BoxedItem[float32, int]{ItemID: 1, AABB: box(0, 0, 4, 4)})
				index.Remove(1)
				index.Insert(BoxedItem[float32, int]{ItemID: 7, AABB: box(1000, 1000, 4, 4)})
			},
			query: box(0, 0, 10, 10),
			want:  nil,
		},
		{
			name: "re-inserting an ID moves the item",
			build: func(index *ChunkedIndex[float32, int]) {
				index.Insert(BoxedItem[float32, int]{ItemID: 1, AABB: box(0, 0, 4, 4)})
				index.Insert(BoxedItem[float32, int]{ItemID: 1, AABB: box(3000, 0, 4, 4)})
			},
			query: box(0, 0, 10, 10),
			want:  nil,
		},
		{
			name: "boxes that only touch overlap, as in the quadtree",
			build: func(index *ChunkedIndex[float32, int]) {
				index.Insert(BoxedItem[float32, int]{ItemID: 1, AABB: box(16, 16, 16, 16)}) // [0,32]
			},
			query: box(40, 16, 8, 4), // [32,48]: touches at x=32
			want:  []int{1},
		},
		{
			name: "clear empties it",
			build: func(index *ChunkedIndex[float32, int]) {
				index.Insert(BoxedItem[float32, int]{ItemID: 1, AABB: box(0, 0, 4, 4)})
				index.Clear()
			},
			query: everywhere,
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			index := NewChunkedIndex[float32, int](32, 4)
			tc.build(index)
			got := collect(index, tc.query)
			slices.Sort(got)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestChunkedIndexStampWrap: the per-query marks that keep an item from being
// yielded twice survive the counter wrapping round.
func TestChunkedIndexStampWrap(t *testing.T) {
	index := NewChunkedIndex[float32, int](32, 4)
	index.Insert(BoxedItem[float32, int]{ItemID: 1, AABB: box(0, 0, 100, 100)})
	index.stamp = math.MaxUint32 - 1

	for range 4 {
		assert.Equal(t, []int{1}, collect(index, box(0, 0, 200, 200)))
	}
}

func BenchmarkRegionLookup(b *testing.B) {
	// A map's worth of wall runs and a few regions, queried with a creep's
	// broad-phase box — the shape of the per-creep region lookups.
	random := rand.New(rand.NewSource(1))
	quadtree := NewBoxedQuadtree[float32, int](1280, 800, 2048)
	index := NewChunkedIndex[float32, int](32, 32)
	for id := range 135 {
		var item BoxedItem[float32, int]
		if id < 126 {
			length := 32 + random.Float32()*1200
			if random.Intn(2) == 0 {
				item = BoxedItem[float32, int]{ItemID: id, AABB: box(random.Float32()*2560, random.Float32()*1600, length/2, 16)}
			} else {
				item = BoxedItem[float32, int]{ItemID: id, AABB: box(random.Float32()*2560, random.Float32()*1600, 16, length/2)}
			}
		} else {
			item = BoxedItem[float32, int]{ItemID: id, AABB: box(random.Float32()*2560, random.Float32()*1600, 48, 48)}
		}
		_, _ = quadtree.Insert(item)
		index.Insert(item)
	}
	queries := make([]gomath.AABB[float32], 4096)
	for position := range queries {
		queries[position] = box(random.Float32()*2560, random.Float32()*1600, 8, 8)
	}

	b.Run("quadtree", func(b *testing.B) {
		hits := 0
		for b.Loop() {
			for _, query := range queries {
				for range quadtree.Iterator(query) {
					hits++
				}
			}
		}
		_ = hits
	})
	b.Run("chunked index", func(b *testing.B) {
		hits := 0
		for b.Loop() {
			for _, query := range queries {
				for range index.Iterator(query) {
					hits++
				}
			}
		}
		_ = hits
	})
}

// TestChunkedIndexQueriesNest: a query run inside another's loop answers in
// full, and so does the outer one — the inner keeps its own record of what it
// has yielded rather than overwriting the outer's marks, which would have made
// the outer skip every item the inner had already seen.
func TestChunkedIndexQueriesNest(t *testing.T) {
	index := NewChunkedIndex[float32, int](32, 4)
	for id := range 40 {
		// Wide items, so each sits in many cells and deduplication matters.
		index.Insert(BoxedItem[float32, int]{ItemID: id, AABB: box(float32(id*50), 0, 80, 80)})
	}
	everywhere := box(1000, 0, 5000, 5000)

	outer := map[int]int{}
	for item := range index.Iterator(everywhere) {
		outer[item.ItemID]++
		inner := map[int]int{}
		for nested := range index.Iterator(everywhere) {
			inner[nested.ItemID]++
		}
		require.Len(t, inner, 40, "a nested query found every item")
		for id, count := range inner {
			require.Equal(t, 1, count, "the nested query yielded item %d once", id)
		}
	}
	assert.Len(t, outer, 40, "the outer query found every item, nested queries notwithstanding")
	for id, count := range outer {
		assert.Equal(t, 1, count, "the outer query yielded item %d once", id)
	}

	// And after the nesting, a plain query is back on the fast path and right.
	assert.Len(t, collect(index, everywhere), 40)
}
