package main

import (
	"fmt"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type Pipeline struct {
	shader            *wgpu.ShaderModule
	bind_group_layout *wgpu.BindGroupLayout
	pipeline_layout   *wgpu.PipelineLayout
	pipeline          *wgpu.RenderPipeline
}

func NewPipeline(state *State, label, src string, bind_group_layout_entries []wgpu.BindGroupLayoutEntry, vbo_layouts []wgpu.VertexBufferLayout) (*Pipeline, error) {
	pipeline := &Pipeline{}
	var err error

	if pipeline.shader, err = state.device.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label: fmt.Sprintf("Shader module (%s)", label),
		WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{
			Code: src,
		},
	}); err != nil {
		return nil, err
	}

	if pipeline.bind_group_layout, err = state.device.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label:   fmt.Sprintf("Bind group layout (%s)", label),
		Entries: bind_group_layout_entries,
	}); err != nil {
		pipeline.shader.Release()
		return nil, err
	}

	if pipeline.pipeline_layout, err = state.device.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label: fmt.Sprintf("Pipeline layout (%s)", label),
		BindGroupLayouts: []*wgpu.BindGroupLayout{
			pipeline.bind_group_layout,
		},
	}); err != nil {
		pipeline.shader.Release()
		pipeline.bind_group_layout.Release()
		return nil, err
	}

	if pipeline.pipeline, err = state.device.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  fmt.Sprintf("Render pipeline (%s)", label),
		Layout: pipeline.pipeline_layout,
		Primitive: wgpu.PrimitiveState{
			Topology:         wgpu.PrimitiveTopology_TriangleList,
			StripIndexFormat: wgpu.IndexFormat_Undefined,
			FrontFace:        wgpu.FrontFace_CCW,
			CullMode:         wgpu.CullMode_None,
		},
		Vertex: wgpu.VertexState{
			Module:     pipeline.shader,
			EntryPoint: "vert_main",
			Buffers: vbo_layouts,
		},
		Fragment: &wgpu.FragmentState{
			Module:     pipeline.shader,
			EntryPoint: "frag_main",
			Targets: []wgpu.ColorTargetState{
				{
					Format:    state.config.Format,
					Blend:     &wgpu.BlendState_Replace,
					WriteMask: wgpu.ColorWriteMask_All,
				},
			},
		},
		DepthStencil: &wgpu.DepthStencilState{
			Format:            DEPTH_FORMAT,
			DepthWriteEnabled: true,
			DepthCompare:      wgpu.CompareFunction_Less,
			StencilFront: wgpu.StencilFaceState{
				Compare: wgpu.CompareFunction_Always,
			},
			StencilBack: wgpu.StencilFaceState{
				Compare: wgpu.CompareFunction_Always,
			},
		},
		Multisample: wgpu.MultisampleState{
			Count:                  1,
			Mask:                   0xFFFFFFFF,
			AlphaToCoverageEnabled: false,
		},
	}); err != nil {
		pipeline.shader.Release()
		pipeline.bind_group_layout.Release()
		pipeline.pipeline_layout.Release()
		return nil, err
	}

	return pipeline, nil
}

func (pipeline *Pipeline) Set(render_pass *wgpu.RenderPassEncoder, bind_group *wgpu.BindGroup) {
	render_pass.SetPipeline(pipeline.pipeline)
	render_pass.SetBindGroup(0, bind_group, nil)
}

func (pipeline *Pipeline) Release() {
	pipeline.shader.Release()
	pipeline.bind_group_layout.Release()
	pipeline.pipeline_layout.Release()
	pipeline.pipeline.Release()
}
