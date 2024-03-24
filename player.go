package main

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type Player struct {
	Entity
	state *State

	p *Mat
	v *Mat

	mvp_buf *wgpu.Buffer
}

func NewPlayer(state *State) (*Player, error) {
	mvp_buf, err := state.device.CreateBuffer(&wgpu.BufferDescriptor{
		Size:  64,
		Usage: wgpu.BufferUsage_Uniform | wgpu.BufferUsage_CopyDst,
	})

	if err != nil {
		return nil, err
	}

	position := [3]float32{0, 0, 0}
	rotation := [2]float32{math.Pi / 2, 0}

	return &Player{
		Entity: *NewEntity(state, position, rotation, 0.2, 1.72*M_TO_AYLIN),
		state:  state,

		p: NewMat().Identity(),
		v: NewMat().Identity(),

		mvp_buf: mvp_buf,
	}, nil
}

func (player *Player) HandleInputs() {
	speed := float32(1.5)

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

	if player.state.win.GetKey(glfw.KeySpace) == glfw.Press {
		player.Jump()
	}

	if input[1] != 0 || input[0] != 0 {
		angle := player.rot[0] - math.Pi/2 + float32(math.Atan2(float64(input[1]), float64(input[0])))
		player.acc[0] = float32(math.Cos(float64(angle))) * speed
		player.acc[2] = float32(math.Sin(float64(angle))) * speed
	}
}

func (player *Player) HandleMouse() {
	sensitivity := 6

	// Camera rotation
	x, y := player.state.win.GetCursorPos()
	width, height := player.state.win.GetSize()

	player.rot[0] += float32((x-float64(width)/2)/float64(width)) * float32(sensitivity)
	player.rot[1] -= float32((y-float64(height)/2)/float64(height)) * float32(sensitivity)

	player.rot[1] = float32(math.Max(-math.Pi/2, math.Min(math.Pi/2, float64(player.rot[1]))))

	// Lock cursor in the center of the window
	player.state.win.SetCursorPos(float64(width)/2, float64(height)/2)

	// Hide cursor
	player.state.win.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
}

func (player *Player) Release() {
	player.mvp_buf.Release()
}

const M_TO_AYLIN = 1 / 1.64

func (player *Player) mvp(m *Mat) *Mat {
	width, height := player.state.win.GetSize()
	eyelevel := float32(1) // exactly one Aylin

	player.p.Perspective(math.Pi/2, float32(width)/float32(height), 0.01, 50)

	player.v.Identity()
	player.v.Multiply(NewMat().Rotate2d((player.rot[0] - math.Pi/2), player.rot[1]))
	player.v.Multiply(NewMat().Translation(-player.pos[0], -player.pos[1]-eyelevel, -player.pos[2]))

	aylin_conversion_mat := NewMat().Scale(M_TO_AYLIN, M_TO_AYLIN, M_TO_AYLIN)

	mvp := NewMat().Multiply(player.p).Multiply(player.v).Multiply(m).Multiply(aylin_conversion_mat)
	player.state.queue.WriteBuffer(player.mvp_buf, 0, wgpu.ToBytes(mvp.Data[:]))

	return mvp
}

func (player *Player) Update() {
	player.HandleInputs()
	player.HandleMouse()

	player.Entity.Update([]*Model{
		player.state.alexis_room.room,
		player.state.apat.landscape,
	})

	if player.state.win.GetKey(glfw.KeyEscape) == glfw.Press {
		println("Escape pressed -> Close window")
		player.state.win.SetShouldClose(true)
	}
}
