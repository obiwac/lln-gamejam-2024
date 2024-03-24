package main

import (
	"unsafe"

	"github.com/rajveermalviya/go-webgpu/wgpu"
)

type TextAttrs struct {
	x, y  float32
	scale, scale_y float32
}

type Text struct {
	state *State
	bind_group *wgpu.BindGroup
	texture     *Texture
	text_attrs_buf *wgpu.Buffer
}

func NewText(state *State, content string, x, y, scale_x, scale_y float32) (*Text, error) {
	text := &Text{state: state}
	var err error

	if text.texture, err = NewTextureFromText(state, content, content); err != nil {
		return nil, err
	}

	if text.text_attrs_buf, err = state.device.CreateBuffer(&wgpu.BufferDescriptor{
		Size:  uint64(unsafe.Sizeof(TextAttrs{})),
		Usage: wgpu.BufferUsage_Uniform | wgpu.BufferUsage_CopyDst,
	}); err != nil {
		text.texture.Release()
		return nil, err
	}

	if text.bind_group, err = state.device.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Layout: state.text_pipeline.bind_group_layout,
		Entries: []wgpu.BindGroupEntry{
			{
				Binding: 0,
				TextureView: text.texture.view,
			},
			{
				Binding: 1,
				Sampler: text.texture.sampler,
			},
			{
				Binding: 2,
				Buffer: text.text_attrs_buf,
				Size: wgpu.WholeSize,
			},
		},
	}); err != nil {
		text.texture.Release()
		text.text_attrs_buf.Release()
		return nil, err
	}

	attrs := TextAttrs{
		x: x, y: y,
		scale: scale_x, scale_y: scale_y,
	}

	var attrs_bytes []byte

	for i := 0; i < int(unsafe.Sizeof(attrs)); i++ {
		byte := (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&attrs))+uintptr(i)))
		attrs_bytes = append(attrs_bytes, *byte)
	}

	state.queue.WriteBuffer(text.text_attrs_buf, 0, attrs_bytes)

	return text, nil
}

func (text *Text) Draw(render_pass *wgpu.RenderPassEncoder) {
	text.state.text_pipeline.Set(render_pass, text.bind_group)
	render_pass.Draw(6, 1, 0, 0)
}

func (text *Text) Release() {
	text.texture.Release()
	text.text_attrs_buf.Release()
	text.bind_group.Release()
}
