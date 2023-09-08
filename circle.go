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
