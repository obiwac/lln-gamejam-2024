struct Text {
	x: f32,
	y: f32,
	scale_x: f32,
	scale_y: f32,
}

struct VertOut {
	@builtin(position) pos: vec4f,
	@location(0) colour: vec3f,
	@location(1) uv: vec2f,
};

@group(0) @binding(2)
var<uniform> text: Text;

@vertex
fn vert_main(
	@builtin(vertex_index) index: u32,
) -> VertOut {
	var out: VertOut;

	if index == 0u || index == 3u {
		out.pos = vec4(text.x - text.scale_x / 2., text.y - text.scale_y / 2., 0., 1.);
		out.colour = vec3(1., 1., 1.);
		out.uv = vec2(0., 0.);
	}

	if index == 1u {
		out.pos = vec4(text.x - text.scale_x / 2., text.y + text.scale_y / 2., 0., 1.);
		out.colour = vec3(1., 1., 1.);
		out.uv = vec2(0., 1.);
	}

	if index == 2u || index == 4u {
		out.pos = vec4(text.x + text.scale_x / 2., text.y + text.scale_y / 2., 0., 1.);
		out.colour = vec3(1., 1., 1.);
		out.uv = vec2(1., 1.);
	}

	if index == 5u {
		out.pos = vec4(text.x + text.scale_x / 2., text.y - text.scale_y / 2., 0., 1.);
		out.colour = vec3(1., 1., 1.);
		out.uv = vec2(1., 0.);
	}

	return out;
}

struct FragOut {
	@location(0) colour: vec4f,
};

@group(0) @binding(0)
var t: texture_2d<f32>;
@group(0) @binding(1)
var s: sampler;

@fragment
fn frag_main(vert: VertOut) -> FragOut {
	var out: FragOut;

	var tex_colour = textureSample(t, s, vec2(vert.uv.x, 1. - vert.uv.y));

	if tex_colour.a < 0.5 {
		discard;
	}

	out.colour = vec4f(tex_colour.rgb, 1.0) * vec4(vert.colour, 1.);

	return out;
}
