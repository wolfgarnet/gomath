package intersection

import "github.com/wolfgarnet/gomath"

func RayCircle[T gomath.Float](ray gomath.Ray[T], circle gomath.Circle[T]) (intersects int, p1 gomath.Vector2[T], p2 gomath.Vector2[T]) {
	/*
	   	C = Circle('C', Point2(0,0), 3.0)
	      R = Line(Point2(2.5, -4.0), Vector2(0,1))

	      U = C.center - R.point

	      # Project U onto V
	      U1 = U.dot(R.direction)*R.direction

	      # Vector to circle center
	      U2 = U - U1

	      d = U2.norm()

	      m = np.sqrt(C.radius**2 - d**2).item()

	      P1 = R.point + U1 + m*R.direction
	      P2 = R.point + U1 - m*R.direction
	*/

	toCircle := circle.Center.Sub(ray.Origin)
	projection := ray.Direction.MultiplyScalar(toCircle.Dot(ray.Direction))
	//d := toCircle.Normalize()
	u2 := toCircle.Sub(projection)
	d := u2.Length()

	if d > circle.Radius {
		return 0, p1, p2
	}

	m := gomath.Sqrt(circle.Radius*circle.Radius - d*d)

	if m == 0 {
		p1 = ray.Origin.Add(projection)
		return 1, p1, p2
	}

	p1 = ray.Origin.Add(projection).Add(ray.Direction.MultiplyScalar(m))
	p2 = ray.Origin.Add(projection).Sub(ray.Direction.MultiplyScalar(m))

	return 2, p1, p2
}
