package shaders

import (
	"fmt"
	"strings"

	"github.com/brandonnelson3/GoRender/gfx/uniforms"
	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	pipShaderOriginalVertexSourceFile = `pipshader.vert`
	pipShaderVertSrc                  = `
	#version 450
	
	in vec2 pos;
	in vec2 uv;
	
	uniform mat4 projection;
	
	out gl_PerVertex
	{
		vec4 gl_Position;
		vec2 uv;
	} vertex_out;
	
	void main() {
		gl_Position = projection * vec4(pos, 0, 1);
		vertex_out.uv = uv;
	}` + "\x00"
	pipShaderOriginalFragmentSourceFile = `pipshader.vert`
	pipShaderFragSrc                    = `
#version 450

uniform sampler2D textureSampler;
uniform vec2 nearFar;

in VERTEX_OUT
{
    vec4 gl_Position;
	vec2 uv;
} fragment_in;

out vec4 outputColor;

void main() {
	float n = nearFar.x;
	float f = nearFar.y;
	float z = texture(textureSampler, fragment_in.uv).r;
	float depth = (2.0 * n) / (f + n - z * (f - n));	

	outputColor = vec4(vec3(z), 1.0);
}` + "\x00"
)

// PipVertexShader is a VertexShader.
type PipVertexShader struct {
	uint32

	Projection *uniforms.Matrix4
}

// NewPipVertexShader instantiates and initializes a PipVertexShader object.
func NewPipVertexShader() (*PipVertexShader, error) {
	program := gl.CreateProgram()
	shader := gl.CreateShader(gl.VERTEX_SHADER)

	csources, free := gl.Strs(pipShaderVertSrc)
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

		return nil, fmt.Errorf("failed to compile %v: %v", pipShaderOriginalVertexSourceFile, log)
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

		return nil, fmt.Errorf("failed to link %v: %v", pipShaderOriginalVertexSourceFile, log)
	}

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))

	gl.DeleteShader(shader)

	return &PipVertexShader{
		uint32:     program,
		Projection: uniforms.NewMatrix4(program, projectionLoc),
	}, nil
}

// BindVertexAttributes binds the attributes per vertex.
func (s *PipVertexShader) BindVertexAttributes() {
	posAttrib := uint32(gl.GetAttribLocation(s.uint32, gl.Str("pos\x00")))
	gl.EnableVertexAttribArray(posAttrib)
	gl.VertexAttribPointer(posAttrib, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	uvAttrib := uint32(gl.GetAttribLocation(s.uint32, gl.Str("uv\x00")))
	gl.EnableVertexAttribArray(uvAttrib)
	gl.VertexAttribPointer(uvAttrib, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(8))
}

// AddToPipeline adds this shader to the provided pipeline.
func (s *PipVertexShader) AddToPipeline(pipeline uint32) {
	gl.UseProgramStages(pipeline, gl.VERTEX_SHADER_BIT, s.uint32)
}

// PipFragmentShader represents a FragmentShader
type PipFragmentShader struct {
	uint32

	DepthMap *uniforms.Sampler2D
	NearFar  *uniforms.Vector2
}

// NewPipFragmentShader instantiates and initializes a PipFragmentShader object.
func NewPipFragmentShader() (*PipFragmentShader, error) {
	program := gl.CreateProgram()
	shader := gl.CreateShader(gl.FRAGMENT_SHADER)

	csources, free := gl.Strs(pipShaderFragSrc)
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

		return nil, fmt.Errorf("failed to compile %v: %v", pipShaderOriginalFragmentSourceFile, log)
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

		return nil, fmt.Errorf("failed to link %v: %v", pipShaderOriginalFragmentSourceFile, log)
	}

	depthMapLoc := gl.GetUniformLocation(program, gl.Str("textureSampler\x00"))
	nearFarLoc := gl.GetUniformLocation(program, gl.Str("nearFar\x00"))

	gl.DeleteShader(shader)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	return &PipFragmentShader{
		uint32:   program,
		DepthMap: uniforms.NewSampler2D(program, depthMapLoc),
		NearFar:  uniforms.NewVector2(program, nearFarLoc),
	}, nil
}

// AddToPipeline adds this shader to the provided pipeline.
func (s *PipFragmentShader) AddToPipeline(pipeline uint32) {
	gl.UseProgramStages(pipeline, gl.FRAGMENT_SHADER_BIT, s.uint32)
}
