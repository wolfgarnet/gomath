package spatial

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wolfgarnet/gomath"
	"github.com/wolfgarnet/gomath/collider"
)

// scene is a random set of boxes used as the oracle's input as well as the
// grid's, so every query can be checked against the answer a linear scan gives.
type scene struct {
	boxes []gomath.Box3[float32]
}

// newScene builds count boxes scattered through a cube of the given extent,
// with sizes spanning an order of magnitude so that some items sit inside one
// cell and others straddle several.
func newScene(seed int64, count int, extent float32, spread float32) scene {
	random := rand.New(rand.NewSource(seed))

	boxes := make([]gomath.Box3[float32], count)
	for index := range boxes {
		half := 0.2 + random.Float32()*2
		boxes[index] = gomath.Box3[float32]{
			Center: gomath.Vector3[float32]{
				X: (random.Float32() - 0.5) * extent,
				Y: (random.Float32() - 0.5) * extent * spread,
				Z: (random.Float32() - 0.5) * extent,
			},
			HalfSize: gomath.Vector3[float32]{X: half, Y: half, Z: half},
		}
	}

	return scene{boxes: boxes}
}

func (s scene) fill(grid *Grid3[float32, int]) *Grid3[float32, int] {
	grid.Clear()
	for index, box := range s.boxes {
		grid.Add(index, box)
	}
	grid.Build()

	return grid
}

// gridShapes are the grid configurations every oracle test runs against, so a
// flat axis is covered by the same assertions as a fully subdivided one.
func gridShapes() []struct {
	name string
	make func() *Grid3[float32, int]
} {
	return []struct {
		name string
		make func() *Grid3[float32, int]
	}{
		{"cubic cells", func() *Grid3[float32, int] { return NewUniformGrid3[float32, int](4) }},
		{"small cells", func() *Grid3[float32, int] { return NewUniformGrid3[float32, int](1) }},
		{"ground plane", func() *Grid3[float32, int] { return NewGroundGrid3[float32, int](4) }},
		{"anisotropic", func() *Grid3[float32, int] {
			return NewGrid3[float32, int](gomath.Vector3[float32]{X: 3, Y: 8, Z: 5})
		}},
	}
}

func TestGrid3Empty(t *testing.T) {
	grid := NewUniformGrid3[float32, int](4)

	assert.Equal(t, 0, grid.Len())
	assert.Equal(t, 0, grid.OccupiedCells())
	assert.Equal(t, gomath.Box3[float32]{}, grid.Bounds())

	x, y, z := grid.CellCounts()
	assert.Equal(t, [3]int{0, 0, 0}, [3]int{x, y, z})

	for range grid.QueryBox(gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 10, Y: 10, Z: 10})) {
		t.Fatal("an empty grid yielded an item")
	}
	for range grid.QuerySphere(gomath.NewSphere(gomath.Vector3[float32]{}, 10)) {
		t.Fatal("an empty grid yielded an item")
	}
	for range grid.RayCandidates(gomath.NewRay3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 1})) {
		t.Fatal("an empty grid yielded an item")
	}
	for range grid.Pairs() {
		t.Fatal("an empty grid yielded a pair")
	}
}

func TestGrid3SingleItem(t *testing.T) {
	grid := NewUniformGrid3[float32, int](4)
	box := gomath.NewBox3(
		gomath.Vector3[float32]{X: 10, Y: 2, Z: -6},
		gomath.Vector3[float32]{X: 1, Y: 1, Z: 1})
	grid.Add(7, box)

	found := make([]int, 0, 1)
	for id := range grid.QueryBox(box) {
		found = append(found, id)
	}
	assert.Equal(t, []int{7}, found)

	// The grid is sized to its items, so it covers the one box it holds.
	assert.True(t, collider.Box3Box3(box, grid.Bounds()))

	for range grid.Pairs() {
		t.Fatal("a single item made a pair with itself")
	}
}

// TestGrid3QueryBoxMatchesLinearScan is the volume query against its oracle:
// the same set as a scan over every item, and each item exactly once even
// though items straddle cells.
func TestGrid3QueryBoxMatchesLinearScan(t *testing.T) {
	world := newScene(1, 400, 60, 1)

	queries := []gomath.Box3[float32]{
		gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 3, Y: 3, Z: 3}),
		gomath.NewBox3(gomath.Vector3[float32]{X: 20, Y: -5, Z: 12}, gomath.Vector3[float32]{X: 8, Y: 2, Z: 1}),
		// Larger than the whole grid.
		gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 500, Y: 500, Z: 500}),
		// Entirely outside it.
		gomath.NewBox3(gomath.Vector3[float32]{X: 900, Y: 900, Z: 900}, gomath.Vector3[float32]{X: 4, Y: 4, Z: 4}),
		// Degenerate: a point query.
		gomath.NewBox3(gomath.Vector3[float32]{X: -7, Y: 1, Z: 4}, gomath.Vector3[float32]{}),
	}

	for _, shape := range gridShapes() {
		t.Run(shape.name, func(t *testing.T) {
			grid := world.fill(shape.make())

			for number, query := range queries {
				expected := map[int]bool{}
				for index, box := range world.boxes {
					if collider.Box3Box3(query, box) {
						expected[index] = true
					}
				}

				seen := map[int]int{}
				for id := range grid.QueryBox(query) {
					seen[id]++
				}

				for id, times := range seen {
					assert.Equal(t, 1, times, "query %d yielded item %d %d times", number, id, times)
				}
				assert.Equal(t, len(expected), len(seen), "query %d returned the wrong number of items", number)
				for id := range expected {
					assert.Contains(t, seen, id, "query %d missed item %d", number, id)
				}
			}
		})
	}
}

func TestGrid3QuerySphereMatchesLinearScan(t *testing.T) {
	world := newScene(2, 400, 60, 1)

	queries := []gomath.Sphere[float32]{
		gomath.NewSphere(gomath.Vector3[float32]{}, 5),
		gomath.NewSphere(gomath.Vector3[float32]{X: 18, Y: 4, Z: -20}, 12),
		gomath.NewSphere(gomath.Vector3[float32]{}, 500),
		gomath.NewSphere(gomath.Vector3[float32]{X: 900, Y: 0, Z: 0}, 3),
	}

	for _, shape := range gridShapes() {
		t.Run(shape.name, func(t *testing.T) {
			grid := world.fill(shape.make())

			for number, query := range queries {
				expected := map[int]bool{}
				for index, box := range world.boxes {
					if collider.SphereBox3(query, box) {
						expected[index] = true
					}
				}

				seen := map[int]int{}
				for id := range grid.QuerySphere(query) {
					seen[id]++
				}

				for id, times := range seen {
					assert.Equal(t, 1, times, "query %d yielded item %d %d times", number, id, times)
				}
				assert.Equal(t, len(expected), len(seen), "query %d returned the wrong number of items", number)
				for id := range expected {
					assert.Contains(t, seen, id, "query %d missed item %d", number, id)
				}
			}
		})
	}
}

func TestGrid3QueryStopsWhenTheCallerBreaks(t *testing.T) {
	world := newScene(3, 200, 40, 1)
	grid := world.fill(NewUniformGrid3[float32, int](4))

	query := gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 100, Y: 100, Z: 100})

	count := 0
	for range grid.QueryBox(query) {
		count++
		if count == 3 {
			break
		}
	}

	assert.Equal(t, 3, count)
}

// TestGrid3RayVisitsCellsInOrder is the property the whole ray traversal exists
// for: the cell entry distance never goes backwards.
func TestGrid3RayVisitsCellsInOrder(t *testing.T) {
	world := newScene(4, 600, 60, 1)
	random := rand.New(rand.NewSource(11))

	for _, shape := range gridShapes() {
		t.Run(shape.name, func(t *testing.T) {
			grid := world.fill(shape.make())

			for cast := 0; cast < 200; cast++ {
				ray := world.randomRay(random)

				previous := float32(-1)
				for cellEntry := range grid.RayCandidates(ray) {
					require.GreaterOrEqual(t, cellEntry, previous,
						"cell entry distances went backwards on cast %d", cast)
					require.GreaterOrEqual(t, cellEntry, float32(0),
						"a cell was entered behind the ray's origin on cast %d", cast)
					previous = cellEntry
				}
			}
		})
	}
}

// TestGrid3RayFindsEveryHit checks the broad phase does not reject a real hit:
// every box the ray actually meets is offered as a candidate.
func TestGrid3RayFindsEveryHit(t *testing.T) {
	world := newScene(5, 600, 60, 1)
	random := rand.New(rand.NewSource(12))

	for _, shape := range gridShapes() {
		t.Run(shape.name, func(t *testing.T) {
			grid := world.fill(shape.make())

			for cast := 0; cast < 300; cast++ {
				ray := world.randomRay(random)

				offered := map[int]bool{}
				for _, id := range grid.RayCandidates(ray) {
					offered[id] = true
				}

				for index, box := range world.boxes {
					if _, hit := collider.Ray3Box3(ray, box); hit {
						require.True(t, offered[index],
							"cast %d missed item %d, which the ray really hits", cast, index)
					}
				}
			}
		})
	}
}

// TestGrid3RayEarlyExitFindsTheNearest is the traversal's contract stated the
// way a caller uses it: stopping as soon as the confirmed hit beats the cell
// now being offered still finds the nearest hit in the world.
func TestGrid3RayEarlyExitFindsTheNearest(t *testing.T) {
	world := newScene(6, 800, 60, 1)
	random := rand.New(rand.NewSource(13))

	for _, shape := range gridShapes() {
		t.Run(shape.name, func(t *testing.T) {
			grid := world.fill(shape.make())

			for cast := 0; cast < 300; cast++ {
				ray := world.randomRay(random)

				expectedDistance := float32(math.Inf(1))
				expected := -1
				for index, box := range world.boxes {
					distance, hit := collider.Ray3Box3(ray, box)
					if hit && distance < expectedDistance {
						expectedDistance, expected = distance, index
					}
				}

				nearest := float32(math.Inf(1))
				found := -1
				tested := 0
				for cellEntry, id := range grid.RayCandidates(ray) {
					if found >= 0 && nearest < cellEntry {
						break
					}

					tested++
					distance, hit := collider.Ray3Box3(ray, world.boxes[id])
					if hit && distance < nearest {
						nearest, found = distance, id
					}
				}

				require.Equal(t, expected, found, "cast %d found the wrong nearest hit", cast)
				if expected >= 0 {
					require.InDelta(t, expectedDistance, nearest, 1e-4, "cast %d", cast)
				}
				require.LessOrEqual(t, tested, len(world.boxes))
			}
		})
	}
}

func TestGrid3RayCases(t *testing.T) {
	grid := NewUniformGrid3[float32, int](2)
	// One box near the origin and one far along +X, so a ray down the X axis
	// meets them in a known order.
	grid.Add(1, gomath.NewBox3(gomath.Vector3[float32]{X: 1}, gomath.Vector3[float32]{X: 0.5, Y: 0.5, Z: 0.5}))
	grid.Add(2, gomath.NewBox3(gomath.Vector3[float32]{X: 21}, gomath.Vector3[float32]{X: 0.5, Y: 0.5, Z: 0.5}))

	tests := []struct {
		name     string
		ray      gomath.Ray3[float32]
		expected []int
	}{
		{
			name:     "starting outside, pointing in",
			ray:      gomath.NewRay3(gomath.Vector3[float32]{X: -50}, gomath.Vector3[float32]{X: 1}),
			expected: []int{1, 2},
		},
		{
			name:     "starting inside",
			ray:      gomath.NewRay3(gomath.Vector3[float32]{X: 1}, gomath.Vector3[float32]{X: 1}),
			expected: []int{1, 2},
		},
		{
			name:     "pointing away",
			ray:      gomath.NewRay3(gomath.Vector3[float32]{X: -50}, gomath.Vector3[float32]{X: -1}),
			expected: nil,
		},
		{
			name:     "missing the grid entirely",
			ray:      gomath.NewRay3(gomath.Vector3[float32]{X: -50, Y: 500}, gomath.Vector3[float32]{X: 1}),
			expected: nil,
		},
		{
			name:     "degenerate direction",
			ray:      gomath.Ray3[float32]{Origin: gomath.Vector3[float32]{X: 1}},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var order []int
			for _, id := range grid.RayCandidates(test.ray) {
				if len(order) == 0 || order[len(order)-1] != id {
					order = append(order, id)
				}
			}

			assert.Equal(t, test.expected, order)
		})
	}
}

// TestGrid3RayDistanceIsFromTheOrigin pins the units of the yielded distance. A
// distance measured from where the ray ENTERS the grid rather than from its
// origin would still be monotonic, and would still be safe, but would compare
// wrongly against what a narrow-phase intersection returns.
func TestGrid3RayDistanceIsFromTheOrigin(t *testing.T) {
	grid := NewUniformGrid3[float32, int](2)
	for step := 0; step < 10; step++ {
		grid.Add(step, gomath.NewBox3(
			gomath.Vector3[float32]{X: float32(step) * 4},
			gomath.Vector3[float32]{X: 0.5, Y: 0.5, Z: 0.5}))
	}

	// The grid starts at x = -0.5, so a ray from x = -100 enters at 99.5.
	ray := gomath.NewRay3(gomath.Vector3[float32]{X: -100}, gomath.Vector3[float32]{X: 1})

	for cellEntry, id := range grid.RayCandidates(ray) {
		distance, hit := collider.Ray3Box3(ray, gomath.NewBox3(
			gomath.Vector3[float32]{X: float32(id) * 4},
			gomath.Vector3[float32]{X: 0.5, Y: 0.5, Z: 0.5}))

		require.True(t, hit)
		assert.LessOrEqual(t, cellEntry, distance+1e-3,
			"item %d was offered at %v but is only hit at %v", id, cellEntry, distance)
		assert.Greater(t, cellEntry, float32(90),
			"the distance looks measured from the grid entry rather than the ray origin")
	}
}

func TestGrid3RayStraightDownAGroundGrid(t *testing.T) {
	grid := NewGroundGrid3[float32, int](4)
	grid.Add(1, gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 1, Y: 1, Z: 1}))
	grid.Add(2, gomath.NewBox3(gomath.Vector3[float32]{X: 40}, gomath.Vector3[float32]{X: 1, Y: 1, Z: 1}))

	// Height is not subdivided, so this crosses no boundary at all: the walk is
	// a single cell and must terminate.
	ray := gomath.NewRay3(gomath.Vector3[float32]{Y: 100}, gomath.Vector3[float32]{Y: -1})

	var found []int
	for _, id := range grid.RayCandidates(ray) {
		found = append(found, id)
	}

	assert.Equal(t, []int{1}, found)
}

// TestGrid3PairsMatchLinearScan is the broad phase against its oracle: every
// overlapping pair present, and no pair offered twice however many cells the
// two items share.
func TestGrid3PairsMatchLinearScan(t *testing.T) {
	world := newScene(7, 500, 30, 1)

	for _, shape := range gridShapes() {
		t.Run(shape.name, func(t *testing.T) {
			grid := world.fill(shape.make())

			seen := map[[2]int]int{}
			for first, second := range grid.Pairs() {
				require.NotEqual(t, first, second, "an item was paired with itself")

				key := [2]int{first, second}
				if first > second {
					key = [2]int{second, first}
				}
				seen[key]++
			}

			for key, times := range seen {
				assert.Equal(t, 1, times, "pair %v was yielded %d times", key, times)
			}

			for first := range world.boxes {
				for second := first + 1; second < len(world.boxes); second++ {
					if collider.Box3Box3(world.boxes[first], world.boxes[second]) {
						assert.Contains(t, seen, [2]int{first, second},
							"the overlapping pair %d/%d was never offered", first, second)
					}
				}
			}
		})
	}
}

func TestGrid3PairsStopWhenTheCallerBreaks(t *testing.T) {
	world := newScene(8, 200, 10, 1)
	grid := world.fill(NewUniformGrid3[float32, int](4))

	count := 0
	for range grid.Pairs() {
		count++
		if count == 5 {
			break
		}
	}

	assert.Equal(t, 5, count)
}

// TestGrid3RebuildReusesStorage is the access pattern this structure is built
// for: a steady scene must allocate nothing after the first rebuild.
func TestGrid3RebuildReusesStorage(t *testing.T) {
	world := newScene(9, 2000, 80, 1)
	grid := world.fill(NewUniformGrid3[float32, int](4))

	allocations := testing.AllocsPerRun(20, func() {
		grid.Clear()
		for index, box := range world.boxes {
			grid.Add(index, box)
		}
		grid.Build()
	})

	assert.Zero(t, allocations, "a steady scene should rebuild without allocating")
}

// TestGrid3RebuildIntoASmallerGrid guards the short clear sweep. A rebuild
// resets only the cells the previous one filled, so a grid that shrinks must
// still leave nothing behind in the slots it no longer covers — otherwise a
// later rebuild that grows again would find items from two scenes ago.
func TestGrid3RebuildIntoASmallerGrid(t *testing.T) {
	grid := NewUniformGrid3[float32, int](1)

	// A wide scene, then a narrow one, then a wide one again.
	for step := 0; step < 40; step++ {
		grid.Add(step, gomath.NewBox3(
			gomath.Vector3[float32]{X: float32(step) * 2},
			gomath.Vector3[float32]{X: 0.25, Y: 0.25, Z: 0.25}))
	}
	grid.Build()

	grid.Clear()
	grid.Add(100, gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 0.25, Y: 0.25, Z: 0.25}))
	grid.Build()

	everything := gomath.NewBox3(gomath.Vector3[float32]{X: 40}, gomath.Vector3[float32]{X: 200, Y: 200, Z: 200})

	var found []int
	for id := range grid.QueryBox(everything) {
		found = append(found, id)
	}
	assert.Equal(t, []int{100}, found, "the shrunk grid kept items from the wider one")

	grid.Clear()
	for step := 0; step < 40; step++ {
		grid.Add(step, gomath.NewBox3(
			gomath.Vector3[float32]{X: float32(step) * 2},
			gomath.Vector3[float32]{X: 0.25, Y: 0.25, Z: 0.25}))
	}
	grid.Build()

	count := 0
	for range grid.QueryBox(everything) {
		count++
	}
	assert.Equal(t, 40, count, "regrowing the grid resurrected stale cell contents")
}

// TestGrid3QueriesBuildLazily guards the footgun: a query on a grid with items
// added since the last Build must return them rather than nothing.
func TestGrid3QueriesBuildLazily(t *testing.T) {
	grid := NewUniformGrid3[float32, int](4)
	box := gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 1, Y: 1, Z: 1})
	grid.Add(1, box)

	found := 0
	for range grid.QueryBox(box) {
		found++
	}
	assert.Equal(t, 1, found, "a query did not build the grid it was asked about")

	grid.Add(2, box)
	found = 0
	for range grid.QueryBox(box) {
		found++
	}
	assert.Equal(t, 2, found, "a query did not rebuild after an item was added")
}

// TestGrid3ClampsCellCount checks the guard against a scene that would ask for
// more cells than it is worth clearing: one enormous item next to small ones.
func TestGrid3ClampsCellCount(t *testing.T) {
	grid := NewUniformGrid3[float32, int](0.001)
	grid.Add(1, gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 1e6, Y: 1e6, Z: 1e6}))
	grid.Add(2, gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 1, Y: 1, Z: 1}))

	x, y, z := grid.CellCounts()
	assert.LessOrEqual(t, x*y*z, MaxGrid3Cells)
	assert.Greater(t, grid.CellSize().X, float32(0.001), "the cell size should have been grown")

	found := 0
	for range grid.QueryBox(gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 2, Y: 2, Z: 2})) {
		found++
	}
	assert.Equal(t, 2, found)
}

func TestGrid3FlatAxisIsOneCellDeep(t *testing.T) {
	grid := NewGroundGrid3[float32, int](4)
	for step := 0; step < 20; step++ {
		grid.Add(step, gomath.NewBox3(
			gomath.Vector3[float32]{Y: float32(step) * 50},
			gomath.Vector3[float32]{X: 1, Y: 1, Z: 1}))
	}

	_, y, _ := grid.CellCounts()
	assert.Equal(t, 1, y, "a flat axis must not be subdivided")

	// Every item stands in the same ground cell, so a box query there sees them
	// all as candidates — and the exact test still narrows it to the one.
	found := 0
	for range grid.QueryBox(gomath.NewBox3(gomath.Vector3[float32]{}, gomath.Vector3[float32]{X: 1, Y: 1, Z: 1})) {
		found++
	}
	assert.Equal(t, 1, found)
}

// randomRay returns a ray somewhere around the scene. Two casts in three are
// aimed at an item, because a ray thrown in a uniformly random direction
// through a sparse world almost never strikes anything, and a ray that hits
// nothing exercises neither the ordering nor the early exit.
func (s scene) randomRay(random *rand.Rand) gomath.Ray3[float32] {
	origin := gomath.Vector3[float32]{
		X: (random.Float32() - 0.5) * 200,
		Y: (random.Float32() - 0.5) * 200,
		Z: (random.Float32() - 0.5) * 200,
	}

	if random.Intn(3) > 0 {
		target := s.boxes[random.Intn(len(s.boxes))].Center
		target.X += (random.Float32() - 0.5) * 2
		target.Y += (random.Float32() - 0.5) * 2
		target.Z += (random.Float32() - 0.5) * 2

		return gomath.NewRay3Between(origin, target)
	}

	direction := gomath.Vector3[float32]{
		X: random.Float32() - 0.5,
		Y: random.Float32() - 0.5,
		Z: random.Float32() - 0.5,
	}
	if direction.IsZero() {
		direction = gomath.Vector3[float32]{X: 1}
	}

	return gomath.NewRay3(origin, direction)
}
