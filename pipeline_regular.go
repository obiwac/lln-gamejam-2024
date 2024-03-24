package main

import (
	_ "embed"
	"unsafe"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type RegularPipeline struct {
	Pipeline
	vbo_layout    wgpu.VertexBufferLayout
}

//go:embed shaders/regular.wgsl
var regular_shader_src string

func NewRegularPipeline(state *State) (*RegularPipeline, error) {
	vbo_layout := wgpu.VertexBufferLayout{
		ArrayStride: uint64(unsafe.Sizeof(Vertex{})),
		StepMode:    wgpu.VertexStepMode_Vertex,
		Attributes: []wgpu.VertexAttribute{
			{
				Format:         wgpu.VertexFormat_Float32x3,
				Offset:         0,
				ShaderLocation: 0,
			},
			{
				Format:         wgpu.VertexFormat_Float32x2,
				Offset:         4 * 3,
				ShaderLocation: 1,
			},
		},
	}

	pipeline, err := NewPipeline(state, "Regular", regular_shader_src,
		[]wgpu.BindGroupLayoutEntry{
			{ // texture
				Binding:    0,
				Visibility: wgpu.ShaderStage_Fragment,
				Texture: wgpu.TextureBindingLayout{
					Multisampled:  false,
					ViewDimension: wgpu.TextureViewDimension_2D,
					SampleType:    wgpu.TextureSampleType_Float,
				},
			},
			{ // sampler
				Binding:    1,
				Visibility: wgpu.ShaderStage_Fragment,
				Sampler: wgpu.SamplerBindingLayout{
					Type: wgpu.SamplerBindingType_Filtering,
				},
			},
			{ // MVP matrix
				Binding:    2,
				Visibility: wgpu.ShaderStage_Vertex,
				Buffer: wgpu.BufferBindingLayout{
					Type: wgpu.BufferBindingType_Uniform,
				},
			},
		},
		[]wgpu.VertexBufferLayout{
			vbo_layout,
		},
	)

	if err != nil {
		return nil, err
	}

	return &RegularPipeline{
		Pipeline: *pipeline,
		vbo_layout: vbo_layout,
	}, nil
}

func (pipeline *RegularPipeline) Release() {
	pipeline.Pipeline.Release()
}
