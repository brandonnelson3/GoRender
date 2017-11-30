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

in vec3 vert;
in vec3 color;

out vec3 color_out;

void main() {
	gl_Position = projection * view * vec4(vert, 1);
	color_out = color;
}` + "\x00"
	lineShaderOriginalFragmentSourceFile = `lineshader.frag`
	lineShaderFragSrc                    = `
#version 450

in vec3 color_out;

out vec4 outputColor;

void main() {
	outputColor = vec4(color_out, 1);
}
` + "\x00"
)

// LineShader is a Shader.
type LineShader struct {
	uint32

	Projection, View *uniforms.Matrix4
}

// NewLineShader instantiates and initializes a shader object.
func NewLineShader() (*LineShader, error) {
	program := gl.CreateProgram()

	// VertexShader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	vertexSrc, freeVertexSrc := gl.Strs(lineShaderVertSrc)
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
		return nil, fmt.Errorf("failed to compile %v: %v", lineShaderOriginalVertexSourceFile, log)
	}
	gl.AttachShader(program, vertexShader)

	// FragmentShader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentSrc, freeFragmentSrc := gl.Strs(lineShaderFragSrc)
	gl.ShaderSource(fragmentShader, 1, fragmentSrc, nil)
	freeFragmentSrc()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile %v: %v", lineShaderOriginalFragmentSourceFile, log)
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
		return nil, fmt.Errorf("failed to link %v: %v", lineShaderOriginalVertexSourceFile, log)
	}

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	return &LineShader{
		uint32:     program,
		Projection: uniforms.NewMatrix4(program, projectionLoc),
		View:       uniforms.NewMatrix4(program, viewLoc),
	}, nil
}

// Program returns the opengl program id of this vertex shader.
func (s *LineShader) Program() uint32 {
	return s.uint32
}

// Use binds this program to be used.
func (s *LineShader) Use() {
	gl.UseProgram(s.uint32)
}
