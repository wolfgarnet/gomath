package spatial

import "github.com/wolfgarnet/gomath"

type GridSpecification[T gomath.Number] struct {
	// width is the width of the grid, number of cells in x direction
	width uint64

	// height is the height of the grid, number of cells in y direction
	height uint64

	// cellSize is the size of a cell in the grid
	cellSize         T
	offsetX, offsetY T

	// xHashMask and yHashMask are used to mask the x and y hash values to the width and height of the grid
	xHashMask, yHashMask uint64

	// inverseCellSize is the inverse of cellSize for easier computations
	inverseCellSize T
}

func NewGridSpecification[T gomath.Number](width, height uint64, cellSize T, offsetX, offsetY T) GridSpecification[T] {
	width = gomath.NextPowerOfTwo(width)
	height = gomath.NextPowerOfTwo(height)

	return GridSpecification[T]{
		width:           width,
		height:          height,
		cellSize:        cellSize,
		offsetX:         offsetX,
		offsetY:         offsetY,
		xHashMask:       width - 1,
		yHashMask:       height - 1,
		inverseCellSize: 1 / cellSize,
	}
}

func (spec GridSpecification[T]) HashPositionToroidalIndex(worldPositionX T, worldPositionY T) uint64 {
	var yHash, xHash uint64
	if worldPositionX < 0 {
		xIndex := -worldPositionX * spec.inverseCellSize
		xHash = spec.width - 1 - uint64(xIndex)&spec.xHashMask
	} else {
		xIndex := worldPositionX * spec.inverseCellSize
		xHash = uint64(xIndex) & spec.xHashMask
	}

	worldPositionX += spec.offsetX
	worldPositionY += spec.offsetY

	if worldPositionY < 0 {
		yIndex := -worldPositionY * spec.inverseCellSize
		masked := uint64(yIndex) & spec.yHashMask
		yHash = spec.yHashMask - masked
	} else {
		yIndex := worldPositionY * spec.inverseCellSize
		yHash = uint64(yIndex) & spec.yHashMask
	}

	hash := xHash + yHash*spec.width
	return hash
}

func (spec GridSpecification[T]) HashPositionIndexCapped(worldPositionX T, worldPositionY T) uint64 {
	worldPositionX += spec.offsetX
	worldPositionY += spec.offsetY

	var xHash, yHash uint64
	if worldPositionX < 0 {

	} else {
		xIndex := worldPositionX * spec.inverseCellSize
		xHash = uint64(xIndex) & spec.xHashMask
		if xHash > spec.xHashMask {
			xHash = spec.width - 1
		}
	}
	if worldPositionY < 0 {

	} else {
		yIndex := worldPositionY * spec.inverseCellSize
		yHash = uint64(yIndex) & spec.yHashMask
		if yHash > spec.yHashMask {
			yHash = spec.height - 1
		}
	}

	hash := xHash + yHash*spec.width
	return hash
}

func (spec GridSpecification[T]) HashPosition(worldPositionX T, worldPositionY T) (uint64, uint64) {
	worldPositionX += spec.offsetX
	worldPositionY += spec.offsetY

	xIndex := worldPositionX * spec.inverseCellSize
	xHash := uint64(xIndex) & spec.xHashMask
	yIndex := worldPositionY * spec.inverseCellSize
	yHash := uint64(yIndex) & spec.yHashMask

	return xHash, yHash
}

func (spec GridSpecification[T]) HashValuesForSquared(position gomath.Vector2[T], radius T) (startHash uint64, offsetX uint64, offsetY uint64) {
	topX := position.X - radius
	topY := position.Y - radius
	bottomX := position.X + radius
	bottomY := position.Y + radius

	startX, startY := spec.HashPosition(topX, topY)
	endX, endY := spec.HashPosition(bottomX, bottomY)
	/*endX &= grid.width
	endY &= grid.height*/

	offsetX = endX - startX + 1
	offsetY = endY - startY + 1

	startHash = startX + startY*spec.width

	return startHash, offsetX, offsetY
}
