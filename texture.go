package main

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"

	"github.com/rajveermalviya/go-webgpu/wgpu"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Texture struct {
	texture *wgpu.Texture
	view    *wgpu.TextureView
	sampler *wgpu.Sampler
}

func NewTextureFromPath(state *State, label string, buf []byte) (*Texture, error) {
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

	return texture, nil
}

func NewTextureFromText(state *State, label, text string) (*Texture, error) {
	img := textToImage(text)
	return NewTextureFromImage(state, label, img)
}

func textToImage(text string) image.Image {
	font := basicfont.Face7x13
	width := 0
	height := 13
	for _, c := range text {
		bounds, _, _ := font.GlyphBounds(c)
		width += (bounds.Max.X - bounds.Min.X).Ceil()
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	x := 0
	for _, c := range text {
		dot := fixed.P(x, 13)
		dr, mask, maskp, _, _ := font.Glyph(dot, c)
		draw.DrawMask(img, img.Bounds(), image.White, image.Point{}, mask, maskp, draw.Over)
		x += int(dr.Max.X - dr.Min.X)
	}

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
