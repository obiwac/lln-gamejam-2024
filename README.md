# lln-gamejam-2024

Entry for the 2024 edition of the [Louvain-li-Nux gamejam](https://louvainlinux.org/activites/game-jam).
Written in Go with custom WebGPU engine using the [`go-webgpu` bindings](https://github.com/rajveermalviya/go-webgpu), and (unfortunately) GLFW for windowing.

## Building

On Linux/FreeBSD with X11 or macOS:

```console
go build
```

On Linux/FreeBSD with Wayland:

```console
go build -tags wayland
```

These will create a `quoicoubeh` executable which you can run directly:

```console
./quoicoubeh
```

### Extra notes for FreeBSD

The way `go-webgpu` works is by distributing pre-compiled static libraries for WebGPU (`libwgpu_native.a`) for Linux and macOS.
These don't exist on FreeBSD, so you must build it yourself:

```console
git clone https://github.com/gfx-rs/wgpu-native
cd wgpu-native
git checkout 2773864
git submodule update --init --recursive .
cargo build --release
cp target/release/libwgpu_native.* /usr/local/lib
```

We must checkout to `2773864` because subsequent commits remove the `wgpuSwapChain*` functions which are currently required for `go-webgpu`.
This isn't an issue for other platforms as they are distributed a version which still has these functions.

It also needs to be modified to support FreeBSD in `wgpuext/glfw`, by creating a `~/go/pkg/mod/github.com/rajveermalviya/go-webgpu/wgpuext/glfw@v0.1.1/surface_wayland_freebsd.go` file:

```go
// go:build freebsd && wayland

package wgpuext_glfw // import "github.com/rajveermalviya/go-webgpu/wgpuext/glfw"

import "C"

import (
        "unsafe"

        "github.com/go-gl/glfw/v3.3/glfw"
        "github.com/rajveermalviya/go-webgpu/wgpu"
)

func GetSurfaceDescriptor(w *glfw.Window) *wgpu.SurfaceDescriptor {
        return &wgpu.SurfaceDescriptor{
                WaylandSurface: &wgpu.SurfaceDescriptorFromWaylandSurface{
                        Display: unsafe.Pointer(glfw.GetWaylandDisplay()),
                        Surface: unsafe.Pointer(w.GetWaylandWindow()),
                },
        }
}
```

(X11 should be self-explanatory, you just need to change the `wayland` build constraint to `!wayland` and copy the rest of the contents from `surface_x11_linux.go`.)
(If you have any issues, first try to clean the build cache with `go clean -cache`.)

Finally, you need to add the following linker flags in `~/go/pkg/mod/github.com/rajveermalviya/go-webgpu/wgpu@v0.17.1/wgpu.go`:

```go
#cgo freebsd LDFLAGS: -lwgpu_native -lm -ldl
```

These are all things I'll hopefully fix in a bit :)

If you want to install the Vulkan validation layer (`VK_LAYER_KHRONOS_validation`):

```console
pkg install vulkan-validation-layers
```

### Extra notes for aquaBSD

Good luck.
