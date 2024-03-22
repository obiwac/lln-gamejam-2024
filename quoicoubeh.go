package main

import (
	"log"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/rajveermalviya/go-webgpu/wgpu"
	wgpuext_glfw "github.com/rajveermalviya/go-webgpu/wgpuext/glfw"

	_ "embed"
)

type State struct {
	win       *glfw.Window
	instance  *wgpu.Instance
	surface   *wgpu.Surface
	adapter   *wgpu.Adapter
	device    *wgpu.Device
	queue     *wgpu.Queue
	config    *wgpu.SwapChainDescriptor
	swapchain *wgpu.SwapChain
	shader    *wgpu.ShaderModule
	pipeline  *wgpu.RenderPipeline
}

func (state *State) resize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	state.config.Width = uint32(width)
	state.config.Height = uint32(height)

	log.Printf("Window resized to %dx%d, recreate swapchain\n", width, height)

	swapchain, err := state.device.CreateSwapChain(state.surface, state.config)
	if err != nil {
		log.Fatal(err)
		return
	}

	if swapchain != nil {
		if state.swapchain != nil {
			state.swapchain.Release()
		}
		state.swapchain = swapchain
	} else {
		log.Fatal("Failed to recreate swapchain")
	}
}

func (state *State) render() {
	next_tex, err := state.swapchain.GetCurrentTextureView()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer next_tex.Release()

	encoder, err := state.device.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{
		Label: "Command encoder",
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer encoder.Release()

	render_pass := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{
			{
				View:       next_tex,
				LoadOp:     wgpu.LoadOp_Clear,
				StoreOp:    wgpu.StoreOp_Store,
				ClearValue: wgpu.Color{R: 0, G: 0, B: 0, A: 1},
			},
		},
	})
	defer render_pass.Release()

	render_pass.SetPipeline(state.pipeline)
	render_pass.Draw(3, 1, 0, 0)
	render_pass.End()

	cmd_buf, err := encoder.Finish(nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer cmd_buf.Release()

	state.queue.Submit(cmd_buf)
	state.swapchain.Present()
}

//go:embed shader.wgsl
var shader_src string

func main() {
	state := State{}

	log.Println("Create GLFW window")

	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	mon_width, mon_height := glfw.GetPrimaryMonitor().GetContentScale()

	if mon_width != 1 || mon_height != 1 {
		panic("Monitor scaling is not 1:1, things might explode, aborting now")
	}

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI) // tell GLFW not to create an OpenGL context automatically
	// TODO once this is added to go-gl/glfw, unset GLFW_SCALE_FRAMEBUFFER

	var err error

	if state.win, err = glfw.CreateWindow(800, 600, "Quoicoubeh", nil, nil); err != nil {
		panic(err)
	}
	defer state.win.Destroy()

	log.Println("Create WebGPU instance")

	state.instance = wgpu.CreateInstance(nil)
	defer state.instance.Release()

	log.Println("Create WebGPU surface")

	surface_descr := wgpuext_glfw.GetSurfaceDescriptor(state.win)
	state.surface = state.instance.CreateSurface(surface_descr)
	defer state.surface.Release()

	log.Println("Request WebGPU adapter")

	if state.adapter, err = state.instance.RequestAdapter(&wgpu.RequestAdapterOptions{
		CompatibleSurface: state.surface,
	}); err != nil {
		panic(err)
	}
	defer state.adapter.Release()

	log.Println("Request WebGPU device")

	if state.device, err = state.adapter.RequestDevice(nil); err != nil {
		panic(err)
	}
	defer state.device.Release()

	log.Println("Get WebGPU queue")

	state.queue = state.device.GetQueue()
	defer state.queue.Release()

	log.Println("Create WebGPU swapchain")

	caps := state.surface.GetCapabilities(state.adapter)
	width, height := state.win.GetSize()

	state.config = &wgpu.SwapChainDescriptor{
		Usage:       wgpu.TextureUsage_RenderAttachment,
		Format:      caps.Formats[0],
		Width:       uint32(width),
		Height:      uint32(height),
		PresentMode: wgpu.PresentMode_Fifo,
		AlphaMode:   caps.AlphaModes[0],
	}

	if state.swapchain, err = state.device.CreateSwapChain(state.surface, state.config); err != nil {
		panic(err)
	}
	defer state.swapchain.Release()

	log.Println("Create WebGPU shader module")

	if state.shader, err = state.device.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label: "shader.wgsl",
		WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{
			Code: shader_src,
		},
	}); err != nil {
		panic(err)
	}
	defer state.shader.Release()

	log.Println("Create WebGPU render pipeline")

	if state.pipeline, err = state.device.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label: "Render pipeline",
		Primitive: wgpu.PrimitiveState{
			Topology:         wgpu.PrimitiveTopology_TriangleList,
			StripIndexFormat: wgpu.IndexFormat_Undefined,
			FrontFace:        wgpu.FrontFace_CCW,
			CullMode:         wgpu.CullMode_None,
		},
		Vertex: wgpu.VertexState{
			Module:     state.shader,
			EntryPoint: "vert_main",
		},
		Fragment: &wgpu.FragmentState{
			Module:     state.shader,
			EntryPoint: "frag_main",
			Targets: []wgpu.ColorTargetState{
				{
					Format:    state.config.Format,
					Blend:     &wgpu.BlendState_Replace,
					WriteMask: wgpu.ColorWriteMask_All,
				},
			},
		},
		Multisample: wgpu.MultisampleState{
			Count:                  1,
			Mask:                   0xFFFFFFFF,
			AlphaToCoverageEnabled: false,
		},
	}); err != nil {
		panic(err)
	}
	defer state.pipeline.Release()

	state.win.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		state.resize(width, height)
	})

	log.Println("Start main loop")

	for !state.win.ShouldClose() {
		glfw.PollEvents()
		state.render()
	}
}
