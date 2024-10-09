package gomath

type Circle[T Number] struct {
	Center Vector2[T]
	Radius T
}

func NewCircle[T Number](x, y, radius T) Circle[T] {
	return Circle[T]{
		Center: Vector2[T]{x, y},
		Radius: radius,
	}
}

func NewCircleWithCenter[T Number](center Vector2[T], radius T) Circle[T] {
	return Circle[T]{
		Center: center,
		Radius: radius,
	}
}

func (c *Circle[T]) Scale(scale T) {
	c.Radius *= scale
}

func (c *Circle[T]) SetCenter(center Vector2[T]) {
	c.Center = center
}

func (c *Circle[T]) Scale(s T) {
	c.Radius *= s
}

func (c *Circle[T]) GetCenter() Vector2[T] {
	return c.Center
}

func (c *Circle[T]) GetHalfWidth() T {
	return c.Radius
}

func (c *Circle[T]) GetHalfHeight() T {
	return c.Radius
}

func (c *Circle[T]) SetHalfWidth(w T) {
	c.Radius = w
}

func (c *Circle[T]) SetHalfHeight(h T) {
	c.Radius = h
}

func (c *Circle[T]) TangentLinePoints(p Vector2[T]) (t1, t2 Vector2[T], hasSolution bool) {
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
