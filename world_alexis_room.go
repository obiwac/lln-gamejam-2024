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
}

//go:embed res/alexis-room-lightmap.png
var alexis_room_lightmap []byte

//go:embed res/alexis-room.ivx
var alexis_room []byte

func NewWorldAlexisRoom(state *State) (*WorldAlexisRoom, error) {
	room := &WorldAlexisRoom{}
	room.World = World{state: state}

	var err error

	if room.room, err = NewModelFromIvx(state, "Alexis room", alexis_room, alexis_room_lightmap); err != nil {
		return nil, err
	}

	// if room.door, err = NewModelFromIvx(state, "Alexis door", alexis_door, alexis_room_lightmap); err != nil {
	// 	room.room.Release()
	// 	return nil, err
	// }

	return room, nil
}

func (world *WorldAlexisRoom) Render() {
	world.state.player.mvp(NewMat())

	world.state.render_pass_manager.Begin(wgpu.LoadOp_Load, wgpu.LoadOp_Load)
	render_pass := world.state.render_pass_manager.render_pass

	world.room.Draw(render_pass)
	// world.door.Draw(render_pass)

	world.state.render_pass_manager.End()
}

func (world *WorldAlexisRoom) Release() {
	world.room.Release()
	world.door.Release()
}
