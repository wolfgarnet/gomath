package spatial

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wolfgarnet/gomath"
)

func TestHashGridCalculateNeighbourHashIndexes(t *testing.T) {
	grid := NewHashGrid[float32, *testItem, uint64](8, 8, 10.0)
	tests := map[string]struct {
		x float32
		y float32
	}{
		"0,0": {
			x: 10,
			y: 10,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			hash := grid.HashPositionIndex(test.x, test.y)
			n := grid.NeighbourHashes(hash)
			for _, cellHash := range n {
				assert.Contains(t, grid.grid[hash].neighbourHashes, cellHash, "Hash %d not in grid", cellHash)
				//neighbour := grid.grid[cellHash]
				//assert.Contains(t, grid.grid[hash].neighbours, cellHash, "Hash %d not in grid", cellHash)
			}
		})
	}
}

func TestHashValuesForSquared(t *testing.T) {
	grid := NewHashGrid[float32, *testItem, uint64](64, 64, 16.0)

	tests := map[string]struct {
		position         gomath.Vector2[float32]
		radius           float32
		hash             uint64
		offsetX, offsetY uint64
	}{
		"very negative (-529, -160)": {
			position: gomath.Vector2[float32]{X: -529, Y: -160},
			radius:   128,
			hash:     2902,
			offsetX:  17,
			offsetY:  17,
		},
		"-32, 596": {
			position: gomath.Vector2[float32]{X: -32, Y: 596},
			radius:   128,
			hash:     1909,
			offsetX:  17,
			offsetY:  17,
		},
		"origin": {
			position: gomath.Vector2[float32]{X: 128, Y: 128},
			radius:   128,
			hash:     0,
			offsetX:  17,
			offsetY:  17,
		},
		"origin++": {
			position: gomath.Vector2[float32]{X: 130, Y: 130},
			radius:   128,
			hash:     0,
			offsetX:  17,
			offsetY:  17,
		},
		"origin almost next": {
			position: gomath.Vector2[float32]{X: 143, Y: 143},
			radius:   128,
			hash:     0,
			offsetX:  17,
			offsetY:  17,
		},
		"origin next": {
			position: gomath.Vector2[float32]{X: 144, Y: 144},
			radius:   128,
			hash:     65,
			offsetX:  17,
			offsetY:  17,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			hash, offsetX, offsetY := grid.HashValuesForSquared(test.position, test.radius)
			assert.Equalf(t, test.hash, hash, "Hash: %d not %d", hash, test.hash)
			assert.Equalf(t, test.offsetX, offsetX, "OffsetX: %d not %d", offsetX, test.offsetX)
			assert.Equalf(t, test.offsetY, offsetY, "OffsetY: %d not %d", offsetY, test.offsetY)
		})
	}
}

func TestHashPosition(t *testing.T) {
	grid := NewHashGrid[float32, *testItem, uint64](8, 8, 10.0)

	tests := map[string]struct {
		x, y     float32
		ix, iy   uint64
		expected uint64
	}{
		"origin": {
			x:        0,
			y:        0,
			ix:       0,
			iy:       0,
			expected: 0,
		},
		"positive_x": {
			x:        15,
			y:        0,
			ix:       1,
			iy:       0,
			expected: 1,
		},
		"positive_x2": {
			x:        25,
			y:        0,
			ix:       2,
			iy:       0,
			expected: 2,
		},
		"positive_x - last": {
			x:        79,
			y:        0,
			ix:       7,
			iy:       0,
			expected: 7,
		},
		"positive_x - last 2": {
			x:        80,
			y:        0,
			ix:       0,
			iy:       0,
			expected: 0,
		},
		"negative_x - capped": {
			x:        -5,
			y:        0,
			ix:       0,
			iy:       0,
			expected: 0,
		},
		"origin - offset": {
			x:        0,
			y:        0,
			ix:       4,
			iy:       4,
			expected: 36,
		},
		"upper left - offset": {
			x:        -20,
			y:        -20,
			ix:       2,
			iy:       2,
			expected: 18,
		},
		"lower right - offset": {
			x:        20,
			y:        20,
			ix:       6,
			iy:       6,
			expected: 54,
		},
		"negative_x2": {
			x:        -15,
			y:        0,
			ix:       6,
			iy:       0,
			expected: 6,
		},
		"positive_y": {
			x:        0,
			y:        25,
			ix:       0,
			iy:       2,
			expected: 16,
		},
		"positive_y2": {
			x:        0,
			y:        15,
			ix:       0,
			iy:       1,
			expected: 8,
		},
		"negative_y": {
			x:        0,
			y:        -55,
			ix:       0,
			iy:       2,
			expected: 16,
		},
		"negative_y 2": {
			x:        0,
			y:        -25,
			ix:       0,
			iy:       5,
			expected: 40,
		},
		"positive_x_y": {
			x:        35,
			y:        45,
			ix:       3,
			iy:       4,
			expected: 35,
		},
		"negative_x_y": {
			x:        -35,
			y:        -45,
			ix:       4,
			iy:       3,
			expected: 28,
		},
		"max_x": {
			x:        math.MaxFloat32,
			y:        0,
			ix:       0,
			iy:       0,
			expected: 0,
		}, // Extreme case
		"min_x": {
			x:        -math.MaxFloat32,
			y:        0,
			ix:       7,
			iy:       0,
			expected: 7,
		}, // Extreme case
		"max_y": {
			x:        0,
			y:        math.MaxFloat32,
			ix:       0,
			iy:       0,
			expected: 0,
		}, // Extreme case
		"min_y": {
			x:        0,
			y:        -math.MaxFloat32,
			ix:       0,
			iy:       7,
			expected: 56,
		}, // Extreme case
		"1,1": {
			x:        9,
			y:        9,
			ix:       0,
			iy:       0,
			expected: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			result := grid.HashPositionIndex(test.x, test.y)
			assert.Equalf(t, test.expected, result, "Hash %d not %d", result, test.expected)
			ix, iy := grid.HashPosition(test.x, test.y)
			assert.Equalf(t, test.ix, ix, "Grid x %d not %d", ix, test.ix)
			assert.Equalf(t, test.iy, iy, "Grid y %d not %d", iy, test.iy)
		})
	}
}

func TestHashGrid_HashToCell(t *testing.T) {
	grid := NewHashGrid[float32, *testItem, uint64](8, 8, 10.0)
	tests := map[string]struct {
		x, y                         float32
		expectedHash                 uint64
		expectedCellX, expectedCellY uint64
		expectedNeighboursPair       [][2]uint64
	}{
		"origin": {
			x:             0,
			y:             0,
			expectedHash:  0,
			expectedCellX: 0,
			expectedCellY: 0,
			expectedNeighboursPair: [][2]uint64{
				{0, 1}, {1, 0}, {1, 1}, {1, 2}, {2, 1},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			hash := grid.HashPositionIndex(test.x, test.y)
			assert.Equal(t, test.expectedHash, hash)
			x, y := grid.HashToCell(hash)
			assert.Equal(t, test.expectedCellX, x)
			assert.Equal(t, test.expectedCellY, y)
			hashes := grid.NeighbourHashesPair(hash)
			require.Lenf(t, hashes, len(test.expectedNeighboursPair), "Hashes length mismatch: %v", hashes)
			for i := range hashes {
				assert.Contains(t, hashes[i], test.expectedNeighboursPair[i])
			}
		})
	}
}

func TestHashGrid(t *testing.T) {
	grid := NewHashGrid[float32, *testItem, uint64](64, 64, 16.0)
	tests := map[string]struct {
		x, y          float32
		expectedHash  uint64
		x2, y2        float32
		expectedHash2 uint64
	}{
		"real": {
			x:             106,
			y:             175,
			expectedHash:  646,
			x2:            100,
			y2:            185,
			expectedHash2: 710,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			item := &testItem{
				IDValue:       uint64(1),
				PositionValue: gomath.Vector2[float32]{X: test.x, Y: test.y},
				radius:        8.0,
			}
			item2 := &testItem{
				IDValue:       uint64(2),
				PositionValue: gomath.Vector2[float32]{X: test.x2, Y: test.y2},
				radius:        8.0,
			}
			hash := grid.AddItem(item)
			assert.Equalf(t, test.expectedHash, hash, "Hash: %d not %d", hash, test.expectedHash)
			hash2 := grid.AddItem(item2)
			assert.Equalf(t, test.expectedHash2, hash2, "Hash: %d not %d", hash2, test.expectedHash2)
			cell := grid.grid[hash]
			assert.Contains(t, cell.neighbourHashes, hash2, "Hash %d not in grid", hash2)

			hash = grid.localToHash(65, 65)
			assert.Equalf(t, hash, 65, "Hash: %d not %d", hash, 65)
		})
	}
}
