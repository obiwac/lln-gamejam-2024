package main

import (
	"fmt"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type Vertex struct {
	pos [3]float32
	uv [2]float32
}

type Model struct {
	vbo *wgpu.Buffer
	ibo *wgpu.Buffer
	index_count uint32
}

func NewModel(state *State, label string, vertices []float32, indices []uint32) (*Model, error) {
	model := Model{}

	var err error

	if model.vbo, err = state.device.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Label: fmt.Sprintf("VBO (%s)", label),
		Contents: wgpu.ToBytes(vertices[:]),
		Usage: wgpu.BufferUsage_Vertex,
	}); err != nil {
		return nil, err
	}

	if model.ibo, err = state.device.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Label: fmt.Sprintf("IBO (%s)", label),
		Contents: wgpu.ToBytes(indices[:]),
		Usage: wgpu.BufferUsage_Index,
	}); err != nil {
		model.vbo.Release()
		return nil, err
	}

	model.index_count = uint32(len(indices))

	return &model, nil
}

func NewModelFromObj(state *State, label, obj string) (*Model, error) {
	// TODO parse obj here

	vertices := []float32{}
	indices := []uint32{}

	return NewModel(state, label, vertices, indices)
}

func (model *Model) Draw(render_pass *wgpu.RenderPassEncoder) {
	render_pass.SetVertexBuffer(0, model.vbo, 0, wgpu.WholeSize)
	render_pass.SetIndexBuffer(model.ibo, wgpu.IndexFormat_Uint32, 0, wgpu.WholeSize)
	render_pass.DrawIndexed(model.index_count, 1, 0, 0, 0)
}

func (model *Model) Release() {
	model.vbo.Release()
	model.ibo.Release()
}
