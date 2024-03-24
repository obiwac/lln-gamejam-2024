package main

import "math"

type Entity struct {
	pos         [3]float32
	rot         [2]float32
	vel         [3]float32
	acc         [3]float32
	jump_height float32
	width       float32
	height      float32
	collider    *Collider
	grounded    bool
}

type PotentialCollision struct {
	entryTime float32
	normal    [3]float32
}

var GRAVITY_ACCEL = []float32{0, -9.81, 0}
var FRICTION = []float32{20, 20, 20}
var DRAG_JUMP = []float32{1.8, 0, 1.8}
var DRAG_FALL = []float32{1.8, .4, 1.8}

func NewEntity(position [3]float32, rotation [2]float32, width float32, height float32) *Entity {
	entity := &Entity{
		pos:         position,
		rot:         rotation,
		vel:         [3]float32{0, 0, 0},
		acc:         [3]float32{0, 0, 0},
		jump_height: 1,
		width:       width,
		height:      height,
		collider:    &Collider{},
		grounded:    false,
	}

	return entity
}

func NewPotentialCollision(entryTime float32, normal [3]float32) *PotentialCollision {
	return &PotentialCollision{
		entryTime: entryTime,
		normal:    normal,
	}
}

func (entity *Entity) Update(models []*Model) {
	dt := float32(1. / 60)

	// compute friction/drag

	fx, fy, fz := DRAG_FALL[0], DRAG_FALL[1], DRAG_FALL[2]

	if entity.grounded {
		fx, fy, fz = FRICTION[0], FRICTION[1], FRICTION[2]
	} else if entity.vel[1] > 0 {
		fx, fy, fz = DRAG_JUMP[0], DRAG_JUMP[1], DRAG_JUMP[2]
	}

	// input acceleration + friction compensation

	entity.vel[0] += entity.acc[0] * fx * dt
	entity.vel[1] += entity.acc[1] * fy * dt
	entity.vel[2] += entity.acc[2] * fz * dt

	entity.acc[0], entity.acc[1], entity.acc[2] = 0, 0, 0

	// update collider

	x, y, z := entity.pos[0], entity.pos[1], entity.pos[2]

	entity.collider.position1 = [3]float32{x - entity.width/2, y, z - entity.width/2}
	entity.collider.position2 = [3]float32{x + entity.width/2, y + entity.height, z + entity.width/2}

	// collide with colliders

	entity.grounded = false

	for i := 0; i < 3; i++ {
		vx := entity.vel[0] * dt
		vy := entity.vel[1] * dt
		vz := entity.vel[2] * dt

		candidates := []PotentialCollision{}

		for _, model := range models {
			for _, collider := range model.colliders {
				collided, normals := entity.collider.Collide(&collider, vx, vy, vz)
				if collided < 1 {
					potentialCollision := NewPotentialCollision(collided, normals)
					candidates = append(candidates, *potentialCollision)
				}
			}
		}

		// get first collision

		var earliest_collision PotentialCollision
		earliest_time := float32(2)

		for _, candidate := range candidates {
			if candidate.entryTime < earliest_time {
				earliest_collision = candidate
				earliest_time = candidate.entryTime
			}
		}

		if earliest_time >= 1 {
			break
		}

		earliest_time -= .001

		if earliest_collision.normal[0] != 0 {
			entity.pos[0] += vx * earliest_time
			entity.vel[0] = 0
		}

		if earliest_collision.normal[1] != 0 {
			entity.pos[1] += vy * earliest_time
			entity.vel[1] = 0

			if earliest_collision.normal[1] > 0 {
				entity.grounded = true
			}
		}

		if earliest_collision.normal[2] != 0 {
			entity.pos[2] += vz * earliest_time
			entity.vel[2] = 0
		}
	}

	// update position

	entity.pos[0] += entity.vel[0] * dt
	entity.pos[1] += entity.vel[1] * dt
	entity.pos[2] += entity.vel[2] * dt

	// apply gravity

	entity.vel[1] += GRAVITY_ACCEL[1] * dt

	// friction

	abs_min := func(x, y float32) float32 {
		if math.Abs(float64(x)) < math.Abs(float64(y)) {
			return x
		}

		return y
	}

	entity.vel[0] -= abs_min(entity.vel[0]*fx*dt, entity.vel[0])
	entity.vel[1] -= abs_min(entity.vel[1]*fy*dt, entity.vel[1])
	entity.vel[2] -= abs_min(entity.vel[2]*fz*dt, entity.vel[2])
}

func (entity *Entity) Jump() {
	if entity.grounded {
		entity.vel[1] = float32(math.Sqrt(-2 * float64(GRAVITY_ACCEL[1]*entity.jump_height)))
	}
}
