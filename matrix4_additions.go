package gomath

// This file rounds Matrix4 out with the value-oriented builders and combinators
// a 3D renderer needs: translation/rotation/scale, the view matrix (LookAtMatrix4)
// and the three projection matrices.
//
// Conventions, stated once and relied on throughout:
//
//   - Right-handed, Y up. The camera looks down its own -Z axis.
//   - Rotations are counter-clockwise when looking down the axis towards the
//     origin, matching Vector3.RotateAxis and Matrix's 2D RotationMatrix.
//   - The projections map the view volume into OpenGL clip space, where the near
//     plane becomes z=-1 and the far plane z=+1 after the perspective divide.
//   - Composition is right-to-left: a.Mul(b) applies b first.

// TranslationMatrix4 returns a transform that translates by (x, y, z).
func TranslationMatrix4[T Float](x, y, z T) Matrix4[T] {
	return Matrix4[T]{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

// ScaleMatrix4 returns a transform that scales by sx, sy and sz about the origin.
func ScaleMatrix4[T Float](sx, sy, sz T) Matrix4[T] {
	return Matrix4[T]{
		sx, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	}
}

// RotationMatrix4X returns a rotation of theta radians about the X axis.
func RotationMatrix4X[T Float](theta T) Matrix4[T] {
	c, s := Cos(theta), Sin(theta)

	return Matrix4[T]{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

// RotationMatrix4Y returns a rotation of theta radians about the Y axis.
func RotationMatrix4Y[T Float](theta T) Matrix4[T] {
	c, s := Cos(theta), Sin(theta)

	return Matrix4[T]{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

// RotationMatrix4Z returns a rotation of theta radians about the Z axis. For the
// XY plane this agrees with the 2D RotationMatrix.
func RotationMatrix4Z[T Float](theta T) Matrix4[T] {
	c, s := Cos(theta), Sin(theta)

	return Matrix4[T]{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// RotationMatrix4Axis returns a rotation of theta radians about an arbitrary
// axis through the origin, via Rodrigues' rotation formula. The axis is
// normalised first, so it need not be of unit length; a zero-length axis yields
// the identity. The result agrees with Vector3.RotateAxis.
func RotationMatrix4Axis[T Float](axis Vector3[T], theta T) Matrix4[T] {
	axis = axis.Normalize()
	if axis.IsZero() {
		return Identity4[T]()
	}

	c, s := Cos(theta), Sin(theta)
	t := 1 - c
	x, y, z := axis.X, axis.Y, axis.Z

	return Matrix4[T]{
		t*x*x + c, t*x*y + s*z, t*x*z - s*y, 0,
		t*x*y - s*z, t*y*y + c, t*y*z + s*x, 0,
		t*x*z + s*y, t*y*z - s*x, t*z*z + c, 0,
		0, 0, 0, 1,
	}
}

// TRSMatrix4 composes a translation, a rotation and a scale into a single
// transform T*R*S: a point is first scaled, then rotated, then translated. The
// rotation is supplied as a matrix; a quaternion overload arrives with
// Quaternion in the next phase.
func TRSMatrix4[T Float](translation Vector3[T], rotation Matrix4[T], scale Vector3[T]) Matrix4[T] {
	return TranslationMatrix4(translation.X, translation.Y, translation.Z).
		Mul(rotation).
		Mul(ScaleMatrix4(scale.X, scale.Y, scale.Z))
}

// ComposeMatrix4 builds the transform T*R*S from a translation, a rotation and a
// scale: a point is first scaled, then rotated, then translated. It is the
// quaternion counterpart of TRSMatrix4 and the form an ECS system wants, since
// entity components naturally store position, rotation and scale separately.
//
// It is written out rather than composed from three matrix products because
// this is the hottest builder in a renderer — one call per visible object per
// frame — and the expanded form does a third of the arithmetic.
func ComposeMatrix4[T Float](translation Vector3[T], rotation Quaternion[T], scale Vector3[T]) Matrix4[T] {
	x2, y2, z2 := rotation.X+rotation.X, rotation.Y+rotation.Y, rotation.Z+rotation.Z
	xx, xy, xz := rotation.X*x2, rotation.X*y2, rotation.X*z2
	yy, yz, zz := rotation.Y*y2, rotation.Y*z2, rotation.Z*z2
	wx, wy, wz := rotation.W*x2, rotation.W*y2, rotation.W*z2

	return Matrix4[T]{
		(1 - (yy + zz)) * scale.X, (xy + wz) * scale.X, (xz - wy) * scale.X, 0,
		(xy - wz) * scale.Y, (1 - (xx + zz)) * scale.Y, (yz + wx) * scale.Y, 0,
		(xz + wy) * scale.Z, (yz - wx) * scale.Z, (1 - (xx + yy)) * scale.Z, 0,
		translation.X, translation.Y, translation.Z, 1,
	}
}

// Decompose splits an affine transform back into the translation, rotation and
// scale that ComposeMatrix4 would rebuild it from.
//
// It is exact only for transforms that are genuinely a composition of those
// three; shear cannot be represented and is silently absorbed into the rotation.
// A negative determinant means the transform includes a reflection, which no
// rotation-plus-positive-scale can express, so the sign is folded into scale.X
// by convention.
//
// The second result is false when the basis is singular (an axis scaled to
// zero), in which case the rotation is unrecoverable and the identity is
// returned for it.
func (m *Matrix4[T]) Decompose() (Vector3[T], Quaternion[T], Vector3[T], bool) {
	translation := m.GetPosition()
	scale := m.GetScale()

	// A reflection shows up as a negative determinant; attribute it to X.
	if m.Determinant() < 0 {
		scale.X = -scale.X
	}

	if scale.X == 0 || scale.Y == 0 || scale.Z == 0 {
		return translation, IdentityQuaternion[T](), scale, false
	}

	// Divide the scale out of the basis columns to leave a pure rotation.
	rotationMatrix := Matrix4[T]{
		m[0] / scale.X, m[1] / scale.X, m[2] / scale.X, 0,
		m[4] / scale.Y, m[5] / scale.Y, m[6] / scale.Y, 0,
		m[8] / scale.Z, m[9] / scale.Z, m[10] / scale.Z, 0,
		0, 0, 0, 1,
	}

	return translation, QuaternionFromRotationMatrix(rotationMatrix), scale, true
}

// LookAtMatrix4 returns a view matrix that places the camera at eye looking
// towards target, with up defining the roll. It is the inverse of the camera's
// world transform: it maps world space into a right-handed eye space where the
// camera sits at the origin looking down -Z.
//
// If the view direction is parallel to up the basis is degenerate and cannot be
// resolved; the identity is returned in that case rather than a matrix full of
// NaNs, so a mis-specified camera fails visibly instead of poisoning the frame.
func LookAtMatrix4[T Float](eye, target, up Vector3[T]) Matrix4[T] {
	// zAxis points from the target back towards the eye, i.e. the camera's +Z.
	zAxis := eye.Sub(target).Normalize()
	if zAxis.IsZero() {
		return Identity4[T]()
	}

	xAxis := up.Cross(zAxis).Normalize()
	if xAxis.IsZero() {
		return Identity4[T]()
	}

	yAxis := zAxis.Cross(xAxis)

	return Matrix4[T]{
		xAxis.X, yAxis.X, zAxis.X, 0,
		xAxis.Y, yAxis.Y, zAxis.Y, 0,
		xAxis.Z, yAxis.Z, zAxis.Z, 0,
		-xAxis.Dot(eye), -yAxis.Dot(eye), -zAxis.Dot(eye), 1,
	}
}

// PerspectiveMatrix4 returns a symmetric perspective projection. fovY is the
// full vertical field of view in radians, aspect is width/height, and near and
// far are positive distances in front of the camera.
func PerspectiveMatrix4[T Float](fovY, aspect, near, far T) Matrix4[T] {
	f := 1 / Tan(fovY/2)

	return Matrix4[T]{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (far + near) / (near - far), -1,
		0, 0, (2 * far * near) / (near - far), 0,
	}
}

// FrustumMatrix4 returns a general (possibly asymmetric) perspective projection
// from the six planes of the view frustum, measured on the near plane. It is the
// building block PerspectiveMatrix4 specialises, and is what an off-centre
// projection needs.
func FrustumMatrix4[T Float](left, right, bottom, top, near, far T) Matrix4[T] {
	return Matrix4[T]{
		(2 * near) / (right - left), 0, 0, 0,
		0, (2 * near) / (top - bottom), 0, 0,
		(right + left) / (right - left), (top + bottom) / (top - bottom), -(far + near) / (far - near), -1,
		0, 0, -(2 * far * near) / (far - near), 0,
	}
}

// OrthographicMatrix4 returns an orthographic projection of the given box. Unlike
// the perspective builders it has no vanishing point, so parallel lines stay
// parallel; near and far may be negative, since there is no eye point to be in
// front of.
func OrthographicMatrix4[T Float](left, right, bottom, top, near, far T) Matrix4[T] {
	return Matrix4[T]{
		2 / (right - left), 0, 0, 0,
		0, 2 / (top - bottom), 0, 0,
		0, 0, -2 / (far - near), 0,
		-(right + left) / (right - left), -(top + bottom) / (top - bottom), -(far + near) / (far - near), 1,
	}
}

// Translated returns m composed with a translation applied in m's local space
// (right-multiply): m.Translated(x, y, z) == m.Mul(TranslationMatrix4(x, y, z)).
func (m Matrix4[T]) Translated(x, y, z T) Matrix4[T] {
	return m.Mul(TranslationMatrix4(x, y, z))
}

// Scaled returns m composed with a scale applied in m's local space
// (right-multiply): m.Scaled(sx, sy, sz) == m.Mul(ScaleMatrix4(sx, sy, sz)).
func (m Matrix4[T]) Scaled(sx, sy, sz T) Matrix4[T] {
	return m.Mul(ScaleMatrix4(sx, sy, sz))
}

// RotatedAxis returns m composed with a rotation applied in m's local space
// (right-multiply).
func (m Matrix4[T]) RotatedAxis(axis Vector3[T], theta T) Matrix4[T] {
	return m.Mul(RotationMatrix4Axis(axis, theta))
}
