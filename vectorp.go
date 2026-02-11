package gomath

type Vector2p[T Number] struct {
	X T
	Y T
}

func (v *Vector2p[T]) Set(x T, y T) {
	v.X = x
	v.Y = y
}

func (v *Vector2p[T]) Add(v2 *Vector2p[T]) {
	v.X = v.X + v2.X
	v.Y = v.Y + v2.Y
}

func (v *Vector2p[T]) Add2(v2 *Vector2p[T]) *Vector2p[T] {
	return &Vector2p[T]{
		X: v.X + v2.X,
		Y: v.Y + v2.Y,
	}
}

func (v *Vector2p[T]) AddScalar(scalar T) {
	v.X = v.X + scalar
	v.Y = v.Y + scalar
}

func (v *Vector2p[T]) Mul(v2 *Vector2p[T]) {
	v.X = v.X * v2.X
	v.Y = v.Y * v2.Y
}

func (v *Vector2p[T]) Mul2(v2 *Vector2p[T]) *Vector2p[T] {
	return &Vector2p[T]{
		X: v.X * v2.X,
		Y: v.Y * v2.Y,
	}
}
