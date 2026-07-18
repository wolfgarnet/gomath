package gomath

// Circle is a circle defined by a centre and radius, treated as a filled disk by
// the containment and overlap helpers. Integer T works exactly for the squared
// tests (Contains, Intersects, ...) but is storage-only for operations that need
// a true magnitude or pi (ClosestPoint, Area, Circumference). See
// circle_additions.go for those helpers.
type Circle[T Number] struct {
	Center Vector2[T]
	Radius T
}

type CircleCollider[T Number] struct {
	Radius T
}

func NewCircle[T Number](x, y, radius T) Circle[T] {
	return Circle[T]{
		Center: Vector2[T]{x, y},
		Radius: radius,
	}
}

func NewCircle2[T Number](radius T) CircleCollider[T] {
	return CircleCollider[T]{
		Radius: radius,
	}
}

func NewCircleWithCenter[T Number](center Vector2[T], radius T) Circle[T] {
	return Circle[T]{
		Center: center,
		Radius: radius,
	}
}

func (c Circle[T]) GetCenter() Vector2[T] {
	return c.Center
}

func (c Circle[T]) GetHalfWidth() T {
	return c.Radius
}

func (c Circle[T]) GetHalfHeight() T {
	return c.Radius
}

// Scaled returns a copy of the circle with its radius multiplied by scale.
func (c Circle[T]) Scaled(scale T) Circle[T] {
	c.Radius *= scale
	return c
}

// WithCenter returns a copy of the circle centred at center.
func (c Circle[T]) WithCenter(center Vector2[T]) Circle[T] {
	c.Center = center
	return c
}

func (c Circle[T]) TangentLinePoints(p Vector2[T]) (t1, t2 Vector2[T], hasSolution bool) {
	dxy := p.Sub(c.Center)
	d2 := dxy.X*dxy.X + dxy.Y*dxy.Y
	d := Sqrt(d2)
	if d >= c.Radius {
		rho := c.Radius / d
		ad := rho * rho
		t := 1 - rho*rho
		bd := rho * Sqrt(t)
		t1 = NewVector2(c.Center.X+ad*dxy.X+bd*-dxy.Y, c.Center.Y+ad*dxy.Y+bd*dxy.X)
		t2 = NewVector2(c.Center.X+ad*dxy.X-bd*-dxy.Y, c.Center.Y+ad*dxy.Y-bd*dxy.X)
		hasSolution = true
	}

	return t1, t2, hasSolution
}
