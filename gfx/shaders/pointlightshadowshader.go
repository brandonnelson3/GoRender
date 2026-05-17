package shaders

import (
	"fmt"
	"strings"

	"github.com/brandonnelson3/GoRender/gfx/uniforms"
	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	pointLightShadowShaderOriginalVertexSourceFile   = `pointlightshadowshader.vert`
	pointLightShadowShaderOriginalGeometrySourceFile = `pointlightshadowshader.geom`
	pointLightShadowShaderOriginalFragmentSourceFile = `pointlightshadowshader.frag`

	// Vertex shader: just transform vertex to world space for the geometry shader.
	pointLightShadowShaderVertSrc = `
#version 450

uniform mat4 model;
uniform int isInstanced;

layout(location = 0) in vec3 vert;
layout(location = 1) in vec3 norm;
layout(location = 2) in vec2 uv;
layout(location = 3) in mat4 instanceModel;

out vec2 uv_geom;
out vec3 worldPos_geom;

void main() {
	mat4 modelMat = (isInstanced != 0) ? instanceModel : model;
	vec4 worldPos = modelMat * vec4(vert, 1.0);
	worldPos_geom = worldPos.xyz;
	uv_geom = uv;
	gl_Position = worldPos;
}` + "\x00"

	// Geometry shader: emit triangles for a single light's cubemap (6 faces).
	pointLightShadowShaderGeomSrc = `
#version 450

layout(triangles) in;
layout(triangle_strip, max_vertices = 18) out;

uniform mat4 shadowMatrices[6];
uniform int shadowLightIndex;

in vec2 uv_geom[];
in vec3 worldPos_geom[];

out vec2 uv_frag;
out vec4 fragPos;

void main() {
	for (int face = 0; face < 6; ++face) {
		gl_Layer = shadowLightIndex * 6 + face;
		for (int i = 0; i < 3; ++i) {
			fragPos = gl_in[i].gl_Position;
			uv_frag = uv_geom[i];
			gl_Position = shadowMatrices[face] * fragPos;
			EmitVertex();
		}
		EndPrimitive();
	}
}` + "\x00"

	// Fragment shader: write linear depth as distance from the correct light.
	pointLightShadowShaderFragSrc = `
#version 450

uniform sampler2D diffuse;
uniform vec3 lightPos[4];
uniform int shadowLightIndex;
uniform float farPlane;

in vec2 uv_frag;
in vec4 fragPos;

void main() {
	vec4 color = textureLod(diffuse, uv_frag, 0);
	if (color.a < 0.5) {
		discard;
	}
	float lightDist = length(fragPos.xyz - lightPos[shadowLightIndex]);
	// Store linear depth normalised to [0,1] relative to farPlane.
	gl_FragDepth = lightDist / farPlane;
}` + "\x00"
)

// PointLightShadowShader renders scene depth into a cubemap for a single point light.
type PointLightShadowShader struct {
	shader

	Model            *uniforms.Matrix4
	IsInstanced      *uniforms.Int
	ShadowMatrices   *uniforms.Matrix4Array
	Diffuse          *uniforms.Sampler2D
	LightPos         *uniforms.Vector3Array
	ShadowLightIndex *uniforms.Int
	FarPlane         *uniforms.Float
}

// NewPointLightShadowShader compiles and links the shader, returning any errors.
func NewPointLightShadowShader() (*PointLightShadowShader, error) {
	program := gl.CreateProgram()

	// --- Vertex shader ---
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	vertexSrc, freeV := gl.Strs(pointLightShadowShaderVertSrc)
	gl.ShaderSource(vertexShader, 1, vertexSrc, nil)
	freeV()
	gl.CompileShader(vertexShader)
	var status int32
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(vertexShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile %v: %v", pointLightShadowShaderOriginalVertexSourceFile, log)
	}
	gl.AttachShader(program, vertexShader)

	// --- Geometry shader ---
	geomShader := gl.CreateShader(gl.GEOMETRY_SHADER)
	geomSrc, freeG := gl.Strs(pointLightShadowShaderGeomSrc)
	gl.ShaderSource(geomShader, 1, geomSrc, nil)
	freeG()
	gl.CompileShader(geomShader)
	gl.GetShaderiv(geomShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(geomShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(geomShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile %v: %v", pointLightShadowShaderOriginalGeometrySourceFile, log)
	}
	gl.AttachShader(program, geomShader)

	// --- Fragment shader ---
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentSrc, freeF := gl.Strs(pointLightShadowShaderFragSrc)
	gl.ShaderSource(fragmentShader, 1, fragmentSrc, nil)
	freeF()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile %v: %v", pointLightShadowShaderOriginalFragmentSourceFile, log)
	}
	gl.AttachShader(program, fragmentShader)

	// --- Linking ---
	gl.LinkProgram(program)
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to link %v: %v", pointLightShadowShaderOriginalVertexSourceFile, log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(geomShader)
	gl.DeleteShader(fragmentShader)

	modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))
	isInstancedLoc := gl.GetUniformLocation(program, gl.Str("isInstanced\x00"))
	shadowMatricesLoc := gl.GetUniformLocation(program, gl.Str("shadowMatrices\x00"))
	diffuseLoc := gl.GetUniformLocation(program, gl.Str("diffuse\x00"))
	lightPosLoc := gl.GetUniformLocation(program, gl.Str("lightPos\x00"))
	shadowLightIndexLoc := gl.GetUniformLocation(program, gl.Str("shadowLightIndex\x00"))
	farPlaneLoc := gl.GetUniformLocation(program, gl.Str("farPlane\x00"))

	return &PointLightShadowShader{
		shader:           shader{program},
		Model:            uniforms.NewMatrix4(program, modelLoc),
		IsInstanced:      uniforms.NewInt(program, isInstancedLoc),
		ShadowMatrices:   uniforms.NewMatrix4Array(program, shadowMatricesLoc),
		Diffuse:          uniforms.NewSampler2D(program, diffuseLoc),
		LightPos:         uniforms.NewVector3Array(program, lightPosLoc),
		ShadowLightIndex: uniforms.NewInt(program, shadowLightIndexLoc),
		FarPlane:         uniforms.NewFloat(program, farPlaneLoc),
	}, nil
}
