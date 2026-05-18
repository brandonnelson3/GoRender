package shaders

import (
	"fmt"
	"strings"

	"github.com/brandonnelson3/GoRender/gfx/uniforms"
	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	hudShaderVertSrc = `
	#version 450
	
	in vec2 pos;
	in vec2 uv;
	
	uniform mat4 projection;
	
	out vec2 uv_out;
	
	void main() {
		gl_Position = projection * vec4(pos, 0.0, 1.0);
		uv_out = uv;
	}` + "\x00"

	hudShaderFragSrc = `
	#version 450
	
	uniform sampler2D hudTexture;
	
	in vec2 uv_out;
	
	out vec4 outputColor;
	
	void main() {
		outputColor = texture(hudTexture, uv_out);
	}` + "\x00"
)

// HudShader is a shader for rendering full-screen or positioned 2D overlays.
type HudShader struct {
	shader

	Projection *uniforms.Matrix4
	HudTexture *uniforms.Sampler2D
}

// NewHudShader compiles the shaders and initializes uniform locations.
func NewHudShader() (*HudShader, error) {
	program := gl.CreateProgram()

	// Compile Vertex Shader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	vertexSrc, freeVertexSrc := gl.Strs(hudShaderVertSrc)
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
		return nil, fmt.Errorf("failed to compile hud vertex shader: %v", log)
	}
	gl.AttachShader(program, vertexShader)

	// Compile Fragment Shader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentSrc, freeFragmentSrc := gl.Strs(hudShaderFragSrc)
	gl.ShaderSource(fragmentShader, 1, fragmentSrc, nil)
	freeFragmentSrc()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile hud fragment shader: %v", log)
	}
	gl.AttachShader(program, fragmentShader)

	// Link Program
	gl.LinkProgram(program)
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to link hud shader program: %v", log)
	}

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	hudTextureLoc := gl.GetUniformLocation(program, gl.Str("hudTexture\x00"))

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	return &HudShader{
		shader:     shader{program},
		Projection: uniforms.NewMatrix4(program, projectionLoc),
		HudTexture: uniforms.NewSampler2D(program, hudTextureLoc),
	}, nil
}
