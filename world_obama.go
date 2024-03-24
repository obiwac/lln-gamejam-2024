package main

import (
	_ "embed"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type WorldObama struct {
	World
	room *Model
	should_draw bool
}

//go:embed res/obama-lightmap.png
var obama_room_lightmap []byte

//go:embed res/obama-room.ivx
var obama_room []byte

func NewWorldObama(state *State) (*WorldObama, error) {
	room := &WorldObama{}
	room.World = World{state: state}

	var err error

	if room.room, err = NewModelFromIvx(state, "Obama room", obama_room, obama_room_lightmap, false); err != nil {
		return nil, err
	}

	return room, nil
}

func (world *WorldObama) Render() {
	if !world.should_draw {
		return
	}

	world.state.player.mvp(NewMat())

	world.state.render_pass_manager.Begin(wgpu.LoadOp_Clear, wgpu.LoadOp_Clear)
	render_pass := world.state.render_pass_manager.render_pass
	world.room.Draw(render_pass)
	world.state.render_pass_manager.End()
}

func (world *WorldObama) Release() {
	world.room.Release()
}
