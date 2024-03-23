package main

import (
	"math"
)

type Collider struct {
	position1 [3]float32
	position2 [3]float32
}

func NewCollider(position1 [3]float32, position2 [3]float32) *Collider {
	return &Collider{
		position1: position1,
		position2: position2,
	}
}

func (collider *Collider) AddPosition(pos [3]float32) {
	collider.position1[0] += pos[0]
	collider.position1[1] += pos[1]
	collider.position1[2] += pos[2]

	collider.position2[0] += pos[0]
	collider.position2[1] += pos[1]
	collider.position2[2] += pos[2]
}

func (collider *Collider) And(other *Collider) bool {
	x := float64(math.Min(float64(collider.position2[0]), float64(other.position2[0]))) - float64(math.Max(float64(collider.position1[0]), float64(other.position1[0])))
	y := float64(math.Min(float64(collider.position2[1]), float64(other.position2[1]))) - float64(math.Max(float64(collider.position1[1]), float64(other.position1[1])))
	z := float64(math.Min(float64(collider.position2[2]), float64(other.position2[2]))) - float64(math.Max(float64(collider.position1[2]), float64(other.position1[2])))

	return x > 0 && y > 0 && z > 0
}

func (collider *Collider) Collide(col *Collider, velocity [3]float32) (int, [3]float32) {
	x_entry := float32(0)
	y_entry := float32(0)
	z_entry := float32(0)
	x_exit := float32(0)
	y_exit := float32(0)
	z_exit := float32(0)

	if velocity[0] > 0 {
		x_entry = time_(col.position1[0]-collider.position2[0], velocity[0])
		x_exit = time_(col.position2[0]-collider.position1[0], velocity[0])
	} else {
		x_entry = time_(col.position2[0]-collider.position1[0], velocity[0])
		x_exit = time_(col.position1[0]-collider.position2[0], velocity[0])
	}
	if velocity[1] > 0 {
		y_entry = time_(col.position1[1]-collider.position2[1], velocity[1])
		y_exit = time_(col.position2[1]-collider.position1[1], velocity[1])
	} else {
		y_entry = time_(col.position2[1]-collider.position1[1], velocity[1])
		y_exit = time_(col.position1[1]-collider.position2[1], velocity[1])
	}
	if velocity[2] > 0 {
		z_entry = time_(col.position1[2]-collider.position2[2], velocity[2])
		z_exit = time_(col.position2[2]-collider.position1[2], velocity[2])
	} else {
		z_entry = time_(col.position2[2]-collider.position1[2], velocity[2])
		z_exit = time_(col.position1[2]-collider.position2[2], velocity[2])
	}

	if x_entry < 0 && y_entry < 0 && z_entry < 0 {
		return 1, [3]float32{}
	}

	if x_entry > 1 || y_entry > 1 || z_entry > 1 {
		return 1, [3]float32{}
	}

	entry := max3(x_entry, y_entry, z_entry)
	exit := min3(x_exit, y_exit, z_exit)

	if entry > exit {
		return 1, [3]float32{}
	}

	normal_x := float32(1)
	normal_y := float32(1)
	normal_z := float32(1)

	if float32(entry) == x_entry {
		if velocity[0] > 0 {
			normal_x = -1
		}
	} else {
		normal_x = 0
	}
	if float32(entry) == y_entry {
		if velocity[1] > 0 {
			normal_y = -1
		}
	} else {
		normal_y = 0
	}
	if float32(entry) == z_entry {
		if velocity[2] > 0 {
			normal_z = -1
		}
	} else {
		normal_z = 0
	}

	return 0, [3]float32{float32(normal_x), float32(normal_y), float32(normal_z)}
}

func time_(x float32, y float32) float32 {
	if y != 0 {
		return x / y
	} else {
		return float32(math.Inf(int(x)))
	}
}

func max3(x float32, y float32, z float32) float32 {
	if x > y {
		if x > z {
			return x
		} else {
			return z
		}
	} else {
		if y > z {
			return y
		} else {
			return z
		}
	}
}

func min3(x float32, y float32, z float32) float32 {
	if x < y {
		if x < z {
			return x
		} else {
			return z
		}
	} else {
		if y < z {
			return y
		} else {
			return z
		}
	}
}
