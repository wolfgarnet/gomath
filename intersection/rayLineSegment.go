package intersection

import "github.com/wolfgarnet/gomath"

func RayLineSegment2[T gomath.Float](ray gomath.Ray[T], lineSegment gomath.LineSegment[T]) (rayRatio, lineRatio T, intersects bool) {
	return RayLineSegment(ray, lineSegment.P1, lineSegment.P2)
}

func RayLineSegment[T gomath.Float](ray gomath.Ray[T], P1, P2 gomath.Vector2[T]) (rayRatio, lineRatio T, intersects bool) {
	seg := P2.Sub(P1) //l2 - l1
	segPerp := gomath.NewVector2(seg.Y, -seg.X)
	perpDotd := ray.Direction.Dot(segPerp) //Vector2.Dot(rd, segPerp);

	// If lines are parallel, return false.
	//if (Math.Abs(perpDotd) <= float.Epsilon)
	if gomath.Abs(perpDotd) <= 0.00001 {
		//t = float.MaxValue;
		return rayRatio, lineRatio, false
	}

	//Vector2 d = l1 - ro;
	d := P1.Sub(ray.Origin)

	//t = Vector2.Dot(segPerp, d) / perpDotd;
	rayRatio = segPerp.Dot(d) / perpDotd
	//float s = Vector2.Dot(new Vector2(rd.Y, -rd.X), d) / perpDotd;
	lineRatio = gomath.NewVector2(ray.Direction.Y, -ray.Direction.X).Dot(d) / perpDotd

	// If intersect is in right direction and in segment bounds, return true.
	//return t >= 0.0f && s >= 0.0f && s <= 1.0f;
	return rayRatio, lineRatio, rayRatio >= 0 && lineRatio >= 0 && lineRatio <= 1
}
