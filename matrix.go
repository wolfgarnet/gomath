package gomath

import (
	"errors"
)

// Matrix is 3x3 matrix column matrix
type Matrix[T Number] [9]T

func (m *Matrix[T]) Identity() {
	m.Set(
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	)
}

func NewMatrix[T Number]() Matrix[T] {

	var m Matrix[T]
	m.Identity()
	return m
}

// Set sets all the elements of the matrix row by row starting at row1, column1,
// row1, column2, row1, column3 and so forth.
// Returns the pointer to this updated Matrix.
func (m *Matrix[T]) Set(n11, n12, n13, n21, n22, n23, n31, n32, n33 T) {

	m[0] = n11
	m[3] = n12
	m[6] = n13
	m[1] = n21
	m[4] = n22
	m[7] = n23
	m[2] = n31
	m[5] = n32
	m[8] = n33
}

func (m *Matrix[T]) SetPosition(x, y T) {
	m[6] = x
	m[7] = y
}

// SetDirection sets the direction of the matrix as a unit vector.
// It bases the angle of the matrix on the right vector.
func (m *Matrix[T]) SetDirection(newDirection Vector2[T]) {
	angle := Vector2[T]{X: 1}.AngleToWithSign(newDirection)
	m.Rotation(angle)
}

func (m *Matrix[T]) GetPosition() Vector2[T] {
	return Vector2[T]{X: m[6], Y: m[7]}
}

// GetDirection returns the direction of the matrix as a unit vector.
func (m *Matrix[T]) GetDirection() Vector2[T] {
	angle := m.Angle()
	return Vector2[T]{X: 1, Y: 0}.Rotate(angle)
}

// Angle returns the angle of the matrix in radians.
func (m *Matrix[T]) Angle() T {
	return Atan2(m[1], m[0])
}

func (m *Matrix[T]) Multiply(b *Matrix[T]) {
	*m = MatrixMultiply(m, b)
}

func MatrixMultiply[T Number](m *Matrix[T], b *Matrix[T]) (c Matrix[T]) {
	c.Identity()
	c[0] = m[0]*b[0] + m[3]*b[1] + m[6]*b[2]
	c[3] = m[0]*b[3] + m[3]*b[4] + m[6]*b[5]
	c[6] = m[0]*b[6] + m[3]*b[7] + m[6]*b[8]

	c[1] = m[1]*b[0] + m[4]*b[1] + m[7]*b[2]
	c[4] = m[1]*b[3] + m[4]*b[4] + m[7]*b[5]
	c[7] = m[1]*b[6] + m[4]*b[7] + m[7]*b[8]

	c[2] = m[2]*b[0] + m[5]*b[1] + m[8]*b[2]
	c[5] = m[2]*b[3] + m[5]*b[4] + m[8]*b[5]
	c[8] = m[2]*b[6] + m[5]*b[7] + m[8]*b[8]

	return c
}

func (m Matrix[T]) MultiplyScalar(s T) Matrix[T] {
	m[0] *= s
	m[3] *= s
	m[6] *= s
	m[1] *= s
	m[4] *= s
	m[7] *= s
	m[2] *= s
	m[5] *= s
	m[8] *= s

	return m
}

func (m Matrix[T]) Determinant() T {

	return m[0]*m[4]*m[8] -
		m[0]*m[5]*m[7] -
		m[1]*m[3]*m[8] +
		m[1]*m[5]*m[6] +
		m[2]*m[3]*m[7] -
		m[2]*m[4]*m[6]
}

// GetInverse sets m to the inverse of src, returning an error if src is not invertible,
// in which case m is set to the identity.
func (m *Matrix[T]) GetInverse(src Matrix[T]) error {

	n11 := src[0]
	n21 := src[1]
	n31 := src[2]
	n12 := src[3]
	n22 := src[4]
	n32 := src[5]
	n13 := src[6]
	n23 := src[7]
	n33 := src[8]

	t11 := n33*n22 - n32*n23
	t12 := n32*n13 - n33*n12
	t13 := n23*n12 - n22*n13

	det := n11*t11 + n21*t12 + n31*t13

	// no inverse
	if det == 0 {
		m.Identity()
		return errors.New("cannot invert matrix")
	}

	detInv := 1 / det

	m[0] = t11 * detInv
	m[1] = (n31*n23 - n33*n21) * detInv
	m[2] = (n32*n21 - n31*n22) * detInv
	m[3] = t12 * detInv
	m[4] = (n33*n11 - n31*n13) * detInv
	m[5] = (n31*n12 - n32*n11) * detInv
	m[6] = t13 * detInv
	m[7] = (n21*n13 - n23*n11) * detInv
	m[8] = (n22*n11 - n21*n12) * detInv

	return nil
}

func (m Matrix[T]) Transpose() Matrix[T] {
	m[1], m[3] = m[3], m[1]
	m[2], m[6] = m[6], m[2]
	m[5], m[7] = m[7], m[5]

	return m
}

// Translation set the translation part of a transformation matrix
func (m *Matrix[T]) Translation(x, y T) {
	m[6] += x
	m[7] += y
}

// Rotation sets the rotation part of a transformation matrix given in radians.
func (m *Matrix[T]) Rotation(theta T) {
	c := Cos(theta)
	s := Sin(theta)

	m[0] = c
	m[1] = s
	m[3] = -s
	m[4] = c
}
