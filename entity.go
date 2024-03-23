package main

type Entity struct {
	position [3]float32
	rotation [2]float32
	velocity [3]float32
	width    float32
	height   float32
	collider *Collider
}

func NewEntity(position [3]float32, rotation [2]float32, velocity [3]float32, width float32, height float32, collider *Collider) *Entity {
	entity := &Entity{
		position: position,
		rotation: rotation,
		velocity: velocity,
		width:    width,
		height:   height,
		collider: collider,
	}

	return entity
}

func (entity *Entity) Update() {
	entity.UpdateCollider()
}

func (entity *Entity) UpdateCollider() {
	x, y, z := entity.position[0], entity.position[1], entity.position[2]

	entity.collider.position1 = [3]float32{x - entity.width/2, y - entity.height/2, z - entity.width/2}
	entity.collider.position2 = [3]float32{x + entity.width/2, y + entity.height/2, z + entity.width/2}
}
