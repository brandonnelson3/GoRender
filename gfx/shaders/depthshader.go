package shaders

import (
	"fmt"
	"strings"

	"github.com/brandonnelson3/GoRender/gfx/uniforms"
	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	depthShaderOriginalVertexSourceFile = `depthshader.vert`
	depthShaderVertSrc                  = `
#version 450

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;

in vec3 vert;
in vec3 norm;
in vec2 uv;

out vec2 uv_out;

void main() {
	gl_Position = projection * view * model * vec4(vert, 1);
	uv_out = uv;
}` + "\x00"
	depthShaderOriginalFragmentSourceFile = `depthshader.frag`
	depthShaderFragSrc                    = `
#version 450

uniform sampler2D diffuse;

in vec2 uv_out;

void main() {
	vec4 color = textureLod(diffuse, uv_out, 0);
	if (color.a < 0.5) {
		discard;
	} 
}` + "\x00"
)

// DepthShader is a Shader.
type DepthShader struct {
	shader

	Projection, View, Model *uniforms.Matrix4

	Diffuse *uniforms.Sampler2D
}

// NewDepthShader instantiates and initializes a shader object.
func NewDepthShader() (*DepthShader, error) {
	program := gl.CreateProgram()

	// VertexShader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	vertexSrc, freeVertexSrc := gl.Strs(depthShaderVertSrc)
	gl.ShaderSource(vertexShader, 1, vertexSrc, nil)
	freeVertexSrc()
	gl.CompileShader(vertexShader)
	var status int32
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(vertexShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile %v: %v", depthShaderOriginalVertexSourceFile, log)
	}
	gl.AttachShader(program, vertexShader)

	// FragmentShader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentSrc, freeFragmentSrc := gl.Strs(depthShaderFragSrc)
	gl.ShaderSource(fragmentShader, 1, fragmentSrc, nil)
	freeFragmentSrc()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile %v: %v", depthShaderOriginalFragmentSourceFile, log)
	}
	gl.AttachShader(program, fragmentShader)

	// Linking
	gl.LinkProgram(program)
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to link %v: %v", depthShaderOriginalVertexSourceFile, log)
	}

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))
	modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))
	diffuseLoc := gl.GetUniformLocation(program, gl.Str("diffuse\x00"))

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return &DepthShader{
		shader:     shader{program},
		Projection: uniforms.NewMatrix4(program, projectionLoc),
		View:       uniforms.NewMatrix4(program, viewLoc),
		Model:      uniforms.NewMatrix4(program, modelLoc),
		Diffuse:    uniforms.NewSampler2D(program, diffuseLoc),
	}, nil
}
