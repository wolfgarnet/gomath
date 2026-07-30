package gomath

import "fmt"

// Box3 is an axis-aligned bounding box in 3D, stored as a centre and
// half-extents. It spans [Center-HalfSize, Center+HalfSize] on each axis.
//
// It is the 3D counterpart of AABB, and stores centre-plus-half-extents for the
// same reason AABB does: overlap and containment tests reduce to comparisons of
// differences, with no min/max reconstruction per test.
//
// Overlap between two Box3, or between a Box3 and any other shape, lives in the
// collider package — the root types carry only self-contained helpers.
type Box3[T Number] struct {
	Center   Vector3[T]
	HalfSize Vector3[T]
}

// NewBox3 returns a box with the given centre and half-extents.
func NewBox3[T Number](center Vector3[T], halfSize Vector3[T]) Box3[T] {
	return Box3[T]{Center: center, HalfSize: halfSize}
}

// NewBox3FromMinMax returns the box spanning two opposite corners. The
// arguments are sorted per axis, so they need not already be ordered.
func NewBox3FromMinMax[T Number](minimum Vector3[T], maximum Vector3[T]) Box3[T] {
	low := minimum.Min(maximum)
	high := minimum.Max(maximum)

	half := high.Sub(low).DivScalar(2)

	return Box3[T]{Center: low.Add(half), HalfSize: half}
}

// NewBox3FromPoints returns the smallest box containing every point, and false
// if there are none.
func NewBox3FromPoints[T Number](points []Vector3[T]) (Box3[T], bool) {
	if len(points) == 0 {
		return Box3[T]{}, false
	}

	low, high := points[0], points[0]
	for _, point := range points[1:] {
		low = low.Min(point)
		high = high.Max(point)
	}

	return NewBox3FromMinMax(low, high), true
}

// Min returns the corner with the lowest coordinate on every axis.
func (b Box3[T]) Min() Vector3[T] {
	return b.Center.Sub(b.HalfSize)
}

// Max returns the corner with the highest coordinate on every axis.
func (b Box3[T]) Max() Vector3[T] {
	return b.Center.Add(b.HalfSize)
}

// Size returns the full extent along each axis, which is twice the half-size.
func (b Box3[T]) Size() Vector3[T] {
	return b.HalfSize.MultiplyScalar(2)
}

// Volume returns the enclosed volume.
func (b Box3[T]) Volume() T {
	size := b.Size()

	return size.X * size.Y * size.Z
}

// SurfaceArea returns the total area of the six faces. It is what a bounding
// volume hierarchy's split heuristic costs its candidates by.
func (b Box3[T]) SurfaceArea() T {
	size := b.Size()

	return 2 * (size.X*size.Y + size.Y*size.Z + size.Z*size.X)
}

// IsEmpty reports whether the box has no extent on some axis, so it encloses no
// volume.
func (b Box3[T]) IsEmpty() bool {
	return b.HalfSize.X <= 0 || b.HalfSize.Y <= 0 || b.HalfSize.Z <= 0
}

// ContainsPoint reports whether the point is inside or on the boundary.
//
// The boundary is inclusive, but note that a point derived from this box's own
// geometry — a corner, or a point returned by ClosestPoint — sits exactly on the
// boundary in exact arithmetic and may round just outside it. Do not rely on
// this returning true for such points; compare with a tolerance instead.
func (b Box3[T]) ContainsPoint(point Vector3[T]) bool {
	offset := point.Sub(b.Center)

	return Abs(offset.X) <= b.HalfSize.X &&
		Abs(offset.Y) <= b.HalfSize.Y &&
		Abs(offset.Z) <= b.HalfSize.Z
}

// ContainsBox reports whether other lies entirely inside b.
func (b Box3[T]) ContainsBox(other Box3[T]) bool {
	offset := other.Center.Sub(b.Center)

	return Abs(offset.X)+other.HalfSize.X <= b.HalfSize.X &&
		Abs(offset.Y)+other.HalfSize.Y <= b.HalfSize.Y &&
		Abs(offset.Z)+other.HalfSize.Z <= b.HalfSize.Z
}

// ClosestPoint returns the point of the box nearest to the given point, which
// is the point itself when it is inside.
func (b Box3[T]) ClosestPoint(point Vector3[T]) Vector3[T] {
	return point.Clamp(b.Min(), b.Max())
}

// Union returns the smallest box containing both.
func (b Box3[T]) Union(other Box3[T]) Box3[T] {
	return NewBox3FromMinMax(
		b.Min().Min(other.Min()),
		b.Max().Max(other.Max()),
	)
}

// Expanded returns the box grown by amount on every axis, in every direction.
// A negative amount shrinks it, possibly to nothing.
func (b Box3[T]) Expanded(amount T) Box3[T] {
	b.HalfSize = b.HalfSize.AddScalar(amount)

	return b
}

// ExpandedToPoint returns the smallest box containing both b and the point.
func (b Box3[T]) ExpandedToPoint(point Vector3[T]) Box3[T] {
	return NewBox3FromMinMax(
		b.Min().Min(point),
		b.Max().Max(point),
	)
}

// Translate returns the box moved by the given offset.
func (b Box3[T]) Translate(translation Vector3[T]) Box3[T] {
	b.Center = b.Center.Add(translation)

	return b
}

// WithCenter returns the box moved to a new centre, keeping its extents.
func (b Box3[T]) WithCenter(center Vector3[T]) Box3[T] {
	b.Center = center

	return b
}

// Scaled returns the box with its half-extents multiplied, keeping its centre.
func (b Box3[T]) Scaled(scale T) Box3[T] {
	b.HalfSize = b.HalfSize.MultiplyScalar(scale)

	return b
}

// Corners returns the eight corners, ordered so that bit 0 of the index selects
// +X, bit 1 selects +Y and bit 2 selects +Z.
func (b Box3[T]) Corners() [8]Vector3[T] {
	var corners [8]Vector3[T]

	for index := range corners {
		offset := Vector3[T]{
			X: -b.HalfSize.X,
			Y: -b.HalfSize.Y,
			Z: -b.HalfSize.Z,
		}
		if index&1 != 0 {
			offset.X = b.HalfSize.X
		}
		if index&2 != 0 {
			offset.Y = b.HalfSize.Y
		}
		if index&4 != 0 {
			offset.Z = b.HalfSize.Z
		}

		corners[index] = b.Center.Add(offset)
	}

	return corners
}

// Equals reports exact equality of centre and half-extents.
func (b Box3[T]) Equals(other Box3[T]) bool {
	return b == other
}

// ApproxEquals reports whether centre and half-extents are within tol.
func (b Box3[T]) ApproxEquals(other Box3[T], tol T) bool {
	return Abs(b.Center.X-other.Center.X) <= tol &&
		Abs(b.Center.Y-other.Center.Y) <= tol &&
		Abs(b.Center.Z-other.Center.Z) <= tol &&
		Abs(b.HalfSize.X-other.HalfSize.X) <= tol &&
		Abs(b.HalfSize.Y-other.HalfSize.Y) <= tol &&
		Abs(b.HalfSize.Z-other.HalfSize.Z) <= tol
}

func (b Box3[T]) String() string {
	return fmt.Sprintf("Box3(center %v, half %v)", b.Center, b.HalfSize)
}
