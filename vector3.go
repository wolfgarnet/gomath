package gomath

import (
	"fmt"
	"math"
)

type Vector3[T Number] struct {
	X T
	Y T
	Z T
}

func NewVector3[T Number](x T, y T, z T) Vector3[T] {
	return Vector3[T]{
		X: x,
		Y: y,
		Z: z,
	}
}

func (v Vector3[T]) Set(x T, y T, z T) Vector3[T] {
	v.X = x
	v.Y = y
	v.Z = z

	return v
}

func (v Vector3[T]) Add(v2 Vector3[T]) Vector3[T] {
	return Vector3[T]{
		X: v.X + v2.X,
		Y: v.Y + v2.Y,
		Z: v.Z + v2.Z,
	}
}

func (v Vector3[T]) Component(i int) T {
	switch i {
	case 0:
		return v.X
	case 1:
		return v.Y
	case 2:
		return v.Z
	default:
		panic("out of range")
	}
}

func (v Vector3[T]) AddScalar(scalar T) Vector3[T] {
	return Vector3[T]{
		X: v.X + scalar,
		Y: v.Y + scalar,
		Z: v.Z + scalar,
	}
}

func (v Vector3[T]) AddScalars(x T, y T, z T) Vector3[T] {
	return Vector3[T]{
		X: v.X + x,
		Y: v.Y + y,
		Z: v.Z + z,
	}
}

func (v Vector3[T]) Sub(v2 Vector3[T]) Vector3[T] {
	return Vector3[T]{
		X: v.X - v2.X,
		Y: v.Y - v2.Y,
		Z: v.Z - v2.Z,
	}
}

func (v Vector3[T]) SubScalar(scalar T) Vector3[T] {
	return Vector3[T]{
		X: v.X - scalar,
		Y: v.Y - scalar,
		Z: v.Z - scalar,
	}
}

func (v Vector3[T]) SubScalars(x T, y T, z T) Vector3[T] {
	return Vector3[T]{
		X: v.X - x,
		Y: v.Y - y,
		Z: v.Z - z,
	}
}

func (v Vector3[T]) Mul(v2 Vector3[T]) Vector3[T] {
	return Vector3[T]{
		X: v.X * v2.X,
		Y: v.Y * v2.Y,
		Z: v.Z * v2.Z,
	}
}

func (v Vector3[T]) MultiplyScalar(scalar T) Vector3[T] {
	return Vector3[T]{
		X: v.X * scalar,
		Y: v.Y * scalar,
		Z: v.Z * scalar,
	}
}

func (v Vector3[T]) MultiplyScalars(x T, y T, z T) Vector3[T] {
	return Vector3[T]{
		X: v.X * x,
		Y: v.Y * y,
		Z: v.Z * z,
	}
}

func (v Vector3[T]) Div(v2 Vector3[T]) Vector3[T] {
	return Vector3[T]{
		X: v.X / v2.X,
		Y: v.Y / v2.Y,
		Z: v.Z / v2.Z,
	}
}

func (v Vector3[T]) DivScalar(scalar T) Vector3[T] {
	return Vector3[T]{
		X: v.X / scalar,
		Y: v.Y / scalar,
		Z: v.Z / scalar,
	}
}

func (v Vector3[T]) DivScalars(x T, y T, z T) Vector3[T] {
	return Vector3[T]{
		X: v.X / x,
		Y: v.Y / y,
		Z: v.Z / z,
	}
}

func (v Vector3[T]) Distance(v2 Vector3[T]) float64 {
	dx := v.X - v2.X
	dy := v.Y - v2.Y
	dz := v.Z - v2.Z
	return math.Sqrt(float64(dx*dx + dy*dy + dz*dz))
}

func (v Vector3[T]) Dot(v2 Vector3[T]) T {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z
}

func (v Vector3[T]) Lerp(v2 Vector3[T], t T) Vector3[T] {
	return Vector3[T]{
		X: v.X + (v2.X-v.X)*t,
		Y: v.Y + (v2.Y-v.Y)*t,
		Z: v.Z + (v2.Z-v.Z)*t,
	}
}

func (v Vector3[T]) Length64() float64 {
	return math.Sqrt(float64(v.X)*float64(v.X) + float64(v.Y)*float64(v.Y) + float64(v.Z)*float64(v.Z))
}

func (v Vector3[T]) Length() T {
	return T(v.Length64())
}

func (v Vector3[T]) LengthSq() T {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vector3[T]) Length32() float32 {
	return float32(v.Length64())
}

func (v Vector3[T]) SetLength(l T) Vector3[T] {
	oldLength := v.Length()
	if oldLength != 0 && l != oldLength {
		return v.MultiplyScalar(l / oldLength)
	}

	return v
}

func (v Vector3[T]) Normalize() Vector3[T] {
	m := v.Length()

	if float64(m) > Epsilon {
		return v.DivScalar(m)
	} else {
		return v
	}
}

// Reflect reflects the vector about the given normal, which is assumed to be of unit length.
func (v Vector3[T]) Reflect(normal Vector3[T]) Vector3[T] {
	factor := -T(2.0) * v.Dot(normal)
	return Vector3[T]{
		X: factor*normal.X + v.X,
		Y: factor*normal.Y + v.Y,
		Z: factor*normal.Z + v.Z,
	}
}

func (v Vector3[T]) Equals(v2 Vector3[T]) bool {
	return v.X == v2.X && v.Y == v2.Y && v.Z == v2.Z
}

// AngleTo returns the unsigned angle in radians between this vector and v2.
func (v Vector3[T]) AngleTo(v2 Vector3[T]) T {
	theta := v.Dot(v2) / (v.Length() * v2.Length())
	return Acos(Clamp[T](theta, T(-1.0), 1))
}

// DistanceTo returns the distance to the given vector
func (v Vector3[T]) DistanceTo(v2 Vector3[T]) T {
	return Sqrt(v.DistanceToSquared(v2))
}

// DistanceToSquared returns the squared distance to the given vector
func (v Vector3[T]) DistanceToSquared(v2 Vector3[T]) T {
	dx := v.X - v2.X
	dy := v.Y - v2.Y
	dz := v.Z - v2.Z

	return dx*dx + dy*dy + dz*dz
}

func (v Vector3[T]) ToVector2() Vector2[T] {
	return Vector2[T]{
		X: v.X,
		Y: v.Y,
	}
}

func (v Vector3[T]) String() string {
	return fmt.Sprintf("Vector3(%v, %v, %v)", v.X, v.Y, v.Z)
}

func (v Vector3[T]) Limit(limit T) Vector3[T] {
	l := v.Length()
	if l > limit {
		return v.Normalize().MultiplyScalar(limit)
	}

	return v
}
