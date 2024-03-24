package main

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/rajveermalviya/go-webgpu/wgpu"

	_ "embed"
)

//go:embed tools/coordinates.csv
var coordinates_csv []byte

//go:embed tools/apat.csv
var apat_csv []byte

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

type Heightmap struct {
	neg_x, neg_z  float32
	pos_x, pos_z  float32
	res  int
	heightmap   [][]float32
}

type Model struct {
	state *State

	vbo         *wgpu.Buffer
	ibo         *wgpu.Buffer
	index_count uint32

	texture    *Texture
	bind_group *wgpu.BindGroup

	collider_off_x, collider_off_y, collider_off_z float32

	heightmap *Heightmap
	colliders []Collider
}

func NewModel(state *State, label string, vertices []Vertex, indices []uint32, texture []byte, heightmap bool) (*Model, error) {
	model := Model{state: state}
	var err error

	// heightmap shit

	if heightmap {
		vertex_count := len(vertices)
		model.heightmap = &Heightmap{}

		// get resolution & bounds

		model.heightmap.res = int(math.Sqrt(float64(vertex_count))) // XXX insh'allah

		model.heightmap.neg_x = 9999
		model.heightmap.neg_z = 9999

		model.heightmap.pos_x = -9999
		model.heightmap.pos_z = -9999

		for i := 0; i < vertex_count; i++ {
			vertex := &vertices[i]

			if vertex.pos[0] < model.heightmap.neg_x { model.heightmap.neg_x = vertex.pos[0] }
			if vertex.pos[2] < model.heightmap.neg_z { model.heightmap.neg_z = vertex.pos[2] }
			if vertex.pos[0] > model.heightmap.pos_x { model.heightmap.pos_x = vertex.pos[0] }
			if vertex.pos[2] > model.heightmap.pos_z { model.heightmap.pos_z = vertex.pos[2] }
		}

		model.heightmap.heightmap = make([][]float32, model.heightmap.res)

		for i := range(model.heightmap.heightmap) {
			model.heightmap.heightmap[i] = make([]float32, model.heightmap.res)
		}

		// fill in values

		for i := 0; i < vertex_count; i++ {
			vertex := &vertices[i]

			x := int(float32(model.heightmap.res) * (vertex.pos[0] - model.heightmap.neg_x) / (model.heightmap.pos_x - model.heightmap.neg_x))
			z := int(float32(model.heightmap.res) * (vertex.pos[2] - model.heightmap.neg_z) / (model.heightmap.pos_z - model.heightmap.neg_z))

			if x < 0 || z < 0 || x >= int(model.heightmap.res) || z >= int(model.heightmap.res) {
				continue
			}

			model.heightmap.heightmap[x][z] = vertex.pos[1]
		}

		for i := 0; i < model.heightmap.res; i++ {
			for j := 0; j < model.heightmap.res; j++ {
				if model.heightmap.heightmap[i][j] == 0 && i > 0 {
					model.heightmap.heightmap[i][j] = model.heightmap.heightmap[i-1][j]
				}
			}
		}
	}

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
				Binding:     0,
				TextureView: model.texture.view,
			},
			{
				Binding: 1,
				Sampler: model.texture.sampler,
			},
			{
				Binding: 2,
				Buffer:  state.player.mvp_buf,
				Size:    wgpu.WholeSize,
			},
		},
	}); err != nil {
		model.vbo.Release()
		model.ibo.Release()
		model.texture.Release()
		return nil, err
	}

	colliders_coords_alexis_room := GetCoordinatesFromCsv(coordinates_csv)
	for _, coords := range colliders_coords_alexis_room {
		collider := NewCollider(coords.MeshName, coords.MostNegative, coords.MostPositive)
		model.colliders = append(model.colliders, *collider)
	}

	colliders_coords_ukulele := GetCoordinatesFromCsv(apat_csv)
	for _, coords := range colliders_coords_ukulele {
		collider := NewCollider(coords.MeshName, coords.MostNegative, coords.MostPositive)
		model.colliders = append(model.colliders, *collider)
	}

	return &model, nil
}

func (model *Model) ColliderOffset(x, y, z float32) {
	model.collider_off_x += x
	model.collider_off_y += y
	model.collider_off_z += z

	for i := 0; i < len(model.colliders); i++ {
		model.colliders[i].position1[0] += x
		model.colliders[i].position1[1] += y
		model.colliders[i].position1[2] += z

		model.colliders[i].position2[0] += x
		model.colliders[i].position2[1] += y
		model.colliders[i].position2[2] += z
	}
}

func NewModelFromIvx(state *State, label string, ivx []byte, texture []byte, heightmap bool) (*Model, error) {
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

	return NewModel(state, label, vertices, indices, texture, heightmap)
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
