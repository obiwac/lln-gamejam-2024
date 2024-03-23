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
	rotation [2]float32

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

		position: [3]float32{0, 1, 0},
		rotation: [2]float32{math.Pi / 2, 0},

		MvpBuf: mvp_buf,
	}, nil
}

func (player *Player) HandleInputs() {
	speed := float32(1.2 / 60)
	multiplier := speed

	// Camera movement

	input := []float32{0, 0}

	if player.state.win.GetKey(glfw.KeyW) == glfw.Press || player.state.win.GetKey(glfw.KeyUp) == glfw.Press {
		input[1] = -1
	}

	if player.state.win.GetKey(glfw.KeyS) == glfw.Press || player.state.win.GetKey(glfw.KeyDown) == glfw.Press {
		input[1] = 1
	}

	if player.state.win.GetKey(glfw.KeyA) == glfw.Press || player.state.win.GetKey(glfw.KeyLeft) == glfw.Press {
		input[0] = -1
	}

	if player.state.win.GetKey(glfw.KeyD) == glfw.Press || player.state.win.GetKey(glfw.KeyRight) == glfw.Press {
		input[0] = 1
	}

	if input[1] != 0 || input[0] != 0 {
		angle := player.rotation[0] - math.Pi/2 + float32(math.Atan2(float64(input[1]), float64(input[0])))
		player.position[0] += float32(math.Cos(float64(angle))) * multiplier
		player.position[2] += float32(math.Sin(float64(angle))) * multiplier
	}
}

func (player *Player) HandleMouse() {
	sensitivity := 6

	// Camera rotation
	x, y := player.state.win.GetCursorPos()
	width, height := player.state.win.GetSize()

	player.rotation[0] += float32((x-float64(width)/2)/float64(width)) * float32(sensitivity)
	player.rotation[1] -= float32((y-float64(height)/2)/float64(height)) * float32(sensitivity)

	player.rotation[1] = float32(math.Max(-math.Pi/2, math.Min(math.Pi/2, float64(player.rotation[1]))))

	// Lock cursor in the center of the window
	player.state.win.SetCursorPos(float64(width)/2, float64(height)/2)

	// Hide cursor
	player.state.win.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
}

func (player *Player) Release() {
	player.MvpBuf.Release()
}

const M_TO_AYLIN = 1 / 1.64

func (player *Player) mvp() *Mat {
	width, height := player.state.win.GetSize()

	player.p.Perspective(math.Pi/2, float32(width)/float32(height), 0.05, 50)
	player.m.Scale(M_TO_AYLIN, M_TO_AYLIN, M_TO_AYLIN) // models are exported in metres, so we must convert to aylins
	player.v.Identity()
	player.v.Multiply(NewMat().Rotate2d((player.rotation[0] - math.Pi/2), player.rotation[1]))
	player.v.Multiply(NewMat().Translation(-player.position[0], -player.position[1], -player.position[2]))

	mvp := NewMat().Multiply(player.p).Multiply(player.v).Multiply(player.m)
	return mvp
}

func (player *Player) Update() {

	mvp := player.mvp()
	player.state.queue.WriteBuffer(player.MvpBuf, 0, wgpu.ToBytes(mvp.Data[:]))

	player.HandleInputs()
	player.HandleMouse()

	if player.state.win.GetKey(glfw.KeyEscape) == glfw.Press {
		println("Escape pressed -> Close window")
		player.state.win.SetShouldClose(true)
	}
}
