package main

import (
	"math"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type Player struct {
	state *State

	p *Mat
	m *Mat
	v *Mat

	MvpBuf *wgpu.Buffer
}

func NewPlayer(state *State) (*Player, error) {
	mvp_buf, err := state.device.CreateBuffer(&wgpu.BufferDescriptor{
		Size:  64,
		Usage: wgpu.BufferUsage_Uniform | wgpu.BufferUsage_CopyDst,
	})

	if err != nil {
		return nil, err
	}

	return &Player{
		state: state,

		p: NewMat().Identity(),
		m: NewMat().Identity(),
		v: NewMat().Identity(),

		MvpBuf: mvp_buf,
	}, nil
}

func (player *Player) Release() {
	player.MvpBuf.Release()
}

func (player *Player) mvp() *Mat {
	width, height := player.state.win.GetSize()

	player.p.Perspective(math.Pi / 6, float32(width) / float32(height), 0.1, 500)
	player.m.Translation(0, 0, 0)
	player.v.Translation(0, 0, -5)

	mvp := NewMat().Multiply(player.m).Multiply(player.v).Multiply(player.p)
	return mvp
}

func (player *Player) Update() {
	mvp := player.mvp()
	player.state.queue.WriteBuffer(player.MvpBuf, 0, wgpu.ToBytes(mvp.Data[:]))
}
