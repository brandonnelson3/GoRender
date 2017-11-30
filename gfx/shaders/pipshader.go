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
	
	out vec2 uv_out;
	
	void main() {
		gl_Position = projection * vec4(pos, 0, 1);
		uv_out = uv;
	}` + "\x00"
	pipShaderOriginalFragmentSourceFile = `pipshader.vert`
	pipShaderFragSrc                    = `
#version 450

uniform sampler2D textureSampler;
uniform vec2 nearFar;

in vec2 uv_out;

out vec4 outputColor;

void main() {
	float n = nearFar.x;
	float f = nearFar.y;
	float z = texture(textureSampler, uv_out).r;
	float depth = (2.0 * n) / (f + n - z * (f - n));	

	outputColor = vec4(vec3(z), 1.0);
}` + "\x00"
)

// PipShader is a Shader.
type PipShader struct {
	uint32

	Projection *uniforms.Matrix4

	DepthMap *uniforms.Sampler2D
	NearFar  *uniforms.Vector2
}

// NewPipShader instantiates and initializes a Shader object.
func NewPipShader() (*PipShader, error) {
	program := gl.CreateProgram()

	// VertexShader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	vertexSrc, freeVertexSrc := gl.Strs(pipShaderVertSrc)
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
		return nil, fmt.Errorf("failed to compile %v: %v", pipShaderOriginalVertexSourceFile, log)
	}
	gl.AttachShader(program, vertexShader)

	// FragmentShader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentSrc, freeFragmentSrc := gl.Strs(pipShaderFragSrc)
	gl.ShaderSource(fragmentShader, 1, fragmentSrc, nil)
	freeFragmentSrc()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile %v: %v", pipShaderOriginalFragmentSourceFile, log)
	}
	gl.AttachShader(program, fragmentShader)

	//Linking
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
	depthMapLoc := gl.GetUniformLocation(program, gl.Str("textureSampler\x00"))
	nearFarLoc := gl.GetUniformLocation(program, gl.Str("nearFar\x00"))

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	return &PipShader{
		uint32:     program,
		Projection: uniforms.NewMatrix4(program, projectionLoc),
		DepthMap:   uniforms.NewSampler2D(program, depthMapLoc),
		NearFar:    uniforms.NewVector2(program, nearFarLoc),
	}, nil
}

// Program returns the opengl program id of this vertex shader.
func (s *PipShader) Program() uint32 {
	return s.uint32
}

// Use binds this program to be used.
func (s *PipShader) Use() {
	gl.UseProgram(s.uint32)
}
