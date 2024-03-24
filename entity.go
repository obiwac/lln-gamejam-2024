package main

import (
	"math"
)

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
	entity.position[0] += entity.velocity[0]
	entity.position[1] += entity.velocity[1]
	entity.position[2] += entity.velocity[2]

	entity.UpdateCollider(models)
}

func (entity *Entity) UpdateCollider(models []*Model) {
	x, y, z := entity.position[0], entity.position[1], entity.position[2]

	entity.collider.position1 = [3]float32{x - entity.width/2, y - entity.height/2, z - entity.width/2}
	entity.collider.position2 = [3]float32{x + entity.width/2, y + entity.height/2, z + entity.width/2}

	entity.CheckCollisions(models)
}

func (entity *Entity) CheckCollisions(models []*Model) {
	for _, model := range models {
		for _, collider := range model.colliders {
			collided, normals := collider.Collide(entity.collider, entity.velocity)
			if collided == 1 {
				potentialCollision := NewPotentialCollision(collided, normals)
				entity.potentialCollisions = append(entity.potentialCollisions, *potentialCollision)
			}
		}
	}
}

func (entity *Entity) GetFirstCollision() {
	entryTime, normal := GetMinPenetrationTime(entity.potentialCollisions)
	entryTime -= 0.0001

	if normal[0] != 0 {
		entity.position[0] += entity.velocity[0] * entryTime
		entity.velocity[0] = 0
	}

	if normal[1] != 0 {
		entity.position[1] += entity.velocity[1] * entryTime
		entity.velocity[1] = 0
	}

	if normal[2] != 0 {
		entity.position[2] += entity.velocity[2] * entryTime
		entity.velocity[2] = 0
	}
}

// 8===================================================================================================D

func GetMinPenetrationTime(potentialCollisions []PotentialCollision) (float32, [3]float32) {
	minTime := float32(math.MaxFloat32)
	var normal [3]float32

	for _, potentialCollision := range potentialCollisions {
		if potentialCollision.entryTime < minTime {
			minTime = potentialCollision.entryTime
			normal = potentialCollision.normal
		}
	}

	return minTime, normal
}
