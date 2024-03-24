package main

import (
	"math"
)

type Collider struct {
	name      string
	position1 [3]float32
	position2 [3]float32
	ignore    bool
}

func NewCollider(name string, position1 [3]float32, position2 [3]float32) *Collider {
	return &Collider{
		name:      name,
		position1: position1,
		position2: position2,
		ignore:    false,
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

func (collider *Collider) Collide(other *Collider, vx, vy, vz float32) (string, float32, [3]float32) {
	x_entry := float32(0)
	y_entry := float32(0)
	z_entry := float32(0)
	x_exit := float32(0)
	y_exit := float32(0)
	z_exit := float32(0)

	if vx > 0 {
		x_entry = time_(other.position1[0]-collider.position2[0], vx)
		x_exit = time_(other.position2[0]-collider.position1[0], vx)
	} else {
		x_entry = time_(other.position2[0]-collider.position1[0], vx)
		x_exit = time_(other.position1[0]-collider.position2[0], vx)
	}
	if vy > 0 {
		y_entry = time_(other.position1[1]-collider.position2[1], vy)
		y_exit = time_(other.position2[1]-collider.position1[1], vy)
	} else {
		y_entry = time_(other.position2[1]-collider.position1[1], vy)
		y_exit = time_(other.position1[1]-collider.position2[1], vy)
	}
	if vz > 0 {
		z_entry = time_(other.position1[2]-collider.position2[2], vz)
		z_exit = time_(other.position2[2]-collider.position1[2], vz)
	} else {
		z_entry = time_(other.position2[2]-collider.position1[2], vz)
		z_exit = time_(other.position1[2]-collider.position2[2], vz)
	}

	if x_entry < 0 && y_entry < 0 && z_entry < 0 {
		return other.name, 1.0, [3]float32{}
	}

	if x_entry > 1 || y_entry > 1 || z_entry > 1 {
		return other.name, 1.0, [3]float32{}
	}

	entry := max3(x_entry, y_entry, z_entry)
	exit := min3(x_exit, y_exit, z_exit)

	if entry > exit {
		return other.name, 1.0, [3]float32{}
	}

	nx := float32(0)
	ny := float32(0)
	nz := float32(0)

	if entry == x_entry {
		if vx > 0 {
			nx = -1
		} else {
			nx = 1
		}
	}
	if entry == y_entry {
		if vy > 0 {
			ny = -1
		} else {
			ny = 1
		}
	}
	if entry == z_entry {
		if vz > 0 {
			nz = -1
		} else {
			nz = 1
		}
	}

	return other.name, entry, [3]float32{float32(nx), float32(ny), float32(nz)}
}

func time_(x, y float32) float32 {
	if y != 0 {
		return x / y
	}

	inf := float32(99999999)

	if x > 0 {
		return -inf
	}

	return inf
}

func max3(x, y, z float32) float32 {
	if x > y && x > z {
		return x
	}

	if y > x && y > z {
		return y
	}

	return z
}

func min3(x, y, z float32) float32 {
	if x < y && x < z {
		return x
	}

	if y < x && y < z {
		return y
	}

	return z
}
