package gomath

import (
	"fmt"
	"math"
)

// Vector4 is a 4D vector. Its primary uses are homogeneous coordinates for 3D
// transforms (a point is (x, y, z, 1), a direction is (x, y, z, 0)), clip-space
// positions produced by a projection matrix, plane equations, and RGBA values.
//
// As with Vector2 and Vector3, integer element types are storage-only for
// geometry: methods that involve magnitude (Length, Normalize, SetLength) or
// division (PerspectiveDivide) compute in float64 and truncate back to T, so
// their results are meaningless for integer T. Use a floating-point T for any
// such work. Pure element-wise operations (Add, Sub, Mul, Min/Max, ...) are
// exact for integers.
type Vector4[T Number] struct {
	X T
	Y T
	Z T
	W T
}

func NewVector4[T Number](x T, y T, z T, w T) Vector4[T] {
	return Vector4[T]{
		X: x,
		Y: y,
		Z: z,
		W: w,
	}
}

// Set returns a copy of the vector with X, Y, Z and W replaced. It has a value
// receiver and does NOT mutate the receiver, so the result must be used:
// `v = v.Set(x, y, z, w)`.
func (v Vector4[T]) Set(x T, y T, z T, w T) Vector4[T] {
	v.X = x
	v.Y = y
	v.Z = z
	v.W = w

	return v
}

// Component returns the i'th component, ordered X, Y, Z, W. It panics for an
// index outside [0, 3], matching Vector2.Component and Vector3.Component.
func (v Vector4[T]) Component(i int) T {
	switch i {
	case 0:
		return v.X
	case 1:
		return v.Y
	case 2:
		return v.Z
	case 3:
		return v.W
	}

	panic(fmt.Sprintf("gomath: Vector4 component index %d out of range [0, 3]", i))
}

func (v Vector4[T]) Add(v2 Vector4[T]) Vector4[T] {
	v.X += v2.X
	v.Y += v2.Y
	v.Z += v2.Z
	v.W += v2.W

	return v
}

func (v Vector4[T]) AddScalar(scalar T) Vector4[T] {
	v.X += scalar
	v.Y += scalar
	v.Z += scalar
	v.W += scalar

	return v
}

func (v Vector4[T]) Sub(v2 Vector4[T]) Vector4[T] {
	v.X -= v2.X
	v.Y -= v2.Y
	v.Z -= v2.Z
	v.W -= v2.W

	return v
}

func (v Vector4[T]) SubScalar(scalar T) Vector4[T] {
	v.X -= scalar
	v.Y -= scalar
	v.Z -= scalar
	v.W -= scalar

	return v
}

// Mul multiplies component-wise (the Hadamard product), not as a dot or cross
// product.
func (v Vector4[T]) Mul(v2 Vector4[T]) Vector4[T] {
	v.X *= v2.X
	v.Y *= v2.Y
	v.Z *= v2.Z
	v.W *= v2.W

	return v
}

func (v Vector4[T]) MultiplyScalar(scalar T) Vector4[T] {
	v.X *= scalar
	v.Y *= scalar
	v.Z *= scalar
	v.W *= scalar

	return v
}

func (v Vector4[T]) Div(v2 Vector4[T]) Vector4[T] {
	v.X /= v2.X
	v.Y /= v2.Y
	v.Z /= v2.Z
	v.W /= v2.W

	return v
}

func (v Vector4[T]) DivScalar(scalar T) Vector4[T] {
	v.X /= scalar
	v.Y /= scalar
	v.Z /= scalar
	v.W /= scalar

	return v
}

func (v Vector4[T]) Dot(v2 Vector4[T]) T {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z + v.W*v2.W
}

// Lerp linearly interpolates from v towards v2 by t, where t == 0 returns v and
// t == 1 returns v2. t is not clamped.
func (v Vector4[T]) Lerp(v2 Vector4[T], t T) Vector4[T] {
	v.X += (v2.X - v.X) * t
	v.Y += (v2.Y - v.Y) * t
	v.Z += (v2.Z - v.Z) * t
	v.W += (v2.W - v.W) * t

	return v
}

func (v Vector4[T]) LengthSq() T {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W
}

func (v Vector4[T]) Length64() float64 {
	return math.Sqrt(float64(v.LengthSq()))
}

func (v Vector4[T]) Length() T {
	return T(v.Length64())
}

func (v Vector4[T]) Length32() float32 {
	return float32(v.Length64())
}

// Normalize returns the vector scaled to unit length, or the zero vector if v
// has zero length.
func (v Vector4[T]) Normalize() Vector4[T] {
	length := v.Length()
	if length == 0 {
		return Vector4[T]{}
	}

	return v.DivScalar(length)
}

// ToVector3 drops the W component. For a homogeneous point produced by a
// projection use PerspectiveDivide first, which divides through by W.
func (v Vector4[T]) ToVector3() Vector3[T] {
	return Vector3[T]{
		X: v.X,
		Y: v.Y,
		Z: v.Z,
	}
}

// PerspectiveDivide divides X, Y and Z by W, converting a homogeneous clip-space
// position into normalised device coordinates, and returns the result as a
// Vector3. A W of zero (a direction, or a point on the eye plane) has no finite
// projection, so the XYZ part is returned unchanged.
func (v Vector4[T]) PerspectiveDivide() Vector3[T] {
	if v.W == 0 {
		return v.ToVector3()
	}

	return Vector3[T]{
		X: v.X / v.W,
		Y: v.Y / v.W,
		Z: v.Z / v.W,
	}
}

// To transform a Vector4 by a matrix, see Matrix4.TransformVector4. It lives on
// Matrix4 rather than here (as Vector2.ApplyMatrix does) because Matrix4 is
// constrained to T Float while Vector4 allows any T Number, and a method cannot
// narrow the constraint of its receiver.

func (v Vector4[T]) Equals(v2 Vector4[T]) bool {
	return v == v2
}

func (v Vector4[T]) String() string {
	return fmt.Sprintf("Vector4(%v, %v, %v, %v)", v.X, v.Y, v.Z, v.W)
}
