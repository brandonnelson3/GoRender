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

out gl_PerVertex
{
	vec4 gl_Position;	
	vec2 uv;
} vertex_out;

void main() {
	gl_Position = projection * view * model * vec4(vert, 1);
	vertex_out.uv = uv;
}` + "\x00"
	depthShaderOriginalFragmentSourceFile = `depthshader.frag`
	depthShaderFragSrc                    = `
#version 450

uniform sampler2D diffuse;

in VERTEX_OUT
{
	vec4 gl_FragCoord;
	vec2 uv;
} fragment_in;

void main() {
	vec4 color = texture(diffuse, fragment_in.uv);
	if (color.a != 1) {
		discard;
	} 
}` + "\x00"
)

// DepthVertexShader is a VertexShader.
type DepthVertexShader struct {
	uint32

	Projection, View, Model *uniforms.Matrix4
}

// NewDepthVertexShader instantiates and initializes a shader object.
func NewDepthVertexShader() (*DepthVertexShader, error) {
	program := gl.CreateProgram()
	shader := gl.CreateShader(gl.VERTEX_SHADER)

	csources, free := gl.Strs(depthShaderVertSrc)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return nil, fmt.Errorf("failed to compile %v: %v", depthShaderOriginalVertexSourceFile, log)
	}

	gl.AttachShader(program, shader)
	gl.ProgramParameteri(program, gl.PROGRAM_SEPARABLE, gl.TRUE)
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

	gl.DeleteShader(shader)

	return &DepthVertexShader{
		uint32:     program,
		Projection: uniforms.NewMatrix4(program, projectionLoc),
		View:       uniforms.NewMatrix4(program, viewLoc),
		Model:      uniforms.NewMatrix4(program, modelLoc),
	}, nil
}

// AddToPipeline adds this shader to the provided pipeline.
func (s *DepthVertexShader) AddToPipeline(pipeline uint32) {
	gl.UseProgramStages(pipeline, gl.VERTEX_SHADER_BIT, s.uint32)
}

// BindVertexAttributes binds the attributes per vertex.
func (s *DepthVertexShader) BindVertexAttributes() {
	vertAttrib := uint32(gl.GetAttribLocation(s.uint32, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	normAttrib := uint32(gl.GetAttribLocation(s.uint32, gl.Str("norm\x00")))
	gl.EnableVertexAttribArray(normAttrib)
	gl.VertexAttribPointer(normAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(12))
	uvAttrib := uint32(gl.GetAttribLocation(s.uint32, gl.Str("uv\x00")))
	gl.EnableVertexAttribArray(uvAttrib)
	gl.VertexAttribPointer(uvAttrib, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(24))
}

// DepthFragmentShader represents a FragmentShader
type DepthFragmentShader struct {
	uint32

	Diffuse *uniforms.Sampler2D
}

// NewDepthFragmentShader instantiates and initializes a DepthFragmentShader object.
func NewDepthFragmentShader() (*DepthFragmentShader, error) {
	program := gl.CreateProgram()
	shader := gl.CreateShader(gl.FRAGMENT_SHADER)

	csources, free := gl.Strs(depthShaderFragSrc)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return nil, fmt.Errorf("failed to compile %v: %v", depthShaderOriginalFragmentSourceFile, log)
	}

	gl.AttachShader(program, shader)
	gl.ProgramParameteri(program, gl.PROGRAM_SEPARABLE, gl.TRUE)
	gl.LinkProgram(program)

	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return nil, fmt.Errorf("failed to link %v: %v", depthShaderOriginalFragmentSourceFile, log)
	}

	diffuseLoc := gl.GetUniformLocation(program, gl.Str("diffuse\x00"))

	gl.DeleteShader(shader)

	return &DepthFragmentShader{
		uint32:  program,
		Diffuse: uniforms.NewSampler2D(program, diffuseLoc),
	}, nil
}

// AddToPipeline adds this shader to the provided pipeline.
func (s *DepthFragmentShader) AddToPipeline(pipeline uint32) {
	gl.UseProgramStages(pipeline, gl.FRAGMENT_SHADER_BIT, s.uint32)
}
