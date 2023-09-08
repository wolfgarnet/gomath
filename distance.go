package gomath

func dist2[T Number](v, w Vector2[T]) T { return Sqr(v.X-w.X) + Sqr(v.Y-w.Y) }

// Minimum Distance calculates the minimum distance from a point, p,
// to the line segment defined by a and b.
// https://stackoverflow.com/questions/849211/shortest-distance-between-a-point-and-a-line-segment
func MinimumDistance[T Number](a, b, p Vector2[T]) T {
	l2 := dist2(a, b)
	if l2 == 0 {
		return p.Sub(a).Length()
	}

	t := ((p.X-a.X)*(b.X-a.X) + (p.Y-a.Y)*(b.Y-a.Y)) / l2
	t = Max(0, Min(1, t))

	tmp := NewVector2(
		a.X+t*(b.X-a.X),
		a.Y+t*(b.Y-a.Y),
	)

	return p.Sub(tmp).Length()
}
