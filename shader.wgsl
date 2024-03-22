struct VertOut {
	@builtin(position) pos: vec4f,
	@location(0) colour: vec3f,
	@location(1) uv: vec2f,
};

@vertex
fn vert_main(@builtin(vertex_index) index: u32) -> VertOut {
	var out: VertOut;

	if index == 0u {
		out.pos = vec4(0., -.5, 0., 1.);
		out.colour = vec3(1., 0., 0.);
		out.uv = vec2(.5, 1.);
	}

	if index == 1u {
		out.pos = vec4(-.5, .5, 0., 1.);
		out.colour = vec3(0., 1., 0.);
		out.uv = vec2(0., 0.);
	}

	if index == 2u {
		out.pos = vec4(.5, .5, 0., 1.);
		out.colour = vec3(0., 0., 1.);
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

	out.colour = textureSample(t, s, vert.uv) * vec4(vert.colour, 1.);

	return out;
}
