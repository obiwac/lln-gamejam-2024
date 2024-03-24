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
	win           *glfw.Window
	instance      *wgpu.Instance
	surface       *wgpu.Surface
	adapter       *wgpu.Adapter
	device        *wgpu.Device
	queue         *wgpu.Queue
	config        *wgpu.SwapChainDescriptor
	swapchain     *wgpu.SwapChain
	depth_texture *Texture
	model         *Model
	text          *Text
	player        *Player
	delta_time    float64

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
			View:              state.depth_texture.view,
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

	state.model.Draw(render_pass)
	state.text.Draw(render_pass)
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

//go:embed res/alexis-room-lightmap.png
var alexis_room_lightmap []byte

//go:embed res/alexis-room.ivx
var alexis_room []byte

//go:embed tools/coordinates.csv
var coordinates_csv []byte

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

	log.Println("Create player")

	if state.player, err = NewPlayer(&state); err != nil {
		panic(err)
	}
	defer state.player.Release()

	log.Println("Get colliders coordinates")

	coordinates := GetCoordinatesFromCsv(coordinates_csv)

	for _, coordinate := range coordinates {
		log.Println(coordinate.MeshName, coordinate.MostPositive, coordinate.MostNegative)
	}

	log.Println("Load model")

	if state.model, err = NewModelFromIvx(&state, "Alexis room", alexis_room, alexis_room_lightmap); err != nil {
		panic(err)
	}
	defer state.model.Release()

	log.Println("Create text")

	if state.text, err = NewText(&state, "Quoi ? Coubeh.", 0, 0, 1, 1); err != nil {
		panic(err)
	}
	defer state.text.Release()

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
