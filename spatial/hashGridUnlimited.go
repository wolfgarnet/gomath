package spatial

import "github.com/wolfgarnet/gomath"

type GridCellIndex [2]int64

type UnlimitedHashGrid[T gomath.Number, I BoundedItem[ID], ID comparable] struct {
	inverseCellSize T
}

func (grid *UnlimitedHashGrid[T, I, ID]) HashPosition(x T, y T) GridCellIndex {
	xIndex := int64(x * grid.inverseCellSize)
	yIndex := int64(y * grid.inverseCellSize)
	return GridCellIndex{xIndex, yIndex}
}

func (grid *UnlimitedHashGrid[T, I, ID]) CalculateNeighbourHashIndexes(index GridCellIndex) []GridCellIndex {
	neighbourIndexes := make([]GridCellIndex, 0, 8)
	x := index[0]
	y := index[1]

	// Top
	neighbourIndexes = append(neighbourIndexes, GridCellIndex{x, y - 1})

	// Top-right
	neighbourIndexes = append(neighbourIndexes, GridCellIndex{x + 1, y - 1})

	// Right
	neighbourIndexes = append(neighbourIndexes, GridCellIndex{x + 1, y})

	// Bottom-right
	neighbourIndexes = append(neighbourIndexes, GridCellIndex{x + 1, y + 1})

	// Bottom
	neighbourIndexes = append(neighbourIndexes, GridCellIndex{x, y + 1})

	// Bottom-left
	neighbourIndexes = append(neighbourIndexes, GridCellIndex{x - 1, y + 1})

	// Left
	neighbourIndexes = append(neighbourIndexes, GridCellIndex{x - 1, y})

	// Top-left
	neighbourIndexes = append(neighbourIndexes, GridCellIndex{x - 1, y - 1})

	return neighbourIndexes
}
