package main

import (
	"log"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type RenderPassManager struct {
	state *State
	encoder *wgpu.CommandEncoder
	render_pass *wgpu.RenderPassEncoder
	next_tex *wgpu.TextureView
}

func NewRenderPassManager(state *State) *RenderPassManager {
	return &RenderPassManager{state: state}
}

func (manager *RenderPassManager) Begin(colour_load_op, depth_load_op wgpu.LoadOp) {
	var err error

	if manager.encoder, err = manager.state.device.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{
		Label: "Command encoder",
	}); err != nil {
		panic(err)
	}

	manager.render_pass = manager.encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{
			{
				View:       manager.next_tex,
				LoadOp:     colour_load_op,
				StoreOp:    wgpu.StoreOp_Store,
				ClearValue: wgpu.Color{R: 1, G: 0, B: 0, A: 1},
			},
		},
		DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
			View:              manager.state.depth_texture.view,
			DepthClearValue:   1,
			DepthLoadOp:       depth_load_op,
			DepthStoreOp:      wgpu.StoreOp_Store,
			DepthReadOnly:     false,
			StencilClearValue: 0,
			StencilLoadOp:     wgpu.LoadOp_Load,
			StencilStoreOp:    wgpu.StoreOp_Store,
			StencilReadOnly:   true,
		},
	})
}

func (manager *RenderPassManager) End() {
	manager.render_pass.End()
	manager.render_pass.Release()

	cmd_buf, err := manager.encoder.Finish(nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer cmd_buf.Release()

	manager.state.queue.Submit(cmd_buf)
	manager.encoder.Release()
}
