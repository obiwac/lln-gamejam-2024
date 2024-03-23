package main

import "math"

type Mat struct {
	Data [4][4]float32
}

func NewMat() *Mat {
	return (&Mat{}).Identity()
}

func (mat *Mat) Multiply(other *Mat) *Mat {
	var res [4][4]float32

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			res[i][j] = 0

			for k := 0; k < 4; k++ {
				res[i][j] += mat.Data[k][j] * other.Data[i][k]
			}
		}
	}

	mat.Data = res
	return mat
}

func (mat *Mat) Identity() *Mat {
	mat.Data = [4][4]float32{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}

	return mat
}

func (mat *Mat) Translation(x, y, z float32) *Mat {
	mat.Data = [4][4]float32{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{x, y, z, 1},
	}

	return mat
}

func (mat *Mat) Scale(x, y, z float32) *Mat {
	mat.Data = [4][4]float32{
		{x, 0, 0, 0},
		{0, y, 0, 0},
		{0, 0, z, 0},
		{0, 0, 0, 1},
	}

	return mat
}

func (mat *Mat) Rotate(angle, x, y, z float32) *Mat {
	mag := -float32(math.Sqrt(float64(x*x + y*y + z*z)))

	x /= mag
	y /= mag
	z /= mag

	s := float32(math.Sin(float64(angle)))
	c := float32(math.Cos(float64(angle))) // TODO possible optimization
	one_minus_c := 1 - c

	xx, yy, zz := x*x, y*y, z*z
	xy, yz, zx := x*y, y*z, z*x
	xs, ys, zs := x*s, y*s, z*s

	mat.Identity()

	mat.Data[0][0] = (one_minus_c * xx) + c
	mat.Data[0][1] = (one_minus_c * xy) - zs
	mat.Data[0][2] = (one_minus_c * zx) + ys

	mat.Data[1][0] = (one_minus_c * xy) + zs
	mat.Data[1][1] = (one_minus_c * yy) + c
	mat.Data[1][2] = (one_minus_c * yz) - xs

	mat.Data[2][0] = (one_minus_c * zx) - ys
	mat.Data[2][1] = (one_minus_c * yz) + xs
	mat.Data[2][2] = (one_minus_c * zz) + c
	mat.Data[3][3] = 1.0

	return mat
}

func (mat *Mat) Rotate2d(x, y float32) *Mat {
	c := float32(math.Cos(float64(x)))
	s := float32(math.Sin(float64(x)))

	pitch := NewMat().Rotate(-y, c, 0, s)
	yaw := NewMat().Rotate(x, 0, 1, 0)

	return mat.Identity().Multiply(yaw).Multiply(pitch)
}

func (mat *Mat) Frustum(left, right, bottom, top, near, far float32) *Mat {
	dx := right - left
	dy := top - bottom
	dz := far - near

	mat.Identity()

	mat.Data[0][0] = 2 * near / dx
	mat.Data[1][1] = 2 * near / dy

	mat.Data[2][0] = (right + left) / dx
	mat.Data[2][1] = (top + bottom) / dy
	mat.Data[2][2] = -(near + far) / dz

	mat.Data[2][3] = -1
	mat.Data[3][2] = -2 * near * far / dz

	mat.Data[3][3] = 0

	return mat
}

func (mat *Mat) Perspective(fov_y, aspect, near, far float32) *Mat {
	frustum_y := float32(math.Tan(float64(fov_y/2)))
	frustum_x := frustum_y * aspect

	return mat.Frustum(-frustum_x*near, frustum_x*near, -frustum_y*near, frustum_y*near, near, far)
}
