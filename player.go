package main

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type Player struct {
	state *State

	p *Mat
	m *Mat
	v *Mat

	position [3]float32

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

		position: [3]float32{0, 0, 0},

		MvpBuf: mvp_buf,
	}, nil
}

func (player *Player) handleInputs() {
	if player.state.win.GetKey(glfw.KeyW) == glfw.Press || player.state.win.GetKey(glfw.KeyUp) == glfw.Press {
		player.position[2] += 0.01
	}

	if player.state.win.GetKey(glfw.KeyS) == glfw.Press || player.state.win.GetKey(glfw.KeyDown) == glfw.Press {
		player.position[2] -= 0.01
	}

	if player.state.win.GetKey(glfw.KeyA) == glfw.Press || player.state.win.GetKey(glfw.KeyLeft) == glfw.Press {
		player.position[0] -= 0.01
	}

	if player.state.win.GetKey(glfw.KeyD) == glfw.Press || player.state.win.GetKey(glfw.KeyRight) == glfw.Press {
		player.position[0] += 0.01
	}
}

func (player *Player) Release() {
	player.MvpBuf.Release()
}

func (player *Player) mvp() *Mat {
	width, height := player.state.win.GetSize()

	player.p.Perspective(math.Pi / 6, float32(width) / float32(height), 0.1, 500)
	player.m.Identity().Translation(0, 0, 0)
	player.v.Identity().Translation(-player.position[0], -player.position[1], -player.position[2])

	player.p.Perspective(math.Pi / 6, float32(width) / float32(height), 0.1, 500)
	player.m.Translation(0, 0, 0)
	player.v.Identity().Translation(-player.position[0], -player.position[1], -player.position[2])

	mvp := NewMat().Multiply(player.m).Multiply(player.v).Multiply(player.p)
	return mvp
}

func (player *Player) Update() {
	mvp := player.mvp()
	player.state.queue.WriteBuffer(player.MvpBuf, 0, wgpu.ToBytes(mvp.Data[:]))

	player.handleInputs()

	if player.state.win.GetKey(glfw.KeyEscape) == glfw.Press {
		println("Escape pressed -> Close window")
		player.state.win.SetShouldClose(true)
	}
}
