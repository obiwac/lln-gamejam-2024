package main

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
				res[i][j] += mat.Data[i][k] * other.Data[k][j]
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
		{x, y, z, 1},
	}

	return mat
}
