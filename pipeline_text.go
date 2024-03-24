package main

import (
	_ "embed"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type TextPipeline struct {
	Pipeline
}

//go:embed shaders/text.wgsl
var text_shader_src string

func NewTextPipeline(state *State) (*TextPipeline, error) {
	pipeline, err := NewPipeline(state, "Text", text_shader_src,
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
			{ // text attributes
				Binding:    2,
				Visibility: wgpu.ShaderStage_Vertex,
				Buffer: wgpu.BufferBindingLayout{
					Type: wgpu.BufferBindingType_Uniform,
				},
			},
		},
		[]wgpu.VertexBufferLayout{},
	)

	if err != nil {
		return nil, err
	}

	return &TextPipeline{
		Pipeline:       *pipeline,
	}, nil
}

func (pipeline *TextPipeline) Release() {
	pipeline.Pipeline.Release()
}
