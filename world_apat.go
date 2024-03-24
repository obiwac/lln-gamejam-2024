package main

import (
	_ "embed"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type WorldApat struct {
	World

	landscape *Model
	portal *Model
	ukulele *Model

	ukulele_picked_up bool
	portal_lit bool
}

//go:embed res/apat-lightmap.png
var apat_lightmap []byte

//go:embed res/apat-landscape.ivx
var apat_landscape []byte

//go:embed res/apat-portal.ivx
var apat_portal []byte

//go:embed res/apat-ukulele.ivx
var apat_ukulele []byte

func NewWorldApat(state *State) (*WorldApat, error) {
	apat := &WorldApat{}
	apat.World = World{state: state}

	var err error

	if apat.landscape, err = NewModelFromIvx(state, "Apat landscape", apat_landscape, apat_lightmap, true); err != nil {
		return nil, err
	}

	apat.landscape.collider_off_y = -10

	if apat.portal, err = NewModelFromIvx(state, "Apat portal", apat_portal, apat_lightmap, false); err != nil {
		apat.landscape.Release()
		return nil, err
	}

	if apat.ukulele, err = NewModelFromIvx(state, "Apat ukulele", apat_ukulele, apat_lightmap, false); err != nil {
		apat.landscape.Release()
		apat.portal.Release()
		return nil, err
	}

	apat.ukulele_picked_up = false
	apat.portal_lit = false

	return apat, nil
}

func (world *WorldApat) Render() {
	world.state.player.mvp(NewMat().Translation(0, world.landscape.collider_off_y, 0))

	world.state.render_pass_manager.Begin(wgpu.LoadOp_Clear, wgpu.LoadOp_Clear)
	render_pass := world.state.render_pass_manager.render_pass

	world.landscape.Draw(render_pass)

	if (world.portal_lit) {
		world.portal.Draw(render_pass)
	}

	if (!world.ukulele_picked_up) {
		world.ukulele.Draw(render_pass)
	}

	world.state.render_pass_manager.End()
}

func (world *WorldApat) Release() {
	world.landscape.Release()
	world.portal.Release()
	world.ukulele.Release()
}
