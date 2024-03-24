package main

import (
	_ "embed"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type WorldAlexisRoom struct {
	World

	room *Model
	door *Model

	door_opened bool
	door_angle  float32

	sink_activated bool
}

//go:embed res/alexis-room-lightmap.png
var alexis_room_lightmap []byte

//go:embed res/alexis-room.ivx
var alexis_room []byte

//go:embed res/alexis-door.ivx
var alexis_door []byte

func NewWorldAlexisRoom(state *State) (*WorldAlexisRoom, error) {
	room := &WorldAlexisRoom{}
	room.World = World{state: state}

	room.sink_activated = false

	var err error

	if room.room, err = NewModelFromIvx(state, "Alexis room", alexis_room, alexis_room_lightmap, false); err != nil {
		return nil, err
	}

	if room.door, err = NewModelFromIvx(state, "Alexis door", alexis_door, alexis_room_lightmap, false); err != nil {
		room.room.Release()
		return nil, err
	}

	return room, nil
}

var DOOR_ORIGIN = [3]float32{2.856, 2.4643, 0.8}

func (world *WorldAlexisRoom) Render() {
	world.state.player.mvp(NewMat())

	world.state.render_pass_manager.Begin(wgpu.LoadOp_Load, wgpu.LoadOp_Load)
	render_pass := world.state.render_pass_manager.render_pass
	world.room.Draw(render_pass)
	world.state.render_pass_manager.End()

	target_door_angle := float32(0)

	if world.door_opened {
		target_door_angle = -3.14 / 5 * 3
	}

	world.door_angle += (target_door_angle - world.door_angle) * world.state.dt * 3

	door_mat := NewMat()
	door_mat.Multiply(NewMat().Translation(DOOR_ORIGIN[0]*M_TO_AYLIN, DOOR_ORIGIN[1]*M_TO_AYLIN, DOOR_ORIGIN[2]*M_TO_AYLIN))
	door_mat.Multiply(NewMat().Rotate(world.door_angle, 0, 0, 1))
	door_mat.Multiply(NewMat().Translation(-DOOR_ORIGIN[0]*M_TO_AYLIN, -DOOR_ORIGIN[1]*M_TO_AYLIN, -DOOR_ORIGIN[2]*M_TO_AYLIN))

	world.state.player.mvp(door_mat)

	world.state.render_pass_manager.Begin(wgpu.LoadOp_Load, wgpu.LoadOp_Load)
	render_pass = world.state.render_pass_manager.render_pass
	world.door.Draw(render_pass)
	world.state.render_pass_manager.End()
}

func (world *WorldAlexisRoom) Release() {
	world.room.Release()
	world.door.Release()
}
