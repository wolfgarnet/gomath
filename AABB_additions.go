package gomath

// This file rounds the AABB type out into a production-grade axis-aligned box
// for games: containment, overlap, union/intersection and construction helpers.
// All additions are value-receiver and non-mutating, layered on top of the
// existing (Center + half-extent) representation.
//
// An AABB spans [Center.X-HalfWidth, Center.X+HalfWidth] x
// [Center.Y-HalfHeight, Center.Y+HalfHeight]; Min/Max return those corners and
// are independent of whether Y points up or down. Constructors and the
// union/intersection helpers derive Center as (min+max)/2, which truncates for
// integer T when a span is odd - use a floating-point T for exactness.

// aabbFromMinMax builds an AABB covering the box spanned by min and max.
func aabbFromMinMax[T Number](min, max Vector2[T]) AABB[T] {
	return AABB[T]{
		Center:     Vector2[T]{X: (min.X + max.X) / 2, Y: (min.Y + max.Y) / 2},
		HalfWidth:  (max.X - min.X) / 2,
		HalfHeight: (max.Y - min.Y) / 2,
	}
}

// Min returns the minimum corner (Center - half-extents).
func (a AABB[T]) Min() Vector2[T] {
	return Vector2[T]{X: a.Center.X - a.HalfWidth, Y: a.Center.Y - a.HalfHeight}
}

// Max returns the maximum corner (Center + half-extents).
func (a AABB[T]) Max() Vector2[T] {
	return Vector2[T]{X: a.Center.X + a.HalfWidth, Y: a.Center.Y + a.HalfHeight}
}

// Width returns the full width (2 * HalfWidth).
func (a AABB[T]) Width() T { return a.HalfWidth * 2 }

// Height returns the full height (2 * HalfHeight).
func (a AABB[T]) Height() T { return a.HalfHeight * 2 }

// Size returns the full extents {Width, Height}.
func (a AABB[T]) Size() Vector2[T] { return Vector2[T]{X: a.HalfWidth * 2, Y: a.HalfHeight * 2} }

// Area returns the area of the box.
func (a AABB[T]) Area() T { return a.Width() * a.Height() }

// Contains reports whether point lies inside the box, edges inclusive.
func (a AABB[T]) Contains(point Vector2[T]) bool {
	return point.X >= a.Center.X-a.HalfWidth && point.X <= a.Center.X+a.HalfWidth &&
		point.Y >= a.Center.Y-a.HalfHeight && point.Y <= a.Center.Y+a.HalfHeight
}

// ContainsAABB reports whether other lies entirely inside a, edges inclusive.
func (a AABB[T]) ContainsAABB(other AABB[T]) bool {
	amin, amax := a.Min(), a.Max()
	omin, omax := other.Min(), other.Max()
	return amin.X <= omin.X && amax.X >= omax.X && amin.Y <= omin.Y && amax.Y >= omax.Y
}

// Intersects reports whether a and other overlap. Edge-touching boxes count as
// intersecting.
func (a AABB[T]) Intersects(other AABB[T]) bool {
	amin, amax := a.Min(), a.Max()
	omin, omax := other.Min(), other.Max()
	return amin.X <= omax.X && amax.X >= omin.X && amin.Y <= omax.Y && amax.Y >= omin.Y
}

// Intersection returns the overlapping box of a and other, and true, or the zero
// AABB and false if they do not overlap.
func (a AABB[T]) Intersection(other AABB[T]) (AABB[T], bool) {
	amin, amax := a.Min(), a.Max()
	omin, omax := other.Min(), other.Max()

	lo := Vector2[T]{X: max(amin.X, omin.X), Y: max(amin.Y, omin.Y)}
	hi := Vector2[T]{X: min(amax.X, omax.X), Y: min(amax.Y, omax.Y)}
	if lo.X > hi.X || lo.Y > hi.Y {
		return AABB[T]{}, false
	}
	return aabbFromMinMax(lo, hi), true
}

// Union returns the smallest box containing both a and other.
func (a AABB[T]) Union(other AABB[T]) AABB[T] {
	amin, amax := a.Min(), a.Max()
	omin, omax := other.Min(), other.Max()
	return aabbFromMinMax(
		Vector2[T]{X: min(amin.X, omin.X), Y: min(amin.Y, omin.Y)},
		Vector2[T]{X: max(amax.X, omax.X), Y: max(amax.Y, omax.Y)},
	)
}

// ClosestPoint returns the point on or inside the box closest to point (point
// clamped into the box).
func (a AABB[T]) ClosestPoint(point Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: Clamp(point.X, a.Center.X-a.HalfWidth, a.Center.X+a.HalfWidth),
		Y: Clamp(point.Y, a.Center.Y-a.HalfHeight, a.Center.Y+a.HalfHeight),
	}
}

// ExpandToInclude returns the smallest box containing both a and point.
func (a AABB[T]) ExpandToInclude(point Vector2[T]) AABB[T] {
	lo, hi := a.Min(), a.Max()
	return aabbFromMinMax(
		Vector2[T]{X: min(lo.X, point.X), Y: min(lo.Y, point.Y)},
		Vector2[T]{X: max(hi.X, point.X), Y: max(hi.Y, point.Y)},
	)
}

// Equals reports whether a and other have the same centre and half-extents.
func (a AABB[T]) Equals(other AABB[T]) bool {
	return a == other
}

// NewAABBFromMinMax builds an AABB spanning the box between corners min and max.
func NewAABBFromMinMax[T Number](min, max Vector2[T]) AABB[T] {
	return aabbFromMinMax(min, max)
}

// NewAABBFromPoints returns the smallest AABB containing all the given points.
// With no points it returns the zero AABB.
func NewAABBFromPoints[T Number](points ...Vector2[T]) AABB[T] {
	if len(points) == 0 {
		return AABB[T]{}
	}
	lo, hi := points[0], points[0]
	for _, p := range points[1:] {
		lo.X = min(lo.X, p.X)
		lo.Y = min(lo.Y, p.Y)
		hi.X = max(hi.X, p.X)
		hi.Y = max(hi.Y, p.Y)
	}
	return aabbFromMinMax(lo, hi)
}
