package spatial

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*func TestHashGridCalculateNeighbourHashIndexes(t *testing.T) {
	grid := NewHashGrid[float64, TestItem2D, uint64](8, 8, 10.0)
	tests := map[string]struct {
		x float64
		y float64
	}{
		"0,0": {
			x: 0,
			y: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			hash := grid.HashPositionIndex(test.x, test.y)
			n := grid.CalculateNeighbourHashIndexes(hash)
			for _, cellHash := range n {
				x, y := grid.HashToCell(cellHash)
				spew.Dump(x, y)
			}
		})
	}
}*/

func TestHashPosition(t *testing.T) {
	grid := NewHashGrid[float64, TestItem2D, uint64](8, 8, 10.0)

	tests := map[string]struct {
		x, y     float64
		offset   float64
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
		"negative_x - wrapped": {
			x:        -5,
			y:        0,
			ix:       7,
			iy:       0,
			expected: 7,
		},
		"origin - offset": {
			x:        0,
			y:        0,
			offset:   40,
			ix:       4,
			iy:       4,
			expected: 20,
		},
		"upper left - offset": {
			x:        -20,
			y:        -20,
			offset:   40,
			ix:       2,
			iy:       2,
			expected: 21,
		},
		"lower right - offset": {
			x:        20,
			y:        20,
			offset:   40,
			ix:       6,
			iy:       6,
			expected: 50,
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
			x:        math.MaxFloat64,
			y:        0,
			ix:       0,
			iy:       0,
			expected: 0,
		}, // Extreme case
		"min_x": {
			x:        -math.MaxFloat64,
			y:        0,
			ix:       7,
			iy:       0,
			expected: 7,
		}, // Extreme case
		"max_y": {
			x:        0,
			y:        math.MaxFloat64,
			ix:       0,
			iy:       0,
			expected: 0,
		}, // Extreme case
		"min_y": {
			x:        0,
			y:        -math.MaxFloat64,
			ix:       0,
			iy:       7,
			expected: 56,
		}, // Extreme case
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			grid.offsetX = test.offset
			grid.offsetY = test.offset

			result := grid.HashPositionIndex(test.x, test.y)
			assert.Equal(t, test.expected, result, "Hash")
			ix, iy := grid.HashPosition(test.x, test.y)
			assert.Equal(t, test.ix, ix, "Grid x")
			assert.Equal(t, test.iy, iy, "Grid y")
		})
	}
}

func TestHashGrid_HashToCell(t *testing.T) {
	grid := NewHashGrid[float64, TestItem2D, uint64](8, 8, 10.0)
	tests := map[string]struct {
		x, y                         float64
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
