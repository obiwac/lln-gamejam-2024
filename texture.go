package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"github.com/rajveermalviya/go-webgpu/wgpu"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Texture struct {
	texture *wgpu.Texture
	view    *wgpu.TextureView
	sampler *wgpu.Sampler
}

func NewTextureFromBytes(state *State, label string, buf []byte) (*Texture, error) {
	img, err := png.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()

	rgba, ok := img.(*image.RGBA)
	if !ok {
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, image.Point{}, draw.Over)
	}

	width := uint32(bounds.Dx())
	height := uint32(bounds.Dy())

	size := wgpu.Extent3D{
		Width:              width,
		Height:             height,
		DepthOrArrayLayers: 1,
	}

	texture := &Texture{}

	if texture.texture, err = state.device.CreateTexture(&wgpu.TextureDescriptor{
		Label:         label,
		Size:          size,
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     wgpu.TextureDimension_2D,
		Format:        wgpu.TextureFormat_RGBA8UnormSrgb,
		Usage:         wgpu.TextureUsage_TextureBinding | wgpu.TextureUsage_CopyDst,
	}); err != nil {
		return nil, err
	}

	if err = state.queue.WriteTexture(
		&wgpu.ImageCopyTexture{
			Aspect:   wgpu.TextureAspect_All,
			Texture:  texture.texture,
			MipLevel: 0,
			Origin:   wgpu.Origin3D{},
		},
		rgba.Pix,
		&wgpu.TextureDataLayout{
			Offset:       0,
			BytesPerRow:  4 * width,
			RowsPerImage: height,
		},
		&size,
	); err != nil {
		texture.texture.Release()
		return nil, err
	}

	if texture.view, err = texture.texture.CreateView(nil); err != nil {
		texture.texture.Release()
		return nil, err
	}

	if texture.sampler, err = state.device.CreateSampler(&wgpu.SamplerDescriptor{
		AddressModeU:   wgpu.AddressMode_ClampToEdge,
		AddressModeV:   wgpu.AddressMode_ClampToEdge,
		AddressModeW:   wgpu.AddressMode_ClampToEdge,
		MagFilter:      wgpu.FilterMode_Linear,
		MinFilter:      wgpu.FilterMode_Linear,
		MipmapFilter:   wgpu.MipmapFilterMode_Linear,
		MaxAnisotrophy: 1,
	}); err != nil {
		texture.texture.Release()
		texture.view.Release()
		return nil, err
	}

	return texture, nil
}

const DEPTH_FORMAT = wgpu.TextureFormat_Depth32Float

func NewDepthTexture(state *State) (*Texture, error) {
	width, height := state.win.GetSize()

	size := wgpu.Extent3D{
		Width:              uint32(width),
		Height:             uint32(height),
		DepthOrArrayLayers: 1,
	}

	texture := &Texture{}
	var err error

	if texture.texture, err = state.device.CreateTexture(&wgpu.TextureDescriptor{
		Label:         "Depth",
		Size:          size,
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     wgpu.TextureDimension_2D,
		Format:        DEPTH_FORMAT,
		Usage:         wgpu.TextureUsage_RenderAttachment | wgpu.TextureUsage_TextureBinding,
	}); err != nil {
		return nil, err
	}

	if texture.view, err = texture.texture.CreateView(nil); err != nil {
		texture.texture.Release()
		return nil, err
	}

	if texture.sampler, err = state.device.CreateSampler(&wgpu.SamplerDescriptor{
		AddressModeU:   wgpu.AddressMode_ClampToEdge,
		AddressModeV:   wgpu.AddressMode_ClampToEdge,
		AddressModeW:   wgpu.AddressMode_ClampToEdge,
		MagFilter:      wgpu.FilterMode_Nearest,
		MinFilter:      wgpu.FilterMode_Nearest,
		MipmapFilter:   wgpu.MipmapFilterMode_Nearest,
		Compare:        wgpu.CompareFunction_Less,
		MaxAnisotrophy: 1,
		LodMinClamp:    0,
		LodMaxClamp:    100,
	}); err != nil {
		texture.texture.Release()
		texture.view.Release()
		return nil, err
	}

	return texture, nil
}

func NewTextureFromText(state *State, label, text string) (*Texture, error) {
	img := textToImage(text)
	return NewTextureFromImage(state, label, img)
}

func textToImage(text string) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 300, 100))
	point := fixed.Point26_6{X: fixed.I(0), Y: fixed.I(13)}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: basicfont.Face7x13,
		Dot:  point,
	}

	drawer.DrawString(text)
	return img
}

func NewTextureFromImage(state *State, label string, img image.Image) (*Texture, error) {
	bounds := img.Bounds()

	rgba, ok := img.(*image.RGBA)
	if !ok {
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, image.Point{}, draw.Over)
	}

	width := uint32(bounds.Dx())
	height := uint32(bounds.Dy())

	size := wgpu.Extent3D{
		Width:              width,
		Height:             height,
		DepthOrArrayLayers: 1,
	}

	texture := &Texture{}

	var err error
	if texture.texture, err = state.device.CreateTexture(&wgpu.TextureDescriptor{
		Label:         label,
		Size:          size,
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     wgpu.TextureDimension_2D,
		Format:        wgpu.TextureFormat_RGBA8UnormSrgb,
		Usage:         wgpu.TextureUsage_TextureBinding | wgpu.TextureUsage_CopyDst,
	}); err != nil {
		return nil, err
	}

	if err = state.queue.WriteTexture(
		&wgpu.ImageCopyTexture{
			Aspect:   wgpu.TextureAspect_All,
			Texture:  texture.texture,
			MipLevel: 0,
			Origin:   wgpu.Origin3D{},
		},
		rgba.Pix,
		&wgpu.TextureDataLayout{
			Offset:       0,
			BytesPerRow:  4 * width,
			RowsPerImage: height,
		},
		&size,
	); err != nil {
		texture.texture.Release()
		return nil, err
	}

	if texture.view, err = texture.texture.CreateView(nil); err != nil {
		texture.texture.Release()
		return nil, err
	}

	if texture.sampler, err = state.device.CreateSampler(nil); err != nil {
		texture.texture.Release()
		texture.view.Release()
		return nil, err
	}

	return texture, nil
}

func (texture *Texture) Release() {
	texture.texture.Release()
	texture.view.Release()
	texture.sampler.Release()
}
