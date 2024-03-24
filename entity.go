package main

type Entity struct {
	position            [3]float32
	rotation            [2]float32
	velocity            [3]float32
	width               float32
	height              float32
	collider            *Collider
	potentialCollisions []PotentialCollision
}

type PotentialCollision struct {
	entryTime float32
	normal    [3]float32
}

func NewEntity(position [3]float32, rotation [2]float32, velocity [3]float32, width float32, height float32) *Entity {
	entity := &Entity{
		position: position,
		rotation: rotation,
		velocity: velocity,
		width:    width,
		height:   height,
		collider: &Collider{},
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

	// update collider

	x, y, z := entity.position[0], entity.position[1], entity.position[2]

	entity.collider.position1 = [3]float32{x - entity.width/2, y, z - entity.width/2}
	entity.collider.position2 = [3]float32{x + entity.width/2, y + entity.height, z + entity.width/2}

	// collide with colliders

	for i := 0; i < 3; i++ {
		vx := entity.velocity[0] * dt
		vy := entity.velocity[1] * dt
		vz := entity.velocity[2] * dt

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
			entity.position[0] += vx * earliest_time
			entity.velocity[0] = 0
		}

		if earliest_collision.normal[1] != 0 {
			entity.position[1] += vy * earliest_time
			entity.velocity[1] = 0
		}

		if earliest_collision.normal[2] != 0 {
			entity.position[2] += vz * earliest_time
			entity.velocity[2] = 0
		}
	}

	// update position

	entity.position[0] += entity.velocity[0] * dt
	entity.position[1] += entity.velocity[1] * dt
	entity.position[2] += entity.velocity[2] * dt
}
