package spatial

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/wolfgarnet/gomath"
)

// BenchmarkSimpleHashGridRebuild measures the per-tick life of a rebuilt
// broad phase, the way pixelmaster's collision system drives it: clear the grid,
// insert every body, then walk the cells pairing each item with its own cell and
// the four "next" neighbours. The pair test is a plain distance check so the
// figure is dominated by the grid rather than a narrow phase.
//
// The zero-item case is the cost a grid charges for existing at all — the
// projectile grid on a tick with no projectiles in flight.
func BenchmarkSimpleHashGridRebuild(b *testing.B) {
	for _, items := range []int{0, 100, 1000, 5000} {
		b.Run(fmt.Sprintf("items=%d", items), func(b *testing.B) {
			grid := NewSimpleHashGrid[float32, uint64](40, 40, 24)
			random := rand.New(rand.NewSource(1))
			extent := float32(grid.Width()) * grid.CellSize()
			positions := make([]gomath.Vector2[float32], items)
			for index := range positions {
				positions[index] = gomath.Vector2[float32]{X: random.Float32() * extent, Y: random.Float32() * extent}
			}

			pairs := 0
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				grid.Clear()
				for index, position := range positions {
					grid.AddItem(uint64(index), position)
				}
				for _, cell := range grid.IterateCells() {
					if cell.Size() == 0 {
						continue
					}
					for index, item := range cell.IterateContent() {
						for other := range cell.IterateContentFrom(index + 1) {
							if positions[item].DistanceToSquared(positions[other]) < 64 {
								pairs++
							}
						}
						for neighbour := range cell.IterateNextNeighbors() {
							for _, other := range neighbour.IterateContent() {
								if positions[item].DistanceToSquared(positions[other]) < 64 {
									pairs++
								}
							}
						}
					}
				}
			}
			_ = pairs
		})
	}
}
