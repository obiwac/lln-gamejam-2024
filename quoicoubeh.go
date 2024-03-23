package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/rajveermalviya/go-webgpu/wgpu"
	wgpuext_glfw "github.com/rajveermalviya/go-webgpu/wgpuext/glfw"

	_ "embed"
)

func init() {
	runtime.LockOSThread()
}

type State struct {
	win               *glfw.Window
	instance          *wgpu.Instance
	surface           *wgpu.Surface
	adapter           *wgpu.Adapter
	device            *wgpu.Device
	queue             *wgpu.Queue
	config            *wgpu.SwapChainDescriptor
	swapchain         *wgpu.SwapChain
	shader            *wgpu.ShaderModule
	bind_group_layout *wgpu.BindGroupLayout
	bind_group        *wgpu.BindGroup
	vbo_layout        wgpu.VertexBufferLayout
	depth_texture     *Texture
	pipeline_layout   *wgpu.PipelineLayout
	pipeline          *wgpu.RenderPipeline
	texture           *Texture
	model             *Model
	player            *Player
	delta_time        float64
}

func (state *State) resize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	state.config.Width = uint32(width)
	state.config.Height = uint32(height)

	log.Printf("Window resized to %dx%d, recreate swapchain and depth texture\n", width, height)

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

	state.depth_texture.Release()
	if state.depth_texture, err = NewDepthTexture(state); err != nil {
		log.Fatal(err)
		return
	}
}

func (state *State) update() {
	state.player.Update()
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
				ClearValue: wgpu.Color{R: 1, G: 0, B: 0, A: 1},
			},
		},
		DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
			View:              state.depth_texture.View,
			DepthClearValue:   1,
			DepthLoadOp:       wgpu.LoadOp_Clear,
			DepthStoreOp:      wgpu.StoreOp_Store,
			DepthReadOnly:     false,
			StencilClearValue: 0,
			StencilLoadOp:     wgpu.LoadOp_Load,
			StencilStoreOp:    wgpu.StoreOp_Store,
			StencilReadOnly:   true,
		},
	})
	defer render_pass.Release()

	render_pass.SetPipeline(state.pipeline)
	render_pass.SetBindGroup(0, state.bind_group, nil)
	state.model.Draw(render_pass)
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

//go:embed res/alexis-room-lightmap.png
var alexis_room_lightmap []byte

//go:embed res/alexis-room.ivx
var alexis_room []byte

func main() {
	state := State{}

	log.Println("Create GLFW window")

	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	mon_width, mon_height := glfw.GetPrimaryMonitor().GetContentScale()

	if runtime.GOOS != "darwin" && (mon_width != 1 || mon_height != 1) {
		panic("Monitor scaling is not 1:1 and not on macOS, things might explode, aborting now")
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

	log.Println("Create WebGPU bind group layout")

	if state.bind_group_layout, err = state.device.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "Bind group layout",
		Entries: []wgpu.BindGroupLayoutEntry{
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
	}); err != nil {
		panic(err)
	}
	defer state.bind_group_layout.Release()

	log.Println("Create WebGPU VBO layout")

	state.vbo_layout = wgpu.VertexBufferLayout{
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

	log.Println("Create depth texture")

	if state.depth_texture, err = NewDepthTexture(&state); err != nil {
		panic(err)
	}
	defer state.depth_texture.Release()

	log.Println("Create WebGPU pipeline layout")

	if state.pipeline_layout, err = state.device.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label: "Pipeline layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{
			state.bind_group_layout,
		},
	}); err != nil {
		panic(err)
	}
	defer state.pipeline_layout.Release()

	log.Println("Create WebGPU render pipeline")

	if state.pipeline, err = state.device.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  "Render pipeline",
		Layout: state.pipeline_layout,
		Primitive: wgpu.PrimitiveState{
			Topology:         wgpu.PrimitiveTopology_TriangleList,
			StripIndexFormat: wgpu.IndexFormat_Undefined,
			FrontFace:        wgpu.FrontFace_CCW,
			CullMode:         wgpu.CullMode_None,
		},
		Vertex: wgpu.VertexState{
			Module:     state.shader,
			EntryPoint: "vert_main",
			Buffers: []wgpu.VertexBufferLayout{
				state.vbo_layout,
			},
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
		panic(err)
	}
	defer state.pipeline.Release()

	log.Println("Load texture")

	if state.texture, err = NewTextureFromBytes(&state, "Alexis room lightmap", alexis_room_lightmap); err != nil {
		panic(err)
	}
	defer state.texture.Release()

	log.Println("Load model")

	if state.model, err = NewModelFromIvx(&state, "Alexis room", alexis_room); err != nil {
		panic(err)
	}
	defer state.model.Release()

	log.Println("Create player")

	if state.player, err = NewPlayer(&state); err != nil {
		panic(err)
	}
	defer state.player.Release()

	log.Println("Create WebGPU bind group")

	if state.bind_group, err = state.device.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "Bind group",
		Layout: state.bind_group_layout,
		Entries: []wgpu.BindGroupEntry{
			{ // texture
				Binding:     0,
				TextureView: state.texture.View,
			},
			{ // sampler
				Binding: 1,
				Sampler: state.texture.sampler,
			},
			{ // MVP matrix
				Binding: 2,
				Buffer:  state.player.MvpBuf,
				Size:    wgpu.WholeSize,
			},
		},
	}); err != nil {
		panic(err)
	}
	defer state.bind_group.Release()

	/*log.Println("Create sound system")
	SoundSystem := NewSoundSystem()

	log.Println("Play music")
	if err := SoundSystem.PlaySound("res/sound/sos.mp3"); err != nil {
		panic(err)
	}*/

	log.Println("Start main loop")

	state.win.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		state.resize(width, height)
	})

	for !state.win.ShouldClose() {
		// Calculate delta time
		current_time := glfw.GetTime()
		state.delta_time = current_time - state.delta_time
		state.delta_time = current_time

		glfw.PollEvents()
		state.update()
		state.render()
	}
}
