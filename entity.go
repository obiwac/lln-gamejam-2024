package main

import "math"

type Entity struct {
	state           *State
	pos             [3]float32
	rot             [2]float32
	vel             [3]float32
	trigger_impulse [3]float32
	acc             [3]float32
	jump_height     float32
	width           float32
	height          float32
	collider        *Collider
	grounded        bool
}

type PotentialCollision struct {
	name      string
	entryTime float32
	normal    [3]float32
	collider  *Collider
}

var GRAVITY_ACCEL = []float32{0, -9.81, 0}
var FRICTION = []float32{20, 20, 20}
var DRAG_JUMP = []float32{1.8, 0, 1.8}
var DRAG_FALL = []float32{1.8, .4, 1.8}

func NewEntity(state *State, position [3]float32, rotation [2]float32, width float32, height float32) *Entity {
	entity := &Entity{
		state:       state,
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

func NewPotentialCollision(name string, entryTime float32, normal [3]float32, collider *Collider) *PotentialCollision {
	return &PotentialCollision{
		name:      name,
		entryTime: entryTime,
		normal:    normal,
		collider:  collider,
	}
}

func (entity *Entity) Update(models []*Model) {
	dt := entity.state.dt

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
	entity.trigger_impulse = [3]float32{0, 0, 0}

	for i := 0; i < 3; i++ {
		vx := entity.vel[0] * dt
		vy := entity.vel[1] * dt
		vz := entity.vel[2] * dt

		candidates := []PotentialCollision{}

		for _, model := range models {
			for j := 0; j < len(model.colliders); j++ {
				if model.colliders[j].ignore {
					continue
				}
				name, collided, normals := entity.collider.Collide(&model.colliders[j], vx, vy, vz)
				if collided < 1 {
					potentialCollision := NewPotentialCollision(name, collided, normals, &model.colliders[j])
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

		trigger := earliest_collision.name
		entity.prossesTrigger(trigger, entity.state, earliest_collision.collider)

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

	// collide with heightmaps

	for _, model := range models {
		if model.heightmap == nil {
			continue
		}

		px, py, pz := entity.pos[0], entity.pos[1], entity.pos[2]

		px -= model.collider_off_x
		py -= model.collider_off_y
		pz -= model.collider_off_z

		px /= M_TO_AYLIN
		py /= M_TO_AYLIN
		pz /= M_TO_AYLIN

		x := int(float32(model.heightmap.res) * (px - model.heightmap.neg_x) / (model.heightmap.pos_x - model.heightmap.neg_x))
		y := int(float32(model.heightmap.res) * (pz - model.heightmap.neg_z) / (model.heightmap.pos_z - model.heightmap.neg_z))

		if x < 0 || y < 0 || x >= int(model.heightmap.res) || y >= int(model.heightmap.res) {
			continue
		}

		height := model.heightmap.heightmap[x][y]

		if py < height {
			entity.pos[1] = height*M_TO_AYLIN + model.collider_off_y
			entity.vel[1] = 0
			entity.grounded = true
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

	// trigger impulse

	entity.vel[0] += entity.trigger_impulse[0]
	entity.vel[1] += entity.trigger_impulse[1]
	entity.vel[2] += entity.trigger_impulse[2]
}

func (entity *Entity) Jump() {
	if entity.grounded {
		entity.vel[1] = float32(math.Sqrt(-2 * float64(GRAVITY_ACCEL[1]*entity.jump_height)))
	}
}

var apat_already_spoken = false

func (entity *Entity) prossesTrigger(trigger string, state *State, collider *Collider) {
	if trigger == "Col_Sink" && !state.alexis_room.sink_activated {
		displayDialogue(getDialogues(), "intro2", state)
		state.alexis_room.sink_activated = true
	} else if trigger == "Col_Door" && state.alexis_room.sink_activated {
		displayDialogue(getDialogues(), "outro4", state)
		// TODO : i input
		state.alexis_room.door_opened = true
		collider.ignore = true
		entity.trigger_impulse[0] = -30
	} else if trigger == "Col_Ukulele" {
		displayDialogue(getDialogues(), "ukulele5", state)
		state.alexis_room.door_opened = true
		state.apat.ukulele_activated = true
		collider.ignore = true
	} else if trigger == "Col_Purple" && state.apat.ukulele_activated {
		displayDialogue(getDialogues(), "nether1", state)
	} else if trigger == "Col_Apat" {
		if apat_already_spoken {
			displayDialogue(getDialogues(), "bonus", state)
		} else {
			displayDialogue(getDialogues(), "ukulele2", state)
		}
	}
}
