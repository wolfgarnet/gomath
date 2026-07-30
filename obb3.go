package gomath

import "fmt"

// OBB3 is an oriented bounding box in 3D: a centre, half-extents along its own
// axes, and a rotation giving those axes in world space.
//
// # Why the rotation is a Matrix3 and not a Quaternion
//
// The separating-axis test needs the three world-space axes on every call. A
// quaternion would have to be converted to a basis each time — fifteen axis
// tests per pair, many pairs per frame. Storing the basis converts once, at
// construction. Build one from a quaternion with NewOBB3FromQuaternion when
// that is what the caller holds.
//
// The rotation is assumed orthonormal. A basis with scale in it makes the
// half-extents mean something different from what every test assumes, so scale
// belongs in HalfSize.
type OBB3[T Float] struct {
	Center   Vector3[T]
	HalfSize Vector3[T]
	// Rotation holds the box's axes as its columns.
	Rotation Matrix3[T]
}

// NewOBB3 returns an oriented box from a centre, half-extents and an orthonormal
// basis whose columns are the axes.
func NewOBB3[T Float](center Vector3[T], halfSize Vector3[T], rotation Matrix3[T]) OBB3[T] {
	return OBB3[T]{Center: center, HalfSize: halfSize, Rotation: rotation}
}

// NewOBB3FromQuaternion returns an oriented box whose axes come from a rotation
// quaternion.
func NewOBB3FromQuaternion[T Float](center Vector3[T], halfSize Vector3[T], rotation Quaternion[T]) OBB3[T] {
	return OBB3[T]{Center: center, HalfSize: halfSize, Rotation: rotation.ToMatrix3()}
}

// NewOBB3FromBox3 returns an axis-aligned box as an oriented one, with the
// identity basis.
func NewOBB3FromBox3[T Float](box Box3[T]) OBB3[T] {
	return OBB3[T]{Center: box.Center, HalfSize: box.HalfSize, Rotation: Identity3[T]()}
}

// NewOBB3FromMatrix4 returns an oriented box from a local-space Box3 and a
// transform.
//
// The transform's basis columns are normalised to give the axes, and their
// lengths are folded into the half-extents — so a matrix carrying scale
// produces the box that scale implies, rather than a basis that quietly
// violates the orthonormality every test assumes.
//
// It reports false when an axis has been scaled to zero, since the box then has
// no valid orientation.
func NewOBB3FromMatrix4[T Float](box Box3[T], transform Matrix4[T]) (OBB3[T], bool) {
	basis := transform.Basis()

	xAxis := basis.Column(0)
	yAxis := basis.Column(1)
	zAxis := basis.Column(2)

	scale := Vector3[T]{X: xAxis.Length(), Y: yAxis.Length(), Z: zAxis.Length()}
	if scale.X == 0 || scale.Y == 0 || scale.Z == 0 {
		return OBB3[T]{}, false
	}

	rotation := Basis3(
		xAxis.DivScalar(scale.X),
		yAxis.DivScalar(scale.Y),
		zAxis.DivScalar(scale.Z),
	)

	return OBB3[T]{
		Center:   transform.TransformPoint(box.Center),
		HalfSize: box.HalfSize.Mul(scale),
		Rotation: rotation,
	}, true
}

// Axis returns the world-space axis for the given local axis index, 0 for X,
// 1 for Y and 2 for Z.
func (o OBB3[T]) Axis(index int) Vector3[T] {
	return o.Rotation.Column(index)
}

// ToLocal expresses a world point in the box's own frame, where the box is the
// axis-aligned span [-HalfSize, +HalfSize].
func (o OBB3[T]) ToLocal(point Vector3[T]) Vector3[T] {
	offset := point.Sub(o.Center)

	return Vector3[T]{
		X: offset.Dot(o.Axis(0)),
		Y: offset.Dot(o.Axis(1)),
		Z: offset.Dot(o.Axis(2)),
	}
}

// ToWorld maps a point from the box's own frame back to world space.
func (o OBB3[T]) ToWorld(local Vector3[T]) Vector3[T] {
	return o.Center.
		Add(o.Axis(0).MultiplyScalar(local.X)).
		Add(o.Axis(1).MultiplyScalar(local.Y)).
		Add(o.Axis(2).MultiplyScalar(local.Z))
}

// ContainsPoint reports whether the point is inside or on the boundary.
//
// The boundary is inclusive, but note that a point derived from this box's own
// geometry — a corner, or a point returned by ClosestPoint — sits exactly on the
// boundary in exact arithmetic and may round just outside it. Do not rely on
// this returning true for such points; compare with a tolerance instead.
func (o OBB3[T]) ContainsPoint(point Vector3[T]) bool {
	local := o.ToLocal(point)

	return Abs(local.X) <= o.HalfSize.X &&
		Abs(local.Y) <= o.HalfSize.Y &&
		Abs(local.Z) <= o.HalfSize.Z
}

// ClosestPoint returns the point of the box nearest the given point, which is
// the point itself when it is inside.
func (o OBB3[T]) ClosestPoint(point Vector3[T]) Vector3[T] {
	local := o.ToLocal(point)
	clamped := local.Clamp(o.HalfSize.Negate(), o.HalfSize)

	return o.ToWorld(clamped)
}

// Corners returns the eight corners in world space, ordered so that bit 0 of the
// index selects the +X face, bit 1 +Y and bit 2 +Z — the same order as
// Box3.Corners.
func (o OBB3[T]) Corners() [8]Vector3[T] {
	var corners [8]Vector3[T]

	for index := range corners {
		local := Vector3[T]{
			X: -o.HalfSize.X,
			Y: -o.HalfSize.Y,
			Z: -o.HalfSize.Z,
		}
		if index&1 != 0 {
			local.X = o.HalfSize.X
		}
		if index&2 != 0 {
			local.Y = o.HalfSize.Y
		}
		if index&4 != 0 {
			local.Z = o.HalfSize.Z
		}

		corners[index] = o.ToWorld(local)
	}

	return corners
}

// BoundingBox returns the axis-aligned box enclosing the oriented one.
//
// The half-extents are the oriented box's extents projected onto each world
// axis, which is the absolute value of the basis applied to the half-size.
func (o OBB3[T]) BoundingBox() Box3[T] {
	xAxis := o.Axis(0).Abs().MultiplyScalar(o.HalfSize.X)
	yAxis := o.Axis(1).Abs().MultiplyScalar(o.HalfSize.Y)
	zAxis := o.Axis(2).Abs().MultiplyScalar(o.HalfSize.Z)

	return Box3[T]{
		Center:   o.Center,
		HalfSize: xAxis.Add(yAxis).Add(zAxis),
	}
}

// Volume returns the enclosed volume, which rotation does not change.
func (o OBB3[T]) Volume() T {
	return 8 * o.HalfSize.X * o.HalfSize.Y * o.HalfSize.Z
}

// Translate returns the box moved by the given offset.
func (o OBB3[T]) Translate(translation Vector3[T]) OBB3[T] {
	o.Center = o.Center.Add(translation)

	return o
}

// Rotated returns the box with an additional rotation applied about its centre.
func (o OBB3[T]) Rotated(rotation Matrix3[T]) OBB3[T] {
	o.Rotation = rotation.Mul(o.Rotation)

	return o
}

func (o OBB3[T]) String() string {
	return fmt.Sprintf("OBB3(center %v, half %v, x %v)", o.Center, o.HalfSize, o.Axis(0))
}
