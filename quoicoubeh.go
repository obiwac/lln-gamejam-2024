package main

import (
	"log"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/rajveermalviya/go-webgpu/wgpu"
	wgpuext_glfw "github.com/rajveermalviya/go-webgpu/wgpuext/glfw"

	_ "embed"
)

func init() {
	runtime.LockOSThread()
}

type State struct {
	win                 *glfw.Window
	instance            *wgpu.Instance
	surface             *wgpu.Surface
	adapter             *wgpu.Adapter
	device              *wgpu.Device
	queue               *wgpu.Queue
	config              *wgpu.SwapChainDescriptor
	swapchain           *wgpu.SwapChain
	depth_texture       *Texture
	render_pass_manager *RenderPassManager
	text                *Text
	player              *Player
	prev_time           float64
	dt                  float32

	// worlds

	alexis_room *WorldAlexisRoom
	apat        *WorldApat

	// pipelines

	regular_pipeline *RegularPipeline
	text_pipeline    *TextPipeline
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

	state.render_pass_manager.next_tex = next_tex
	state.render_pass_manager.encoder = encoder

	state.apat.Render()
	state.alexis_room.Render()

	// draw text

	state.render_pass_manager.Begin(wgpu.LoadOp_Load, wgpu.LoadOp_Clear)
	state.text.Draw(state.render_pass_manager.render_pass)
	state.render_pass_manager.End()

	state.swapchain.Present()
}

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

	backend_type := wgpu.BackendType_Undefined

	if runtime.GOOS == "freebsd" {
		log.Println("FreeBSD detected, there are some driver issues with Vulkan on MESA, using OpenGL backend instead")
		backend_type = wgpu.BackendType_OpenGL
	}

	if state.adapter, err = state.instance.RequestAdapter(&wgpu.RequestAdapterOptions{
		ForceFallbackAdapter: false,
		BackendType:          backend_type,
		CompatibleSurface:    state.surface,
	}); err != nil {
		panic(err)
	}
	defer state.adapter.Release()

	log.Println("Adapter name:\t", state.adapter.GetProperties().Name)
	log.Println("Adapter vendor:\t", state.adapter.GetProperties().VendorName)
	log.Println("Adapter driver:\t", state.adapter.GetProperties().DriverDescription)
	log.Println("Adapter architecture:\t", state.adapter.GetProperties().Architecture)

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

	log.Println("Create WebGPU regular pipeline")

	if state.regular_pipeline, err = NewRegularPipeline(&state); err != nil {
		panic(err)
	}
	defer state.regular_pipeline.Release()

	log.Println("Create WebGPU text pipeline")

	if state.text_pipeline, err = NewTextPipeline(&state); err != nil {
		panic(err)
	}
	defer state.text_pipeline.Release()

	log.Println("Create depth texture")

	if state.depth_texture, err = NewDepthTexture(&state); err != nil {
		panic(err)
	}
	defer state.depth_texture.Release()

	log.Println("Create render pass manager")
	state.render_pass_manager = NewRenderPassManager(&state)

	log.Println("Create player")

	if state.player, err = NewPlayer(&state); err != nil {
		panic(err)
	}
	defer state.player.Release()

	log.Println("Create Alexis' room")

	if state.alexis_room, err = NewWorldAlexisRoom(&state); err != nil {
		panic(err)
	}
	defer state.alexis_room.Release()

	log.Println("Create Apat")

	if state.apat, err = NewWorldApat(&state); err != nil {
		panic(err)
	}
	defer state.apat.Release()

	log.Println("Create text")

	displayDialogue(getDialogues(), "intro1", &state)

	log.Println("Start main loop")

	state.win.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		state.resize(width, height)
	})

	for !state.win.ShouldClose() {
		// Calculate delta time
		current_time := glfw.GetTime()
		state.dt = float32(current_time - state.prev_time)
		state.prev_time = current_time

		glfw.PollEvents()
		state.update()
		state.render()
	}
}
