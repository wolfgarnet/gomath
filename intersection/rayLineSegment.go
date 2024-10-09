package intersection

import "github.com/wolfgarnet/gomath"

func RayLineSegment2[T gomath.Float](ray gomath.Ray[T], lineSegment gomath.LineSegment[T]) (rayRatio, lineRatio T, intersects bool) {
	return RayLineSegment(ray, lineSegment.P1, lineSegment.P2)
}

// RayLineSegment determines if a ray intersects with a line segment.
// The rayRatio determines how far along the ray, it intersects with the line segment. It is multiples of the direction vector, which is typically normalized.
// The lineRatio determines the ratio of the line segment where it intersects with the ray, beginning from p1.
func RayLineSegment[T gomath.Float](ray gomath.Ray[T], p1, p2 gomath.Vector2[T]) (rayRatio, lineRatio T, intersects bool) {
	seg := p2.Sub(p1) //l2 - l1
	segPerp := gomath.NewVector2(seg.Y, -seg.X)
	perpDotd := ray.Direction.Dot(segPerp) //Vector2.Dot(rd, segPerp);

	// If lines are parallel, return false.
	//if (Math.Abs(perpDotd) <= float.Epsilon)
	if gomath.Abs(perpDotd) <= 0.00001 {
		//t = float.MaxValue;
		return rayRatio, lineRatio, false
	}

	//Vector2 d = l1 - ro;
	d := p1.Sub(ray.Origin)

	//t = Vector2.Dot(segPerp, d) / perpDotd;
	rayRatio = segPerp.Dot(d) / perpDotd
	//float s = Vector2.Dot(new Vector2(rd.Y, -rd.X), d) / perpDotd;
	lineRatio = gomath.NewVector2(ray.Direction.Y, -ray.Direction.X).Dot(d) / perpDotd

	// If intersect is in right direction and in segment bounds, return true.
	//return t >= 0.0f && s >= 0.0f && s <= 1.0f;
	return rayRatio, lineRatio, rayRatio >= 0 && lineRatio >= 0 && lineRatio <= 1
}
