package intersection

import "github.com/wolfgarnet/gomath"

func RayCircle[T gomath.Float](ray gomath.Ray[T], circle gomath.Circle[T]) (intersects int, p1 gomath.Vector2[T], p2 gomath.Vector2[T]) {

	toCircle := circle.Center.Sub(ray.Origin)
	projection := ray.Direction.MultiplyScalar(toCircle.Dot(ray.Direction))
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
