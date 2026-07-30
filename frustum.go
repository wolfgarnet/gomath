package gomath

import "fmt"

// Frustum is the six-plane view volume a camera renders, with every plane's
// normal pointing INWARD.
//
// That inward convention is the whole point: a shape is visible if and only if
// no plane reports it entirely behind, so every test is a run of six comparisons
// with an early-out and no special cases.
type Frustum[T Float] struct {
	Planes [6]Plane3[T]
}

// Plane indices into Frustum.Planes.
const (
	FrustumLeft = iota
	FrustumRight
	FrustumBottom
	FrustumTop
	FrustumNear
	FrustumFar
)

// NewFrustumFromMatrix extracts the six planes from a view-projection matrix.
//
// # How this works
//
// The matrix takes a world point to clip space, where OpenGL's view volume is
// simply -w <= x,y,z <= w. Each of those six inequalities, rearranged, is a
// plane equation in world space made from a sum or difference of two ROWS of the
// matrix. The left plane, for instance, is x >= -w, which is
// (row0 + row3) . p >= 0.
//
// So no trigonometry, no camera parameters and no inverse are needed — and it
// works for an orthographic matrix as much as a perspective one, since both map
// their volume onto the same cube. This is the Gribb-Hartmann extraction.
//
// The planes are normalised, so signed distances come back in world units and
// can be compared against object radii.
//
// Note that Matrix4 is column-major with element (row r, col c) at c*4+r, so
// row r is m[r], m[4+r], m[8+r], m[12+r] — reading columns here instead is the
// easiest way to get a frustum that looks plausible and culls the wrong things.
func NewFrustumFromMatrix[T Float](viewProjection Matrix4[T]) Frustum[T] {
	row := func(r int) Vector4[T] {
		return Vector4[T]{
			X: viewProjection[r],
			Y: viewProjection[4+r],
			Z: viewProjection[8+r],
			W: viewProjection[12+r],
		}
	}

	x, y, z, w := row(0), row(1), row(2), row(3)

	fromVector4 := func(v Vector4[T]) Plane3[T] {
		return Plane3[T]{
			Normal:   Vector3[T]{X: v.X, Y: v.Y, Z: v.Z},
			Distance: v.W,
		}.Normalized()
	}

	var frustum Frustum[T]

	frustum.Planes[FrustumLeft] = fromVector4(w.Add(x))   // x >= -w
	frustum.Planes[FrustumRight] = fromVector4(w.Sub(x))  // x <=  w
	frustum.Planes[FrustumBottom] = fromVector4(w.Add(y)) // y >= -w
	frustum.Planes[FrustumTop] = fromVector4(w.Sub(y))    // y <=  w
	frustum.Planes[FrustumNear] = fromVector4(w.Add(z))   // z >= -w
	frustum.Planes[FrustumFar] = fromVector4(w.Sub(z))    // z <=  w

	return frustum
}

// ContainsPoint reports whether the point is inside the view volume.
func (f Frustum[T]) ContainsPoint(point Vector3[T]) bool {
	for _, plane := range f.Planes {
		if plane.SignedDistance(point) < 0 {
			return false
		}
	}

	return true
}

// IntersectsBox3 reports whether an axis-aligned box is at least partly inside.
//
// It is CONSERVATIVE: a box just outside a frustum corner can pass every
// individual plane test and be reported visible. That is the correct bias for
// culling, where a false positive costs one wasted draw and a false negative
// costs a missing object.
func (f Frustum[T]) IntersectsBox3(box Box3[T]) bool {
	for _, plane := range f.Planes {
		// The box's extent along the plane normal: the support radius.
		radius := Abs(plane.Normal.X)*box.HalfSize.X +
			Abs(plane.Normal.Y)*box.HalfSize.Y +
			Abs(plane.Normal.Z)*box.HalfSize.Z

		if plane.SignedDistance(box.Center)+radius < 0 {
			// Entirely behind this plane, so entirely outside.
			return false
		}
	}

	return true
}

// ContainsBox3 reports whether the box lies wholly inside every plane.
//
// This is what makes hierarchical culling worthwhile: a node fully inside needs
// no further tests for any of its children.
func (f Frustum[T]) ContainsBox3(box Box3[T]) bool {
	for _, plane := range f.Planes {
		radius := Abs(plane.Normal.X)*box.HalfSize.X +
			Abs(plane.Normal.Y)*box.HalfSize.Y +
			Abs(plane.Normal.Z)*box.HalfSize.Z

		if plane.SignedDistance(box.Center)-radius < 0 {
			return false
		}
	}

	return true
}

// IntersectsOBB3 reports whether an oriented box is at least partly inside.
//
// It is tighter than testing the box's axis-aligned bounds, because the support
// radius is measured along the box's own axes rather than the world's.
func (f Frustum[T]) IntersectsOBB3(box OBB3[T]) bool {
	xAxis, yAxis, zAxis := box.Axis(0), box.Axis(1), box.Axis(2)

	for _, plane := range f.Planes {
		radius := Abs(plane.Normal.Dot(xAxis))*box.HalfSize.X +
			Abs(plane.Normal.Dot(yAxis))*box.HalfSize.Y +
			Abs(plane.Normal.Dot(zAxis))*box.HalfSize.Z

		if plane.SignedDistance(box.Center)+radius < 0 {
			return false
		}
	}

	return true
}

// IntersectsSphere reports whether a sphere is at least partly inside.
//
// This is the cheapest of the volume tests — one dot product and an add per
// plane, with no support-radius sum — which is why a cull chain that has a
// bounding sphere should try it before anything box-shaped.
func (f Frustum[T]) IntersectsSphere(sphere Sphere[T]) bool {
	for _, plane := range f.Planes {
		if plane.SignedDistance(sphere.Center)+sphere.Radius < 0 {
			return false
		}
	}

	return true
}

// ContainsSphere reports whether the sphere lies wholly inside every plane, which
// lets a hierarchy skip testing anything beneath it.
func (f Frustum[T]) ContainsSphere(sphere Sphere[T]) bool {
	for _, plane := range f.Planes {
		if plane.SignedDistance(sphere.Center)-sphere.Radius < 0 {
			return false
		}
	}

	return true
}

func (f Frustum[T]) String() string {
	return fmt.Sprintf("Frustum(near %v, far %v)", f.Planes[FrustumNear], f.Planes[FrustumFar])
}
