package intersection

import (
	"math"

	"github.com/wolfgarnet/gomath"
)

/*
https://www.geeksforgeeks.org/check-if-a-point-is-inside-outside-or-on-the-ellipse/
*/

func EllipsoidPoint[T gomath.Number](point gomath.Vector2[T], ellipsoid gomath.Ellipsoid[T]) bool {
	// Normalize the point in the ellipsoid's local space
	normalizedPoint := gomath.Vector2[T]{X: point.X / ellipsoid.Radii.X, Y: point.Y / ellipsoid.Radii.Y}

	// Check if the normalized point is inside the unit circle
	return normalizedPoint.Length() <= 1.0
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
