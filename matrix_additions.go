package gomath

import (
	"fmt"
	"math"
)

// This file rounds Matrix out into a production-grade 2D affine transform for
// games: value-oriented builders and combinators layered on top of the existing
// (mostly pointer-mutating) API.
//
// Matrix is a column-major 3x3 matrix representing a 2D affine transform. The
// element in row r, column c is stored at index c*3+r; columns 0 and 1 are the
// linear (rotation/scale/shear) basis and column 2 (indices 6, 7) is the
// translation. Point transforms use the full affine; direction transforms use
// only the linear part.
//
// As with the vector types, integer T is storage-only for operations that go
// through float64 (RotationMatrix, Angle) or division (Inverse): use a
// floating-point T for real transform work.

// TranslationMatrix returns an affine matrix that translates by (x, y).
func TranslationMatrix[T Number](x, y T) Matrix[T] {
	return Matrix[T]{1, 0, 0, 0, 1, 0, x, y, 1}
}

// RotationMatrix returns an affine matrix that rotates counter-clockwise by
// theta radians (consistent with Vector2.Rotate).
func RotationMatrix[T Number](theta T) Matrix[T] {
	c, s := Cos(theta), Sin(theta)
	return Matrix[T]{c, s, 0, -s, c, 0, 0, 0, 1}
}

// ScaleMatrix returns an affine matrix that scales by sx along X and sy along Y.
func ScaleMatrix[T Number](sx, sy T) Matrix[T] {
	return Matrix[T]{sx, 0, 0, 0, sy, 0, 0, 0, 1}
}

// ShearMatrix returns an affine matrix that shears: x += shx*y, y += shy*x.
func ShearMatrix[T Number](shx, shy T) Matrix[T] {
	return Matrix[T]{1, shy, 0, shx, 1, 0, 0, 0, 1}
}

// TRSMatrix composes a translation, rotation and scale into a single affine
// transform T*R*S: a point is first scaled, then rotated, then translated.
func TRSMatrix[T Number](translation Vector2[T], rotation T, scale Vector2[T]) Matrix[T] {
	return TranslationMatrix(translation.X, translation.Y).
		Mul(RotationMatrix(rotation)).
		Mul(ScaleMatrix(scale.X, scale.Y))
}

// Mul returns the matrix product m*b as a new matrix, leaving both operands
// unchanged. Transforms compose right-to-left: (m*b) applies b first, then m.
func (m Matrix[T]) Mul(b Matrix[T]) Matrix[T] {
	return MatrixMultiply(&m, &b)
}

// TransformPoint applies the full affine transform (including translation) to v.
// It is equivalent to v.ApplyMatrix(m).
func (m Matrix[T]) TransformPoint(v Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: m[0]*v.X + m[3]*v.Y + m[6],
		Y: m[1]*v.X + m[4]*v.Y + m[7],
	}
}

// TransformDirection applies only the linear part of the transform (rotation,
// scale, shear) to v, ignoring translation. Use it for directions and normals.
func (m Matrix[T]) TransformDirection(v Vector2[T]) Vector2[T] {
	return Vector2[T]{
		X: m[0]*v.X + m[3]*v.Y,
		Y: m[1]*v.X + m[4]*v.Y,
	}
}

// Inverse returns the inverse of m and true, or the identity and false if m is
// singular (non-invertible).
func (m Matrix[T]) Inverse() (Matrix[T], bool) {
	var inv Matrix[T]
	err := inv.GetInverse(m)
	return inv, err == nil
}

// GetScale returns the scale magnitudes encoded in the linear part, as the
// lengths of the two basis columns. The values are always non-negative (a
// reflection is not distinguished from a positive scale).
func (m Matrix[T]) GetScale() Vector2[T] {
	return Vector2[T]{
		X: T(math.Hypot(float64(m[0]), float64(m[1]))),
		Y: T(math.Hypot(float64(m[3]), float64(m[4]))),
	}
}

// Equals reports whether every element of m equals the corresponding element
// of b exactly.
func (m Matrix[T]) Equals(b Matrix[T]) bool {
	return m == b
}

// ApproxEquals reports whether every element of m is within tol of the
// corresponding element of b.
func (m Matrix[T]) ApproxEquals(b Matrix[T], tol T) bool {
	for i := range m {
		if Abs(m[i]-b[i]) > tol {
			return false
		}
	}
	return true
}

// Translated returns m composed with a translation applied in m's local space
// (right-multiply): m.Translated(x, y) == m.Mul(TranslationMatrix(x, y)).
func (m Matrix[T]) Translated(x, y T) Matrix[T] {
	return m.Mul(TranslationMatrix(x, y))
}

// Rotated returns m composed with a rotation applied in m's local space
// (right-multiply): m.Rotated(theta) == m.Mul(RotationMatrix(theta)).
func (m Matrix[T]) Rotated(theta T) Matrix[T] {
	return m.Mul(RotationMatrix(theta))
}

// Scaled returns m composed with a scale applied in m's local space
// (right-multiply): m.Scaled(sx, sy) == m.Mul(ScaleMatrix(sx, sy)).
func (m Matrix[T]) Scaled(sx, sy T) Matrix[T] {
	return m.Mul(ScaleMatrix(sx, sy))
}

// String renders the matrix in row-major form for debugging.
func (m Matrix[T]) String() string {
	return fmt.Sprintf("Matrix[[%v %v %v] [%v %v %v] [%v %v %v]]",
		m[0], m[3], m[6],
		m[1], m[4], m[7],
		m[2], m[5], m[8])
}
