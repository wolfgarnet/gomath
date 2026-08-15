package spatial

import (
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/wolfgarnet/gomath"
	"github.com/wolfgarnet/gomath/collider"
)

// The scale is 5 000 items because that is the number pixelmaster's
// render3d/pickindex.go measures its rebuild at, which makes the comparison
// direct: 411 µs there, for a grid over two axes rather than three.
const benchmarkItems = 5000

// benchmarkScenes are the two worlds the height question turns on. "flat" is a
// ground plane with things standing on it — pickindex's assumption. "spread"
// has as much vertical extent as horizontal, which is what a world with
// bridges or flying units looks like.
var benchmarkScenes = []struct {
	name   string
	spread float32
}{
	{"flat", 0.05},
	{"spread", 1},
}

var benchmarkGrids = []struct {
	name string
	make func() *Grid3[float32, int]
}{
	{"subdividedY", func() *Grid3[float32, int] { return NewUniformGrid3[float32, int](4) }},
	{"flatY", func() *Grid3[float32, int] { return NewGroundGrid3[float32, int](4) }},
}

func BenchmarkGrid3Rebuild(b *testing.B) {
	for _, world := range benchmarkScenes {
		for _, shape := range benchmarkGrids {
			b.Run(world.name+"/"+shape.name, func(b *testing.B) {
				boxes := newScene(1, benchmarkItems, 200, world.spread).boxes
				grid := shape.make()

				// The first rebuild allocates the backing storage; the ones
				// being measured must not.
				for index, box := range boxes {
					grid.Add(index, box)
				}
				grid.Build()

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					grid.Clear()
					for index, box := range boxes {
						grid.Add(index, box)
					}
					grid.Build()
				}
			})
		}
	}
}

// BenchmarkGrid3RebuildCellSize sweeps the cell size at a fixed scene, because
// the rule of thumb a 2D grid runs on — cells about the size of the things in
// them — does not carry to three axes. Halving the cell size octuples the cell
// count, and the grid goes from partitioning the scene to mostly storing
// emptiness. The occupancy each size reports is the number to read alongside
// the time.
func BenchmarkGrid3RebuildCellSize(b *testing.B) {
	for _, cellSize := range []float32{2, 4, 8, 16, 32} {
		b.Run("cell="+formatCellSize(cellSize), func(b *testing.B) {
			boxes := newScene(1, benchmarkItems, 200, 1).boxes
			grid := NewUniformGrid3[float32, int](cellSize)

			for index, box := range boxes {
				grid.Add(index, box)
			}
			grid.Build()

			x, y, z := grid.CellCounts()
			cells := x * y * z
			occupancy := float64(grid.OccupiedCells()) / float64(cells)

			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				grid.Clear()
				for index, box := range boxes {
					grid.Add(index, box)
				}
				grid.Build()
			}

			b.ReportMetric(float64(cells), "cells")
			b.ReportMetric(occupancy, "occupancy")
		})
	}
}

func formatCellSize(cellSize float32) string {
	return strconv.FormatFloat(float64(cellSize), 'g', -1, 32)
}

// BenchmarkGrid3RayNearestHit is the query the ordering exists for: cast a ray,
// narrow-phase each candidate, and stop as soon as the confirmed hit beats the
// cell being offered. It is measured against the linear scan it replaces.
func BenchmarkGrid3RayNearestHit(b *testing.B) {
	for _, world := range benchmarkScenes {
		boxes := newScene(1, benchmarkItems, 200, world.spread).boxes
		rays := benchmarkRays(boxes, 256)

		b.Run(world.name+"/linearScan", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for step := range b.N {
				ray := rays[step%len(rays)]

				nearest := float32(math.Inf(1))
				for _, box := range boxes {
					if distance, hit := collider.Ray3Box3(ray, box); hit && distance < nearest {
						nearest = distance
					}
				}
			}
		})

		for _, shape := range benchmarkGrids {
			b.Run(world.name+"/"+shape.name, func(b *testing.B) {
				grid := shape.make()
				for index, box := range boxes {
					grid.Add(index, box)
				}
				grid.Build()

				b.ReportAllocs()
				b.ResetTimer()

				for step := range b.N {
					ray := rays[step%len(rays)]

					nearest := float32(math.Inf(1))
					found := false
					for cellEntry, id := range grid.RayCandidates(ray) {
						if found && nearest < cellEntry {
							break
						}

						if distance, hit := collider.Ray3Box3(ray, boxes[id]); hit && distance < nearest {
							nearest, found = distance, true
						}
					}
				}
			})
		}
	}
}

// BenchmarkGrid3RayCandidateCount reports how many candidates each grid shape
// hands the narrow phase per cast, which is the number the height decision
// actually turns on — the traversal's own cost is the same either way.
func BenchmarkGrid3RayCandidateCount(b *testing.B) {
	for _, world := range benchmarkScenes {
		boxes := newScene(1, benchmarkItems, 200, world.spread).boxes
		rays := benchmarkRays(boxes, 256)

		for _, shape := range benchmarkGrids {
			b.Run(world.name+"/"+shape.name, func(b *testing.B) {
				grid := shape.make()
				for index, box := range boxes {
					grid.Add(index, box)
				}
				grid.Build()

				tested := 0
				b.ResetTimer()

				for step := range b.N {
					ray := rays[step%len(rays)]

					nearest := float32(math.Inf(1))
					found := false
					for cellEntry, id := range grid.RayCandidates(ray) {
						if found && nearest < cellEntry {
							break
						}

						tested++
						if distance, hit := collider.Ray3Box3(ray, boxes[id]); hit && distance < nearest {
							nearest, found = distance, true
						}
					}
				}

				b.ReportMetric(float64(tested)/float64(b.N), "candidates/cast")
			})
		}
	}
}

func BenchmarkGrid3QueryBox(b *testing.B) {
	for _, world := range benchmarkScenes {
		for _, shape := range benchmarkGrids {
			b.Run(world.name+"/"+shape.name, func(b *testing.B) {
				boxes := newScene(1, benchmarkItems, 200, world.spread).boxes
				grid := shape.make()
				for index, box := range boxes {
					grid.Add(index, box)
				}
				grid.Build()

				query := gomath.NewBox3(
					gomath.Vector3[float32]{},
					gomath.Vector3[float32]{X: 10, Y: 10, Z: 10})

				b.ReportAllocs()
				b.ResetTimer()

				found := 0
				for range b.N {
					for range grid.QueryBox(query) {
						found++
					}
				}
				b.ReportMetric(float64(found)/float64(b.N), "items/query")
			})
		}
	}
}

func BenchmarkGrid3Pairs(b *testing.B) {
	for _, world := range benchmarkScenes {
		for _, shape := range benchmarkGrids {
			b.Run(world.name+"/"+shape.name, func(b *testing.B) {
				boxes := newScene(1, benchmarkItems, 200, world.spread).boxes
				grid := shape.make()
				for index, box := range boxes {
					grid.Add(index, box)
				}
				grid.Build()

				b.ReportAllocs()
				b.ResetTimer()

				pairs := 0
				for range b.N {
					for range grid.Pairs() {
						pairs++
					}
				}
				b.ReportMetric(float64(pairs)/float64(b.N), "pairs/sweep")
			})
		}
	}
}

// benchmarkRays are casts a camera would make: from outside and above the
// scene, aimed at something in it, so the traversal has a hit to exit early on.
func benchmarkRays(boxes []gomath.Box3[float32], count int) []gomath.Ray3[float32] {
	random := rand.New(rand.NewSource(99))

	rays := make([]gomath.Ray3[float32], count)
	for index := range rays {
		origin := gomath.Vector3[float32]{
			X: (random.Float32() - 0.5) * 400,
			Y: 150,
			Z: (random.Float32() - 0.5) * 400,
		}

		rays[index] = gomath.NewRay3Between(origin, boxes[random.Intn(len(boxes))].Center)
	}

	return rays
}
