package gomath

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math"
)

type Vector32 struct {
	Vector2[float32]
}

func NewVector32(x, y float32) Vector32 {
	return Vector32{
		Vector2[float32]{x, y},
	}
}

const Epsilon = 0.00001

type Number interface {
	constraints.Integer | constraints.Float
}

type Vector2[T Number] struct {
	X T
	Y T
}

func New[T Number](x T, y T) Vector2[T] {
	return Vector2[T]{
		X: x,
		Y: y,
	}
}

func (v Vector2[T]) Set(x T, y T) Vector2[T] {
	v.X = x
	v.Y = y

	return v
}

func (v Vector2[T]) Add(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: v.X + v2.X,
		Y: v.Y + v2.Y,
	}
}

func (v Vector2[T]) AddScalar(scalar T) Vector2[T] {
	return Vector2[T]{
		X: v.X + scalar,
		Y: v.Y + scalar,
	}
}

func (v Vector2[T]) AddScalars(x T, y T) Vector2[T] {
	return Vector2[T]{
		X: v.X + x,
		Y: v.Y + y,
	}
}

func (v Vector2[T]) Sub(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: v.X - v2.X,
		Y: v.Y - v2.Y,
	}
}

func (v Vector2[T]) SubScalar(scalar T) Vector2[T] {
	return Vector2[T]{
		X: v.X - scalar,
		Y: v.Y - scalar,
	}
}

func (v Vector2[T]) SubScalars(x T, y T) Vector2[T] {
	return Vector2[T]{
		X: v.X - x,
		Y: v.Y - y,
	}
}

func (v Vector2[T]) Mul(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: v.X * v2.X,
		Y: v.Y * v2.Y,
	}
}

func (v Vector2[T]) MultiplyScalar(scalar T) Vector2[T] {
	return Vector2[T]{
		X: v.X * scalar,
		Y: v.Y * scalar,
	}
}

func (v Vector2[T]) MultiplyScalars(x T, y T) Vector2[T] {
	return Vector2[T]{
		X: v.X * x,
		Y: v.Y * y,
	}
}

func (v Vector2[T]) Div(v2 Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: v.X / v2.X,
		Y: v.Y / v2.Y,
	}
}

func (v Vector2[T]) DivScalar(scalar float64) Vector2[T] {
	return Vector2[T]{X: v.X / T(scalar), Y: v.Y / T(scalar)}
}

func (v Vector2[T]) DivScalars(x T, y T) Vector2[T] {
	return Vector2[T]{
		X: v.X / x,
		Y: v.Y / y,
	}
}

func (v Vector2[T]) Distance(v2 Vector2[T]) float64 {
	dx := v.X - v2.X
	dy := v.Y - v2.Y
	return math.Sqrt(float64(dx*dx + dy*dy))
}

func (v Vector2[T]) Dot(v2 Vector2[T]) T {
	return v.X*v2.X + v.Y*v2.Y
}

func (v Vector2[T]) Lerp(v2 *Vector2[T], t T) Vector2[T] {
	return Vector2[T]{
		v.X + (v2.X-v.X)*t,
		v.Y + (v2.Y-v.Y)*t,
	}
}

func (v Vector2[T]) Length() float64 {
	return math.Sqrt(float64(v.X)*float64(v.X) + float64(v.Y)*float64(v.Y))
}

func (v Vector2[T]) Normalize() Vector2[T] {
	m := v.Length()

	if m > Epsilon {
		return v.DivScalar(m)
	} else {
		return v
	}
}

func (v Vector2[T]) Reflect(v2 Vector2[T]) Vector2[T] {
	factor := -T(2.0) * v.Dot(v2)
	return New(
		factor*v.X+v2.X,
		factor*v.Y+v2.Y,
	)
}

func (v Vector2[T]) Equals(v2 Vector2[T]) bool {
	return v.X == v2.X && v.Y == v2.Y
}

func (v Vector2[T]) AngleTo(v2 Vector2[T]) float32 {
	theta := float64(v.Dot(v2)) / v.Length() * v2.Length()
	return float32(math.Acos(Clamp[float64](theta, -1, 1)))
}

func (v Vector2[T]) String() string {
	return fmt.Sprintf("Vector2(%v, %v)", v.X, v.Y)
}
