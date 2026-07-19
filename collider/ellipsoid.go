package collider

import (
	"math"

	"github.com/wolfgarnet/gomath"
)

/*
https://www.geeksforgeeks.org/check-if-a-point-is-inside-outside-or-on-the-ellipse/
*/

// EllipsoidPoint reports whether point lies inside the ellipse e, boundary
// inclusive. The test is cross-multiplied so it is exact for integer T. The
// ellipse is placed at e.Position, so the point is given in world space (unlike
// the shape-first arg order matching CirclePoint/AABBPoint).
func EllipsoidPoint[T gomath.Number](e gomath.Ellipsoid[T], point gomath.Vector2[T]) bool {
	dx := point.X - e.Position.X
	dy := point.Y - e.Position.Y
	rx, ry := e.Radii.X, e.Radii.Y
	// (dx/rx)^2 + (dy/ry)^2 <= 1, cross-multiplied to avoid division.
	return dx*dx*ry*ry+dy*dy*rx*rx <= rx*rx*ry*ry
}

// EllipsoidLineSegment reports whether the segment touches or crosses the filled
// ellipse e (boundary inclusive).
//
// Scaling by (1/Radii.X, 1/Radii.Y) turns e into the unit disk at the origin and
// the segment into another segment; an invertible affine map preserves whether
// two regions intersect, so the question becomes whether the scaled segment comes
// within distance 1 of the origin.
func EllipsoidLineSegment[T gomath.Number](e gomath.Ellipsoid[T], s gomath.LineSegment[T]) bool {
	a, b := float64(e.Radii.X), float64(e.Radii.Y)
	if a <= 0 || b <= 0 {
		return false
	}

	p1x := (float64(s.P1.X) - float64(e.Position.X)) / a
	p1y := (float64(s.P1.Y) - float64(e.Position.Y)) / b
	p2x := (float64(s.P2.X) - float64(e.Position.X)) / a
	p2y := (float64(s.P2.Y) - float64(e.Position.Y)) / b

	t := closestParamToOrigin(p1x, p1y, p2x, p2y)
	cx := p1x + t*(p2x-p1x)
	cy := p1y + t*(p2y-p1y)
	return cx*cx+cy*cy <= 1
}

// RayEllipsoid returns the points where the ray meets the ellipse: p1 is the far
// point and p2 the near one, mirroring RayCircle.
//
// The ray is a half-line, so intersections behind its origin are not reported: a
// ray starting inside the ellipse yields only its exit point, and an ellipse
// wholly behind the origin yields none. The direction is assumed non-zero;
// unlike RayCircle it need not be unit length, since the hit points are exact for
// any direction magnitude.
//
// Working in the space where e is the unit disk, a ray point O+t*D maps to
// O'+t*D' with the SAME parameter t (the affine map is linear in t), so the real
// hit points come straight from the original ray at the t solving |O'+t*D'| = 1.
func RayEllipsoid[T gomath.Float](ray gomath.Ray[T], e gomath.Ellipsoid[T]) (intersects int, p1 gomath.Vector2[T], p2 gomath.Vector2[T]) {
	a, b := e.Radii.X, e.Radii.Y
	if a <= 0 || b <= 0 {
		return 0, p1, p2
	}

	ox := (ray.Origin.X - e.Position.X) / a
	oy := (ray.Origin.Y - e.Position.Y) / b
	dx := ray.Direction.X / a
	dy := ray.Direction.Y / b

	// A t^2 + 2B t + C = 0, so the roots are (-B +/- sqrt(B^2 - A C)) / A.
	pA := dx*dx + dy*dy
	if pA == 0 {
		return 0, p1, p2 // zero-length direction
	}
	pB := ox*dx + oy*dy
	pC := ox*ox + oy*oy - 1
	disc := pB*pB - pA*pC

	if disc < 0 {
		return 0, p1, p2
	}

	// Tangent: one grazing point, on the ray only if it is ahead of the origin.
	if disc == 0 {
		t := -pB / pA
		if t < 0 {
			return 0, p1, p2
		}
		return 1, ray.PointAt(t), p2
	}

	sq := gomath.Sqrt(disc)
	near := (-pB - sq) / pA // pA > 0, so near <= far
	far := (-pB + sq) / pA

	if far < 0 {
		// The ellipse is entirely behind the ray.
		return 0, p1, p2
	}

	p1 = ray.PointAt(far)
	if near < 0 {
		// The origin is inside the ellipse, so only the exit point is on the ray.
		return 1, p1, p2
	}

	p2 = ray.PointAt(near)
	return 2, p1, p2
}

// EllipsoidPointPenetration returns the depth and separation normal for a point
// lying strictly inside the ellipse e. The normal points from the ellipse toward
// the point (the shape1->shape2 MTV convention), and the depth is the true
// Euclidean distance to the nearest boundary point, so translating the point by
// normal*depth moves it exactly onto the boundary. overlapping is false (with a
// zero depth/normal) for a point on or outside the boundary.
func EllipsoidPointPenetration[T gomath.Number](e gomath.Ellipsoid[T], point gomath.Vector2[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	a, b := float64(e.Radii.X), float64(e.Radii.Y)
	if a <= 0 || b <= 0 {
		return 0, normal, false
	}

	qx := float64(point.X) - float64(e.Position.X)
	qy := float64(point.Y) - float64(e.Position.Y)

	// Strictly inside is required for a meaningful push-out.
	if (qx/a)*(qx/a)+(qy/b)*(qy/b) >= 1 {
		return 0, normal, false
	}

	return ellipsoidPushOut[T](a, b, qx, qy)
}

// EllipsoidLineSegmentPenetration returns the depth and separation normal for a
// segment overlapping the filled ellipse e. The normal points from the ellipse
// toward the segment (shape1->shape2), and the depth is the true Euclidean
// distance from the segment's deepest point to the nearest boundary point.
//
// "Deepest" is measured in the ellipse's own metric: the point on the segment
// minimising the implicit value (dx/a)^2+(dy/b)^2 is exactly the point closest to
// the origin once scaled into unit-disk space. For a circular ellipse this
// reduces to CircleLineSegmentPenetration.
func EllipsoidLineSegmentPenetration[T gomath.Number](e gomath.Ellipsoid[T], s gomath.LineSegment[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	a, b := float64(e.Radii.X), float64(e.Radii.Y)
	if a <= 0 || b <= 0 {
		return 0, normal, false
	}

	// Endpoints in centred coords, then scaled into unit-disk space.
	e1x := float64(s.P1.X) - float64(e.Position.X)
	e1y := float64(s.P1.Y) - float64(e.Position.Y)
	e2x := float64(s.P2.X) - float64(e.Position.X)
	e2y := float64(s.P2.Y) - float64(e.Position.Y)

	t := closestParamToOrigin(e1x/a, e1y/b, e2x/a, e2y/b)

	// Deepest point in centred coords (t maps back through the original segment).
	qx := e1x + t*(e2x-e1x)
	qy := e1y + t*(e2y-e1y)

	if (qx/a)*(qx/a)+(qy/b)*(qy/b) >= 1 {
		return 0, normal, false
	}

	return ellipsoidPushOut[T](a, b, qx, qy)
}

// ellipsoidPushOut builds the (depth, unit normal, true) MTV that moves the
// interior centred point (qx, qy) onto the nearest boundary of the origin-centred
// ellipse with semi-axes (a, b). Callers must have already established the point
// is strictly inside.
func ellipsoidPushOut[T gomath.Number](a, b, qx, qy float64) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	cx, cy := closestPointOnEllipse(a, b, qx, qy)
	nx := cx - qx
	ny := cy - qy
	depth := math.Hypot(nx, ny)
	if depth == 0 {
		return 0, normal, false
	}
	inv := 1 / depth
	return T(depth), gomath.Vector2[T]{X: T(nx * inv), Y: T(ny * inv)}, true
}

// closestParamToOrigin returns the parameter t in [0, 1] of the point on the
// segment (p1)->(p2) closest to the origin.
func closestParamToOrigin(p1x, p1y, p2x, p2y float64) float64 {
	dx := p2x - p1x
	dy := p2y - p1y
	den := dx*dx + dy*dy
	if den == 0 {
		return 0
	}
	t := -(p1x*dx + p1y*dy) / den
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

// EllipsoidCircle reports whether the circle c touches or overlaps the filled
// ellipse e (boundary inclusive). It is exact: the circle overlaps iff its centre
// is within c.Radius of the ellipse boundary.
func EllipsoidCircle[T gomath.Number](e gomath.Ellipsoid[T], c gomath.Circle[T]) bool {
	a, b := float64(e.Radii.X), float64(e.Radii.Y)
	if a <= 0 || b <= 0 {
		return false
	}
	cx := float64(c.Center.X) - float64(e.Position.X)
	cy := float64(c.Center.Y) - float64(e.Position.Y)
	return distancePointEllipse(a, b, cx, cy) <= float64(c.Radius)
}

// EllipsoidAABB reports whether the axis-aligned box overlaps the filled ellipse
// e (boundary inclusive). Scaling by (1/Radii.X, 1/Radii.Y) turns e into the unit
// disk and keeps the box axis-aligned, so the test reduces exactly to a
// circle-vs-box overlap in that space.
func EllipsoidAABB[T gomath.Number](e gomath.Ellipsoid[T], box gomath.AABB[T]) bool {
	a, b := float64(e.Radii.X), float64(e.Radii.Y)
	if a <= 0 || b <= 0 {
		return false
	}
	scaled := gomath.AABB[float64]{
		Center: gomath.Vector2[float64]{
			X: (float64(box.Center.X) - float64(e.Position.X)) / a,
			Y: (float64(box.Center.Y) - float64(e.Position.Y)) / b,
		},
		HalfWidth:  float64(box.HalfWidth) / a,
		HalfHeight: float64(box.HalfHeight) / b,
	}
	return CircleAABB(gomath.Circle[float64]{Radius: 1}, scaled)
}

// EllipsoidEllipsoidPenetration returns an approximate depth and separation normal
// for two overlapping ellipses. The normal points from e1 toward e2 (shape1->shape2)
// along the line of centres, and the depth is (r1+r2) - dist where r1, r2 are each
// ellipse's radius toward the other. This is a cheap, orientation-aware MTV meant
// for an iterative solver, not the exact minimum translation; it reduces to
// CircleCirclePenetration for circular ellipses.
func EllipsoidEllipsoidPenetration[T gomath.Number](e1, e2 gomath.Ellipsoid[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	a1, b1 := float64(e1.Radii.X), float64(e1.Radii.Y)
	a2, b2 := float64(e2.Radii.X), float64(e2.Radii.Y)
	if a1 <= 0 || b1 <= 0 || a2 <= 0 || b2 <= 0 {
		return 0, normal, false
	}

	dx := float64(e2.Position.X) - float64(e1.Position.X)
	dy := float64(e2.Position.Y) - float64(e1.Position.Y)
	dist := math.Hypot(dx, dy)

	if dist == 0 {
		// Coincident centres: separate along whichever axis needs the least push.
		if a1+a2 <= b1+b2 {
			return T(a1 + a2), gomath.Vector2[T]{X: 1}, true
		}
		return T(b1 + b2), gomath.Vector2[T]{Y: 1}, true
	}

	pen := ellipseRadius(a1, b1, dx, dy) + ellipseRadius(a2, b2, dx, dy) - dist
	if pen <= 0 {
		return 0, normal, false
	}
	inv := 1 / dist
	return T(pen), gomath.Vector2[T]{X: T(dx * inv), Y: T(dy * inv)}, true
}

// EllipsoidCirclePenetration returns an approximate depth and separation normal
// for a circle overlapping the ellipse e. The normal points from the ellipse
// toward the circle (shape1->shape2). A circle is a degenerate ellipse, so this
// delegates to EllipsoidEllipsoidPenetration.
func EllipsoidCirclePenetration[T gomath.Number](e gomath.Ellipsoid[T], c gomath.Circle[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	return EllipsoidEllipsoidPenetration(e, gomath.Ellipsoid[T]{
		Position: c.Center,
		Radii:    gomath.Vector2[T]{X: c.Radius, Y: c.Radius},
	})
}

// EllipsoidAABBPenetration returns an approximate depth and separation normal for
// an ellipse overlapping an axis-aligned box. The normal points from the ellipse
// toward the box (shape1->shape2), matching CircleAABBPenetration's convention so
// consumers can wire it the same way.
//
// When the centre is outside the box it uses the box's closest point and the
// ellipse's radius toward it; when the centre is inside it pushes out the nearest
// face (as CircleAABBPenetration does). Cheap and orientation-aware, for an
// iterative solver rather than an exact MTV.
func EllipsoidAABBPenetration[T gomath.Number](e gomath.Ellipsoid[T], box gomath.AABB[T]) (penetration T, normal gomath.Vector2[T], overlapping bool) {
	a, b := float64(e.Radii.X), float64(e.Radii.Y)
	if a <= 0 || b <= 0 {
		return 0, normal, false
	}

	closestX := gomath.Clamp(e.Position.X, box.Center.X-box.HalfWidth, box.Center.X+box.HalfWidth)
	closestY := gomath.Clamp(e.Position.Y, box.Center.Y-box.HalfHeight, box.Center.Y+box.HalfHeight)

	dx := float64(e.Position.X - closestX)
	dy := float64(e.Position.Y - closestY)
	distSq := dx*dx + dy*dy

	// Centre inside the box: push out the nearest face, using the ellipse radius
	// along that axis (a on X, b on Y).
	if distSq == 0 {
		px := float64(box.HalfWidth) - math.Abs(float64(e.Position.X-box.Center.X))
		py := float64(box.HalfHeight) - math.Abs(float64(e.Position.Y-box.Center.Y))
		if px < py {
			if e.Position.X > box.Center.X {
				return T(px + a), gomath.Vector2[T]{X: -1}, true
			}
			return T(px + a), gomath.Vector2[T]{X: 1}, true
		}
		if e.Position.Y > box.Center.Y {
			return T(py + b), gomath.Vector2[T]{Y: -1}, true
		}
		return T(py + b), gomath.Vector2[T]{Y: 1}, true
	}

	dist := math.Sqrt(distSq)
	pen := ellipseRadius(a, b, dx, dy) - dist
	if pen <= 0 {
		return 0, normal, false
	}
	inv := 1 / dist
	return T(pen), gomath.Vector2[T]{X: T(-dx * inv), Y: T(-dy * inv)}, true
}

// ellipseRadius returns the distance from the centre of the origin-centred,
// axis-aligned ellipse with semi-axes (a, b) to its boundary along the direction
// (dx, dy). The direction need not be unit length.
func ellipseRadius(a, b, dx, dy float64) float64 {
	l := math.Hypot(dx, dy)
	if l == 0 {
		return 0
	}
	return l / math.Sqrt((dx/a)*(dx/a)+(dy/b)*(dy/b))
}

// closestPointOnEllipse returns the point on the origin-centred, axis-aligned
// ellipse with semi-axes (a, b) closest to (px, py), for a query point inside,
// on, or outside the ellipse. It reflects the query into the first quadrant and
// puts the major axis first (Eberly's preconditions), then maps the result back.
func closestPointOnEllipse(a, b, px, py float64) (float64, float64) {
	sx, ax := 1.0, px
	if px < 0 {
		sx, ax = -1, -px
	}
	sy, ay := 1.0, py
	if py < 0 {
		sy, ay = -1, -py
	}

	var cx, cy float64
	if a >= b {
		cx, cy = eberlyClosestPoint(a, b, ax, ay)
	} else {
		// Major axis is Y: solve with axes swapped, then unswap the result.
		y0, y1 := eberlyClosestPoint(b, a, ay, ax)
		cx, cy = y1, y0
	}
	return sx * cx, sy * cy
}

// eberlyClosestPoint returns the closest point on the origin-centred ellipse with
// semi-axes e0 >= e1 > 0 to (y0, y1), requiring y0, y1 >= 0. It is the
// closest-point companion of eberlyDistance, from the same Eberly reference.
func eberlyClosestPoint(e0, e1, y0, y1 float64) (x0, x1 float64) {
	if y1 > 0 {
		if y0 > 0 {
			z0, z1 := y0/e0, y1/e1
			g := z0*z0 + z1*z1 - 1
			if g == 0 {
				return y0, y1 // already on the ellipse
			}
			r0 := (e0 / e1) * (e0 / e1)
			sbar := eberlyRoot(r0, z0, z1, g)
			return r0 * y0 / (sbar + r0), y1 / (sbar + 1)
		}
		// y0 == 0: closest point is on the minor axis.
		return 0, e1
	}

	// y1 == 0.
	numer0 := e0 * y0
	denom0 := e0*e0 - e1*e1
	if numer0 < denom0 {
		xde0 := numer0 / denom0
		return e0 * xde0, e1 * math.Sqrt(1-xde0*xde0)
	}
	return e0, 0
}

// EllipsoidEllipsoid reports whether two axis-aligned ellipsoids overlap. Touching
// counts as overlapping.
//
// Scaling by (1/e1.Radii.X, 1/e1.Radii.Y) is invertible, and an invertible affine
// map preserves whether two regions intersect. It turns e1 into the unit disk at
// the origin and leaves e2 axis-aligned, so the question becomes whether e2 passes
// within distance 1 of the origin.
func EllipsoidEllipsoid[T gomath.Number](e1, e2 gomath.Ellipsoid[T]) bool {
	a1, b1 := float64(e1.Radii.X), float64(e1.Radii.Y)
	a2, b2 := float64(e2.Radii.X), float64(e2.Radii.Y)
	if a1 <= 0 || b1 <= 0 || a2 <= 0 || b2 <= 0 {
		return false
	}

	// e2's centre relative to e1, in the space where e1 is the unit disk.
	cx := (float64(e2.Position.X) - float64(e1.Position.X)) / a1
	cy := (float64(e2.Position.Y) - float64(e1.Position.Y)) / b1

	return distancePointEllipse(a2/a1, b2/b1, cx, cy) <= 1+1e-9
}

// distancePointEllipse returns the distance from the origin to the region bounded
// by the axis-aligned ellipse with semi-axes (a, b) centred at (cx, cy). It is 0
// when the origin lies inside the ellipse.
func distancePointEllipse(a, b, cx, cy float64) float64 {
	// Work in the ellipse's own frame: the query point is the origin, so relative
	// to the centre it sits at (-cx, -cy).
	y0, y1 := -cx, -cy

	// Inside (or on) the ellipse means distance 0.
	if (y0/a)*(y0/a)+(y1/b)*(y1/b) <= 1 {
		return 0
	}

	// Eberly's method needs the first quadrant and the major axis first; the
	// ellipse is symmetric about both axes, so reflect and swap as needed.
	sx, sy := math.Abs(y0), math.Abs(y1)
	if a >= b {
		return eberlyDistance(a, b, sx, sy)
	}
	return eberlyDistance(b, a, sy, sx)
}

// eberlyDistance returns the distance from (y0, y1) to the origin-centred
// axis-aligned ellipse with semi-axes e0 >= e1 > 0, requiring y0, y1 >= 0.
// From Eberly, "Distance from a Point to an Ellipse, an Ellipsoid, or a
// Hyperellipsoid" (https://www.geometrictools.com/Documentation/DistancePointEllipseEllipsoid.pdf).
func eberlyDistance(e0, e1, y0, y1 float64) float64 {
	if y1 > 0 {
		if y0 > 0 {
			z0, z1 := y0/e0, y1/e1
			g := z0*z0 + z1*z1 - 1
			if g == 0 {
				return 0
			}
			r0 := (e0 / e1) * (e0 / e1)
			sbar := eberlyRoot(r0, z0, z1, g)
			x0 := r0 * y0 / (sbar + r0)
			x1 := y1 / (sbar + 1)
			return math.Hypot(x0-y0, x1-y1)
		}
		// y0 == 0: closest point is on the minor axis.
		return math.Abs(y1 - e1)
	}

	// y1 == 0.
	numer0 := e0 * y0
	denom0 := e0*e0 - e1*e1
	if numer0 < denom0 {
		xde0 := numer0 / denom0
		x0 := e0 * xde0
		x1 := e1 * math.Sqrt(1-xde0*xde0)
		return math.Hypot(x0-y0, x1)
	}
	return math.Abs(y0 - e0)
}

// eberlyRoot bisects for the root of the Lagrange-multiplier equation.
func eberlyRoot(r0, z0, z1, g float64) float64 {
	n0 := r0 * z0
	s0 := z1 - 1
	s1 := 0.0
	if g >= 0 {
		s1 = math.Hypot(n0, z1) - 1
	}

	s := s0
	for i := 0; i < 149; i++ {
		s = (s0 + s1) / 2
		if s == s0 || s == s1 {
			break
		}
		ratio0 := n0 / (s + r0)
		ratio1 := z1 / (s + 1)
		g := ratio0*ratio0 + ratio1*ratio1 - 1
		if g > 0 {
			s0 = s
		} else if g < 0 {
			s1 = s
		} else {
			break
		}
	}
	return s
}
