package main

import (
	"fmt"
	"unsafe"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type IvxHeader struct {
	version_major uint64
	version_minor uint64
	name          [1024]byte

	index_count  uint64
	index_offset uint64

	vertex_count uint64
	components   uint64
	offset       uint64
}

type Vertex struct {
	pos [3]float32
	uv  [2]float32
}

type Model struct {
	state *State

	vbo         *wgpu.Buffer
	ibo         *wgpu.Buffer
	index_count uint32

	texture *Texture
	bind_group *wgpu.BindGroup
}

func NewModel(state *State, label string, vertices []Vertex, indices []uint32, texture []byte) (*Model, error) {
	model := Model{state: state}
	var err error

	// vertex buffer shit

	if model.vbo, err = state.device.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Label:    fmt.Sprintf("VBO (%s)", label),
		Contents: wgpu.ToBytes(vertices[:]),
		Usage:    wgpu.BufferUsage_Vertex,
	}); err != nil {
		return nil, err
	}

	if model.ibo, err = state.device.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Label:    fmt.Sprintf("IBO (%s)", label),
		Contents: wgpu.ToBytes(indices[:]),
		Usage:    wgpu.BufferUsage_Index,
	}); err != nil {
		model.vbo.Release()
		return nil, err
	}

	model.index_count = uint32(len(indices))

	// texture shit

	if model.texture, err = NewTextureFromBytes(state, label, texture); err != nil {
		model.vbo.Release()
		model.ibo.Release()
		return nil, err
	}

	// bind group shit

	if model.bind_group, err = state.device.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Layout: state.regular_pipeline.bind_group_layout,
		Entries: []wgpu.BindGroupEntry{
			{
				Binding: 0,
				TextureView: model.texture.view,
			},
			{
				Binding: 1,
				Sampler: model.texture.sampler,
			},
			{
				Binding: 2,
				Buffer: state.player.mvp_buf,
				Size: wgpu.WholeSize,
			},
		},
	}); err != nil {
		model.vbo.Release()
		model.ibo.Release()
		model.texture.Release()
		return nil, err
	}

	return &model, nil
}

func NewModelFromIvx(state *State, label string, ivx []byte, texture []byte) (*Model, error) {
	header := (*IvxHeader)(unsafe.Pointer(&ivx[0]))

	var indices []uint32

	for i := uint64(0); i < header.index_count; i++ {
		index := (*uint32)(unsafe.Pointer(&ivx[header.index_offset+i*uint64(unsafe.Sizeof(indices[0]))]))
		indices = append(indices, *index)
	}

	var vertices []Vertex

	for i := uint64(0); i < header.vertex_count; i++ {
		vertex := (*Vertex)(unsafe.Pointer(&ivx[header.offset+i*uint64(unsafe.Sizeof(vertices[0]))]))
		vertices = append(vertices, *vertex)
	}

	return NewModel(state, label, vertices, indices, texture)
}

func (model *Model) Draw(render_pass *wgpu.RenderPassEncoder) {
	model.state.regular_pipeline.Set(render_pass, model.bind_group)
	render_pass.SetVertexBuffer(0, model.vbo, 0, wgpu.WholeSize)
	render_pass.SetIndexBuffer(model.ibo, wgpu.IndexFormat_Uint32, 0, wgpu.WholeSize)
	render_pass.DrawIndexed(model.index_count, 1, 0, 0, 0)
}

func (model *Model) Release() {
	model.vbo.Release()
	model.ibo.Release()
}
