struct VertOut {
	@builtin(position) pos: vec4f,
	@location(0) colour: vec3f,
	@location(1) uv: vec2f,
};

@group(0) @binding(2)
var<uniform> mvp: mat4x4<f32>;

@vertex
fn vert_main(
	@location(0) pos: vec3f,
	@location(1) uv: vec2f,
) -> VertOut {
	var out: VertOut;

	out.pos = mvp * vec4(pos, 1.);
	out.colour = vec3(1., 1., 1.);
	out.uv = uv;

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

	var tex_colour = textureSample(t, s, vert.uv.yx);

	out.colour = vec4f(tex_colour.rgb, 1.0) * vec4(vert.colour, 1.);

	return out;
}
