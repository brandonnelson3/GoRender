package shaders

import (
	"fmt"
	"strings"

	"github.com/brandonnelson3/GoRender/gfx/uniforms"

	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	lineShaderOriginalVertexSourceFile = `lineshader.vert`
	lineShaderVertSrc                  = `
#version 450

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;

in vec3 vert;
in vec3 color;

out gl_PerVertex
{
	vec4 gl_Position;
	vec3 color;
} vertex_out;

void main() {
	gl_Position = projection * view * model * vec4(vert, 1);
	vertex_out.color = color;
}` + "\x00"
	lineShaderOriginalFragmentSourceFile = `lineshader.frag`
	lineShaderFragSrc                    = `
#version 450

in VERTEX_OUT
{
	vec4 gl_FragCoord;
	vec3 color;
} fragment_in;

out vec4 outputColor;

void main() {
	outputColor = vec4(fragment_in.color, 1);
}
` + "\x00"
)

// LineVertexShader is a VertexShader.
type LineVertexShader struct {
	uint32

	Projection, View, Model *uniforms.Matrix4
}

// NewLineVertexShader instantiates and initializes a shader object.
func NewLineVertexShader() (*LineVertexShader, error) {
	program := gl.CreateProgram()
	shader := gl.CreateShader(gl.VERTEX_SHADER)

	csources, free := gl.Strs(lineShaderVertSrc)
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

		return nil, fmt.Errorf("failed to compile %v: %v", lineShaderOriginalVertexSourceFile, log)
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

		return nil, fmt.Errorf("failed to link %v: %v", lineShaderOriginalVertexSourceFile, log)
	}

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))
	modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))

	gl.DeleteShader(shader)

	return &LineVertexShader{
		uint32:     program,
		Projection: uniforms.NewMatrix4(program, projectionLoc),
		View:       uniforms.NewMatrix4(program, viewLoc),
		Model:      uniforms.NewMatrix4(program, modelLoc),
	}, nil
}

// AddToPipeline adds this shader to the provided pipeline.
func (s *LineVertexShader) AddToPipeline(pipeline uint32) {
	gl.UseProgramStages(pipeline, gl.VERTEX_SHADER_BIT, s.uint32)
}

// BindVertexAttributes binds the attributes per vertex.
func (s *LineVertexShader) BindVertexAttributes() {
	vertAttrib := uint32(gl.GetAttribLocation(s.uint32, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	colorAttrib := uint32(gl.GetAttribLocation(s.uint32, gl.Str("color\x00")))
	gl.EnableVertexAttribArray(colorAttrib)
	gl.VertexAttribPointer(colorAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(12))
}

// LineFragmentShader represents a LineFragmentShader
type LineFragmentShader struct {
	uint32
}

// NewLineFragmentShader instantiates and initializes a LineFragmentShader object.
func NewLineFragmentShader() (*LineFragmentShader, error) {
	program := gl.CreateProgram()
	shader := gl.CreateShader(gl.FRAGMENT_SHADER)

	csources, free := gl.Strs(lineShaderFragSrc)
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

		return nil, fmt.Errorf("failed to compile %v: %v", lineShaderOriginalFragmentSourceFile, log)
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

		return nil, fmt.Errorf("failed to link %v: %v", lineShaderOriginalFragmentSourceFile, log)
	}

	gl.DeleteShader(shader)

	gl.BindFragDataLocation(program, 0, gl.Str("outputLine\x00"))

	fs := &LineFragmentShader{
		uint32: program,
	}
	return fs, nil
}

// AddToPipeline adds this shader to the provided pipeline.
func (s *LineFragmentShader) AddToPipeline(pipeline uint32) {
	gl.UseProgramStages(pipeline, gl.FRAGMENT_SHADER_BIT, s.uint32)
}
