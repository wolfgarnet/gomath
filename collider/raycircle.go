package collider

import "github.com/wolfgarnet/gomath"

// RayCircle returns the points where the ray meets the circle: p1 is the far point
// and p2 the near one.
//
// The ray is a half-line, so intersections behind its origin are not reported: a
// ray starting inside the circle yields only its exit point, and a circle wholly
// behind the origin yields none. The direction is assumed to be of unit length.
func RayCircle[T gomath.Float](ray gomath.Ray[T], circle gomath.Circle[T]) (intersects int, p1 gomath.Vector2[T], p2 gomath.Vector2[T]) {

	toCircle := circle.Center.Sub(ray.Origin)
	along := toCircle.Dot(ray.Direction)
	projection := ray.Direction.MultiplyScalar(along)
	u2 := toCircle.Sub(projection)
	d := u2.Length()

	if d > circle.Radius {
		return 0, p1, p2
	}

	m := gomath.Sqrt(circle.Radius*circle.Radius - d*d)

	// Tangent: one grazing point, on the ray only if it is ahead of the origin.
	if m == 0 {
		if along < 0 {
			return 0, p1, p2
		}
		p1 = ray.Origin.Add(projection)
		return 1, p1, p2
	}

	// How far along the ray the two intersections lie. Negative means behind the
	// origin, and so not on the ray at all.
	near, far := along-m, along+m

	if far < 0 {
		// The circle is entirely behind the ray.
		return 0, p1, p2
	}

	p1 = ray.Origin.Add(ray.Direction.MultiplyScalar(far))

	if near < 0 {
		// The origin is inside the circle, so only the exit point is on the ray.
		return 1, p1, p2
	}

	p2 = ray.Origin.Add(ray.Direction.MultiplyScalar(near))

	return 2, p1, p2
}
