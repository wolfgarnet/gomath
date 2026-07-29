package gomath

import (
	"fmt"
	"math"
)

// Quaternion represents a rotation in 3D as (X, Y, Z, W) with the scalar part
// last, matching the convention used by g3n, three.js and glTF.
//
// A rotation quaternion is a unit quaternion. Most methods here assume that and
// say so; Normalize restores it after arithmetic has drifted. Note that q and -q
// describe the *same* rotation, which is why Slerp negates one operand to take
// the shortest path and why Equals is rarely what you want — compare the
// rotations, not the components.
//
// Quaternions exist here for three reasons matrices cannot cover well: they
// interpolate correctly (Slerp), they compose without accumulating shear, and
// they cannot hit gimbal lock the way Euler angles do.
//
// Unlike Matrix3 and Matrix4, Quaternion is a 4-field struct — 32 bytes at
// float64 — so it passes in registers and keeps VALUE receivers throughout. The
// pointer-receiver carve-out documented on Matrix4 applies only to the large
// array types.
type Quaternion[T Float] struct {
	X T
	Y T
	Z T
	W T
}

// NewQuaternion returns the quaternion with the given components. It does not
// normalise; use FromAxisAngle or Normalize for a rotation.
func NewQuaternion[T Float](x T, y T, z T, w T) Quaternion[T] {
	return Quaternion[T]{X: x, Y: y, Z: z, W: w}
}

// IdentityQuaternion returns the rotation that does nothing.
func IdentityQuaternion[T Float]() Quaternion[T] {
	return Quaternion[T]{X: 0, Y: 0, Z: 0, W: 1}
}

// QuaternionFromAxisAngle returns the rotation of theta radians about the given
// axis, counter-clockwise looking down the axis towards the origin — the same
// handedness as Vector3.RotateAxis and RotationMatrix4Axis. The axis is
// normalised first; a zero-length axis yields the identity.
func QuaternionFromAxisAngle[T Float](axis Vector3[T], theta T) Quaternion[T] {
	axis = axis.Normalize()
	if axis.IsZero() {
		return IdentityQuaternion[T]()
	}

	half := theta / 2
	sin := Sin(half)

	return Quaternion[T]{
		X: axis.X * sin,
		Y: axis.Y * sin,
		Z: axis.Z * sin,
		W: Cos(half),
	}
}

// QuaternionFromEulerXYZ returns the rotation obtained by rotating about X, then
// Y, then Z. It is exactly equivalent to
//
//	RotationMatrix4X(x).Mul(RotationMatrix4Y(y)).Mul(RotationMatrix4Z(z))
//
// The order is in the name deliberately: "Euler angles" alone is ambiguous
// across twelve conventions, and silently picking one is how rotations end up
// subtly wrong.
func QuaternionFromEulerXYZ[T Float](x T, y T, z T) Quaternion[T] {
	qx := QuaternionFromAxisAngle(Vector3[T]{X: 1}, x)
	qy := QuaternionFromAxisAngle(Vector3[T]{Y: 1}, y)
	qz := QuaternionFromAxisAngle(Vector3[T]{Z: 1}, z)

	return qx.Mul(qy).Mul(qz)
}

// QuaternionFromUnitVectors returns the shortest rotation taking from to to.
// Both are normalised first. If they are opposite there is no unique shortest
// rotation, so an arbitrary perpendicular axis is chosen.
func QuaternionFromUnitVectors[T Float](from Vector3[T], to Vector3[T]) Quaternion[T] {
	from = from.Normalize()
	to = to.Normalize()
	if from.IsZero() || to.IsZero() {
		return IdentityQuaternion[T]()
	}

	w := from.Dot(to) + 1
	if w < T(Epsilon) {
		// The vectors are opposite: rotate by pi about any perpendicular axis.
		axis := Vector3[T]{X: 1}.Cross(from)
		if axis.LengthSq() < T(Epsilon) {
			axis = Vector3[T]{Y: 1}.Cross(from)
		}
		axis = axis.Normalize()

		return Quaternion[T]{X: axis.X, Y: axis.Y, Z: axis.Z, W: 0}
	}

	axis := from.Cross(to)

	return Quaternion[T]{X: axis.X, Y: axis.Y, Z: axis.Z, W: w}.Normalize()
}

// QuaternionFromRotationMatrix extracts the rotation from the upper-left 3x3 of
// m, which is assumed to be orthonormal — decompose the scale out first if it
// has any (see Matrix4.Decompose). It uses Shepperd's method, branching on the
// largest diagonal term so the division is never by a near-zero value.
func QuaternionFromRotationMatrix[T Float](m Matrix4[T]) Quaternion[T] {
	// Row r, column c lives at m[c*4+r].
	m00, m01, m02 := m[0], m[4], m[8]
	m10, m11, m12 := m[1], m[5], m[9]
	m20, m21, m22 := m[2], m[6], m[10]

	trace := m00 + m11 + m22

	switch {
	case trace > 0:
		s := T(0.5) / Sqrt(trace+1)

		return Quaternion[T]{
			X: (m21 - m12) * s,
			Y: (m02 - m20) * s,
			Z: (m10 - m01) * s,
			W: T(0.25) / s,
		}

	case m00 > m11 && m00 > m22:
		s := 2 * Sqrt(1+m00-m11-m22)

		return Quaternion[T]{
			X: T(0.25) * s,
			Y: (m01 + m10) / s,
			Z: (m02 + m20) / s,
			W: (m21 - m12) / s,
		}

	case m11 > m22:
		s := 2 * Sqrt(1+m11-m00-m22)

		return Quaternion[T]{
			X: (m01 + m10) / s,
			Y: T(0.25) * s,
			Z: (m12 + m21) / s,
			W: (m02 - m20) / s,
		}

	default:
		s := 2 * Sqrt(1+m22-m00-m11)

		return Quaternion[T]{
			X: (m02 + m20) / s,
			Y: (m12 + m21) / s,
			Z: T(0.25) * s,
			W: (m10 - m01) / s,
		}
	}
}

// Mul returns the composed rotation q*b. Like Matrix4.Mul it composes
// right-to-left: the result applies b first, then q. Quaternion multiplication
// is not commutative.
func (q Quaternion[T]) Mul(b Quaternion[T]) Quaternion[T] {
	return Quaternion[T]{
		X: q.W*b.X + q.X*b.W + q.Y*b.Z - q.Z*b.Y,
		Y: q.W*b.Y - q.X*b.Z + q.Y*b.W + q.Z*b.X,
		Z: q.W*b.Z + q.X*b.Y - q.Y*b.X + q.Z*b.W,
		W: q.W*b.W - q.X*b.X - q.Y*b.Y - q.Z*b.Z,
	}
}

// MultiplyScalar scales every component. The result is no longer a unit
// quaternion and so no longer a rotation; Slerp uses it internally.
func (q Quaternion[T]) MultiplyScalar(s T) Quaternion[T] {
	return Quaternion[T]{X: q.X * s, Y: q.Y * s, Z: q.Z * s, W: q.W * s}
}

// Add sums componentwise. As with MultiplyScalar the result is not generally a
// rotation.
func (q Quaternion[T]) Add(b Quaternion[T]) Quaternion[T] {
	return Quaternion[T]{X: q.X + b.X, Y: q.Y + b.Y, Z: q.Z + b.Z, W: q.W + b.W}
}

// Dot is the componentwise dot product. For unit quaternions its magnitude is
// the cosine of half the angle between the two rotations, which is what Slerp
// keys off.
func (q Quaternion[T]) Dot(b Quaternion[T]) T {
	return q.X*b.X + q.Y*b.Y + q.Z*b.Z + q.W*b.W
}

func (q Quaternion[T]) LengthSq() T {
	return q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W
}

func (q Quaternion[T]) Length() T {
	return T(math.Sqrt(float64(q.LengthSq())))
}

// Normalize scales the quaternion to unit length, making it a valid rotation.
// A zero quaternion has no direction, so the identity is returned instead.
func (q Quaternion[T]) Normalize() Quaternion[T] {
	length := q.Length()
	if length == 0 {
		return IdentityQuaternion[T]()
	}

	return q.MultiplyScalar(1 / length)
}

// Conjugate negates the vector part, which reverses the rotation for a unit
// quaternion.
func (q Quaternion[T]) Conjugate() Quaternion[T] {
	return Quaternion[T]{X: -q.X, Y: -q.Y, Z: -q.Z, W: q.W}
}

// Inverse returns the rotation that undoes q. For a unit quaternion this equals
// Conjugate, which is cheaper; Inverse also handles the non-unit case. A zero
// quaternion has no inverse, so the identity is returned.
func (q Quaternion[T]) Inverse() Quaternion[T] {
	lengthSq := q.LengthSq()
	if lengthSq == 0 {
		return IdentityQuaternion[T]()
	}

	return q.Conjugate().MultiplyScalar(1 / lengthSq)
}

// RotateVector3 applies the rotation to v, assuming q is a unit quaternion. It
// uses the expanded form v + w*t + cross(xyz, t) with t = 2*cross(xyz, v),
// which avoids building a matrix for a single vector.
func (q Quaternion[T]) RotateVector3(v Vector3[T]) Vector3[T] {
	vector := Vector3[T]{X: q.X, Y: q.Y, Z: q.Z}

	t := vector.Cross(v).MultiplyScalar(2)

	return v.Add(t.MultiplyScalar(q.W)).Add(vector.Cross(t))
}

// ToMatrix4 converts the rotation to a 4x4 transform with no translation or
// scale. q is assumed to be a unit quaternion.
func (q Quaternion[T]) ToMatrix4() Matrix4[T] {
	x2, y2, z2 := q.X+q.X, q.Y+q.Y, q.Z+q.Z
	xx, xy, xz := q.X*x2, q.X*y2, q.X*z2
	yy, yz, zz := q.Y*y2, q.Y*z2, q.Z*z2
	wx, wy, wz := q.W*x2, q.W*y2, q.W*z2

	return Matrix4[T]{
		1 - (yy + zz), xy + wz, xz - wy, 0,
		xy - wz, 1 - (xx + zz), yz + wx, 0,
		xz + wy, yz - wx, 1 - (xx + yy), 0,
		0, 0, 0, 1,
	}
}

// ToMatrix3 converts the rotation to a 3x3 linear basis. q is assumed to be a
// unit quaternion. Since a rotation basis is orthonormal, this doubles as its
// own normal matrix.
func (q Quaternion[T]) ToMatrix3() Matrix3[T] {
	x2, y2, z2 := q.X+q.X, q.Y+q.Y, q.Z+q.Z
	xx, xy, xz := q.X*x2, q.X*y2, q.X*z2
	yy, yz, zz := q.Y*y2, q.Y*z2, q.Z*z2
	wx, wy, wz := q.W*x2, q.W*y2, q.W*z2

	return Matrix3[T]{
		1 - (yy + zz), xy + wz, xz - wy,
		xy - wz, 1 - (xx + zz), yz + wx,
		xz + wy, yz - wx, 1 - (xx + yy),
	}
}

// ToAxisAngle returns the axis and angle the rotation represents, with the angle
// in [0, pi]. For the identity the angle is zero and the axis is arbitrary; X is
// returned so the result is always usable.
func (q Quaternion[T]) ToAxisAngle() (Vector3[T], T) {
	q = q.Normalize()

	// Clamp guards acos against a w nudged just outside [-1, 1] by rounding.
	w := Clamp(q.W, -1, 1)
	angle := 2 * Acos(w)

	sin := Sqrt(1 - w*w)
	if sin < T(Epsilon) {
		return Vector3[T]{X: 1}, 0
	}

	return Vector3[T]{X: q.X / sin, Y: q.Y / sin, Z: q.Z / sin}, angle
}

// Slerp interpolates along the shortest arc from q to b, where t == 0 returns q
// and t == 1 returns b, both exactly. Because q and -q are the same rotation,
// b is negated when the two point away from each other so the path taken is the
// short way round rather than the long one.
//
// This is the reason quaternions are here at all: interpolating the elements of
// two rotation matrices does not produce a rotation, it produces a squashed
// mess.
func (q Quaternion[T]) Slerp(b Quaternion[T], t T) Quaternion[T] {
	// Exact at the ends, so callers can rely on the endpoints being untouched
	// rather than merely very close.
	if t <= 0 {
		return q
	}
	if t >= 1 {
		return b
	}

	cosHalfTheta := q.Dot(b)

	// Take the shortest path.
	if cosHalfTheta < 0 {
		b = b.MultiplyScalar(-1)
		cosHalfTheta = -cosHalfTheta
	}

	// Nearly identical rotations: the arc is degenerate, so blend linearly and
	// renormalise rather than dividing by a vanishing sine.
	if cosHalfTheta > 1-T(Epsilon) {
		return q.MultiplyScalar(1 - t).Add(b.MultiplyScalar(t)).Normalize()
	}

	halfTheta := Acos(Clamp(cosHalfTheta, -1, 1))
	sinHalfTheta := Sqrt(1 - cosHalfTheta*cosHalfTheta)

	ratioQ := Sin((1-t)*halfTheta) / sinHalfTheta
	ratioB := Sin(t*halfTheta) / sinHalfTheta

	return q.MultiplyScalar(ratioQ).Add(b.MultiplyScalar(ratioB))
}

// AngleTo returns the angle in radians between the two rotations, in [0, pi].
func (q Quaternion[T]) AngleTo(b Quaternion[T]) T {
	dot := Abs(q.Normalize().Dot(b.Normalize()))

	return 2 * Acos(Clamp(dot, -1, 1))
}

// Equals reports exact componentwise equality. Remember that q and -q are the
// same rotation but are not Equal; use ApproxEqualRotation to compare rotations.
func (q Quaternion[T]) Equals(b Quaternion[T]) bool {
	return q == b
}

// ApproxEquals reports whether every component is within tol of b's.
func (q Quaternion[T]) ApproxEquals(b Quaternion[T], tol T) bool {
	return Abs(q.X-b.X) <= tol &&
		Abs(q.Y-b.Y) <= tol &&
		Abs(q.Z-b.Z) <= tol &&
		Abs(q.W-b.W) <= tol
}

// ApproxEqualRotation reports whether q and b describe the same rotation to
// within tol, treating q and -q as equal. This is almost always the comparison
// you want.
func (q Quaternion[T]) ApproxEqualRotation(b Quaternion[T], tol T) bool {
	return q.ApproxEquals(b, tol) || q.ApproxEquals(b.MultiplyScalar(-1), tol)
}

func (q Quaternion[T]) String() string {
	return fmt.Sprintf("Quaternion(%v, %v, %v, %v)", q.X, q.Y, q.Z, q.W)
}
