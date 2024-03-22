package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type Texture struct {
	texture   *wgpu.Texture
	view      *wgpu.TextureView
	sampler   *wgpu.Sampler
	BindGroup *wgpu.BindGroup
}

func NewTextureFromPath(state State, label string, buf []byte) (*Texture, error) {
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

	if texture.sampler, err = state.device.CreateSampler(nil); err != nil {
		texture.texture.Release()
		texture.view.Release()
		return nil, err
	}

	if texture.BindGroup, err = state.device.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  fmt.Sprintf("Bind group (%s)", label),
		Layout: state.texture_bind_group_layout,
		Entries: []wgpu.BindGroupEntry{
			{ // texture
				Binding:     0,
				TextureView: texture.view,
			},
			{ // sampler
				Binding: 1,
				Sampler: texture.sampler,
			},
		},
	}); err != nil {
		texture.texture.Release()
		texture.view.Release()
		texture.sampler.Release()
		return nil, err
	}

	return texture, nil
}

func (texture *Texture) Release() {
	texture.texture.Release()
	texture.view.Release()
	texture.sampler.Release()
	texture.BindGroup.Release()
}
