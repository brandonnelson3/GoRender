package shaders

import (
	"fmt"
	"strings"

	"github.com/brandonnelson3/GoRender/gfx/uniforms"
	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	frustumShaderVertSrc = `
#version 450

uniform mat4 projection;
uniform mat4 view;

in vec3 vert;
in vec2 uv;

out vec2 uv_out;

void main() {
	gl_Position = projection * view * vec4(vert, 1);
	uv_out = uv;
}` + "\x00"

	frustumShaderFragSrc = `
#version 450

uniform vec3 color;
uniform float alpha;

out vec4 outputColor;

void main() {
	// Procedural screen-space crosshatch
	float lineDist = 10.0; // Distance between lines
	float thickness = 1.5; // Line thickness
	
	// Diagonal lines
	float val1 = mod(gl_FragCoord.x + gl_FragCoord.y, lineDist);
	float val2 = mod(gl_FragCoord.x - gl_FragCoord.y + 1000.0, lineDist);
	
	bool isLine = (val1 < thickness) || (val2 < thickness);
	
	float finalAlpha = (isLine ? 0.8 : 0.15) * alpha;
	outputColor = vec4(color, finalAlpha);
}` + "\x00"
)

type FrustumShader struct {
	shader
	Projection, View *uniforms.Matrix4
	Color            *uniforms.Vector3
	Alpha            *uniforms.Float
}

func NewFrustumShader() (*FrustumShader, error) {
	program := gl.CreateProgram()

	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	vertexSrc, freeVertexSrc := gl.Strs(frustumShaderVertSrc)
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
		return nil, fmt.Errorf("failed to compile frustum vert: %v", log)
	}
	gl.AttachShader(program, vertexShader)

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentSrc, freeFragmentSrc := gl.Strs(frustumShaderFragSrc)
	gl.ShaderSource(fragmentShader, 1, fragmentSrc, nil)
	freeFragmentSrc()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile frustum frag: %v", log)
	}
	gl.AttachShader(program, fragmentShader)

	gl.LinkProgram(program)
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to link frustum shader: %v", log)
	}

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))
	colorLoc := gl.GetUniformLocation(program, gl.Str("color\x00"))
	alphaLoc := gl.GetUniformLocation(program, gl.Str("alpha\x00"))

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return &FrustumShader{
		shader:     shader{program},
		Projection: uniforms.NewMatrix4(program, projectionLoc),
		View:       uniforms.NewMatrix4(program, viewLoc),
		Color:      uniforms.NewVector3(program, colorLoc),
		Alpha:      uniforms.NewFloat(program, alphaLoc),
	}, nil
}
