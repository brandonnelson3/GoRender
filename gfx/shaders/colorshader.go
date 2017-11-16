package shaders

import (
	"fmt"
	"strings"

	"github.com/brandonnelson3/GoRender/gfx/buffers"
	"github.com/brandonnelson3/GoRender/gfx/uniforms"
	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	colorShaderOriginalVertexSourceFile = `colorshader.vert`
	colorShaderVertSrc                  = `
#version 450

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;
uniform mat4 lightViewProj1;
uniform mat4 lightViewProj2;
uniform mat4 lightViewProj3;

in vec3 vert;
in vec3 norm;
in vec2 uv;

out gl_PerVertex
{
	vec4 gl_Position;
	vec4 position;
	vec3 worldPosition;
	vec3 normal;
	vec2 uv;	
	vec4 lightPosition1;
	vec4 lightPosition2;
	vec4 lightPosition3;
} vertex_out;

void main() {
	gl_Position = projection * view * model * vec4(vert, 1);
	vertex_out.position = projection * view * model * vec4(vert, 1);
	vertex_out.worldPosition = vec3(model * vec4(vert, 1));
	vertex_out.normal = norm;
	vertex_out.uv = uv;
	vertex_out.lightPosition1 = lightViewProj1 * model * vec4(vert, 1);
	vertex_out.lightPosition2 = lightViewProj2 * model * vec4(vert, 1);
	vertex_out.lightPosition3 = lightViewProj3 * model * vec4(vert, 1);
}` + "\x00"
	colorShaderOriginalFragmentSourceFile = `colorshader.frag`
	colorShaderFragSrc                    = `
#version 450

// TODO: Probably can pull this out into a common place.
struct PointLight {
	vec3 color;
	float intensity;
	vec3 position;
	float radius;
};

struct VisibleIndex {
	int index;
};

struct DirectionalLight {
	vec3 color;
	float brightness;
	vec3 direction;
};

// Shader storage buffer objects
layout(std430, binding = 0) readonly buffer LightBuffer {
	PointLight data[];
} lightBuffer;

layout(std430, binding = 1) readonly buffer VisibleLightIndicesBuffer {
	VisibleIndex data[];
} visibleLightIndicesBuffer;

layout(std430, binding = 2) readonly buffer DirectionalLightBuffer {
	DirectionalLight data;
} directionalLightBuffer;

uniform int renderMode;
uniform uint numTilesX;
uniform float zNear;
uniform float zFar;
uniform vec3 ambientLightColor;
uniform sampler2D diffuse;
uniform sampler2D shadowMap1;
uniform sampler2D shadowMap2;
uniform sampler2D shadowMap3;

in VERTEX_OUT
{
	vec4 gl_FragCoord;
	vec4 position;
	vec3 worldPosition;
	vec3 normal;
	vec2 uv;	
	vec4 lightPosition1;
	vec4 lightPosition2;
	vec4 lightPosition3;
} fragment_in;

out vec4 outputColor;

float linearize(float depth)
{
	return (2 * zNear) / (zFar + zNear - depth * (zFar - zNear));
}

vec3 saturate(vec3 v) {
	return vec3(clamp(v.x, 0.0, 1.0), clamp(v.y, 0.0, 1.0), clamp(v.z, 0.0, 1.0));
}

float getShadowFactor(int index, vec3 projCoords)
{
	float shadowMapDepth = 0.0f;
    if(index == 0) {
		shadowMapDepth = texture(shadowMap1, projCoords.xy).x;
    } else if(index == 1) {
		shadowMapDepth = texture(shadowMap2, projCoords.xy).x;
    } else {
		shadowMapDepth = texture(shadowMap3, projCoords.xy).x;
	}
	
	float currentDepth = projCoords.z;
	if (linearize(currentDepth-0.0005) > linearize(shadowMapDepth)) {
		return 0.0f;
	}
	return 1.0f;
}

void main() {
	ivec2 location = ivec2(gl_FragCoord.xy);
	// TODO: Put this 16 somewhere constant.
	ivec2 tileID = location / ivec2(16, 16);
	uint index = tileID.y * numTilesX + tileID.x;

	// TODO 1024 should be somewhere constant.
	uint offset = index * 1024;
	
	if (renderMode == 0 || renderMode == 5) {
		vec3 pointLightColor = vec3(0, 0, 0);

		uint i=0;
		for (i; i < 1024 && visibleLightIndicesBuffer.data[offset + i].index != -1; i++) {
			uint lightIndex = visibleLightIndicesBuffer.data[offset + i].index;
			PointLight light = lightBuffer.data[lightIndex];
			vec3 lightVector = light.position - fragment_in.worldPosition;
			float dist = length(lightVector);
			float NdL = max(0.0f, dot(fragment_in.normal, lightVector*(1.0f/dist)));
			float attenuation = 1.0f - clamp(dist * (1.0/(light.radius)), 0.0, 1.0);
			vec3 diffuse = NdL * light.color * light.intensity;
			pointLightColor += attenuation * diffuse;
		}
		
		DirectionalLight directionalLight = directionalLightBuffer.data;
		float NdL = max(0.0f, dot(fragment_in.normal, -1*directionalLight.direction));
		vec3 directionalLightColor = (NdL) * directionalLight.color * directionalLight.brightness;
		
		float inputPositionInv = 1.0/fragment_in.position.w;
		float lightPositionInv1 = 1.0/fragment_in.lightPosition1.w;
		float lightPositionInv2 = 1.0/fragment_in.lightPosition2.w;
		float lightPositionInv3 = 1.0/fragment_in.lightPosition3.w;

		float depthTest = fragment_in.position.z;

		vec3 shadowCoords[3] = vec3[](
			fragment_in.lightPosition1 * 0.5 + 0.5, 
			fragment_in.lightPosition2 * 0.5 + 0.5, 
			fragment_in.lightPosition3 * 0.5 + 0.5
		);

		int shadowIndex = 3;
		vec3 shadowIndexColor = vec3(1, 1, 1);
		if (depthTest < 15) {			
			shadowIndex = 0;
			shadowIndexColor = vec3(1, .5, .5);
		} else if(depthTest < 100) {
			shadowIndex = 1;
			shadowIndexColor = vec3(.5, 1, .5);
		} else if(depthTest < 500) {
			shadowIndex = 2;
			shadowIndexColor = vec3(.5, .5, 1);
		}

		if (renderMode == 0) {
			shadowIndexColor = vec3(1, 1, 1);
		}
		
		float shadowFactor = 1.0f;		
		if (shadowIndex != 3) {
			shadowFactor = getShadowFactor(shadowIndex, shadowCoords[shadowIndex]);
		} 
		
		outputColor = texture(diffuse, fragment_in.uv) * vec4(shadowIndexColor, 1.0) * vec4(saturate(pointLightColor + directionalLightColor*shadowFactor + ambientLightColor), 1.0);
	} else if (renderMode == 1) {
		uint i=0;
		for (i; i < 1024 && visibleLightIndicesBuffer.data[offset + i].index != -1; i++) {}
		outputColor = vec4(vec3(float(i)/256)+vec3(0.1), 1.0);
	} else if (renderMode == 2) {
		outputColor = vec4(abs(fragment_in.normal), 1.0);
	} else if (renderMode == 3) {
		outputColor = vec4(fragment_in.uv, 0, 1.0);
	} else if (renderMode == 4) {
		outputColor = texture(diffuse, fragment_in.uv);
	} 
}
` + "\x00"
)

// ColorVertexShader is a VertexShader.
type ColorVertexShader struct {
	uint32

	Projection, View, Model, LightViewProj1, LightViewProj2, LightViewProj3 *uniforms.Matrix4
}

// NewColorVertexShader instantiates and initializes a shader object.
func NewColorVertexShader() (*ColorVertexShader, error) {
	program := gl.CreateProgram()
	shader := gl.CreateShader(gl.VERTEX_SHADER)

	csources, free := gl.Strs(colorShaderVertSrc)
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

		return nil, fmt.Errorf("failed to compile %v: %v", colorShaderOriginalVertexSourceFile, log)
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

		return nil, fmt.Errorf("failed to link %v: %v", colorShaderOriginalVertexSourceFile, log)
	}

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))
	modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))
	lightViewProj1Loc := gl.GetUniformLocation(program, gl.Str("lightViewProj1\x00"))
	lightViewProj2Loc := gl.GetUniformLocation(program, gl.Str("lightViewProj2\x00"))
	lightViewProj3Loc := gl.GetUniformLocation(program, gl.Str("lightViewProj3\x00"))

	gl.DeleteShader(shader)

	return &ColorVertexShader{
		uint32:         program,
		Projection:     uniforms.NewMatrix4(program, projectionLoc),
		View:           uniforms.NewMatrix4(program, viewLoc),
		Model:          uniforms.NewMatrix4(program, modelLoc),
		LightViewProj1: uniforms.NewMatrix4(program, lightViewProj1Loc),
		LightViewProj2: uniforms.NewMatrix4(program, lightViewProj2Loc),
		LightViewProj3: uniforms.NewMatrix4(program, lightViewProj3Loc),
	}, nil
}

// AddToPipeline adds this shader to the provided pipeline.
func (s *ColorVertexShader) AddToPipeline(pipeline uint32) {
	gl.UseProgramStages(pipeline, gl.VERTEX_SHADER_BIT, s.uint32)
}

// BindVertexAttributes binds the attributes per vertex.
func (s *ColorVertexShader) BindVertexAttributes() {
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

// ColorFragmentShader represents a ColorFragmentShader
type ColorFragmentShader struct {
	uint32

	RenderMode        *uniforms.Int
	NumTilesX         *uniforms.UInt
	ZNear             *uniforms.Float
	ZFar              *uniforms.Float
	AmbientLightColor *uniforms.Vector3
	Diffuse           *uniforms.Sampler2D

	LightBuffer, VisibleLightIndicesBuffer, DirectionalLightBuffer *buffers.Binding

	ShadowMap1, ShadowMap2, ShadowMap3 *uniforms.Sampler2D
}

// NewColorFragmentShader instantiates and initializes a ColorFragmentShader object.
func NewColorFragmentShader() (*ColorFragmentShader, error) {
	program := gl.CreateProgram()
	shader := gl.CreateShader(gl.FRAGMENT_SHADER)

	csources, free := gl.Strs(colorShaderFragSrc)
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

		return nil, fmt.Errorf("failed to compile %v: %v", colorShaderOriginalFragmentSourceFile, log)
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

		return nil, fmt.Errorf("failed to link %v: %v", colorShaderOriginalFragmentSourceFile, log)
	}

	renderModeLoc := gl.GetUniformLocation(program, gl.Str("renderMode\x00"))
	numTilesXLoc := gl.GetUniformLocation(program, gl.Str("numTilesX\x00"))
	zNearLoc := gl.GetUniformLocation(program, gl.Str("zNear\x00"))
	zFarLoc := gl.GetUniformLocation(program, gl.Str("zFar\x00"))
	ambientLightColorLoc := gl.GetUniformLocation(program, gl.Str("ambientLightColor\x00"))
	diffuseLoc := gl.GetUniformLocation(program, gl.Str("diffuse\x00"))
	shadowMap1Loc := gl.GetUniformLocation(program, gl.Str("shadowMap1\x00"))
	shadowMap2Loc := gl.GetUniformLocation(program, gl.Str("shadowMap2\x00"))
	shadowMap3Loc := gl.GetUniformLocation(program, gl.Str("shadowMap3\x00"))

	gl.DeleteShader(shader)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	fs := &ColorFragmentShader{
		uint32:                    program,
		RenderMode:                uniforms.NewInt(program, renderModeLoc),
		NumTilesX:                 uniforms.NewUInt(program, numTilesXLoc),
		ZNear:                     uniforms.NewFloat(program, zNearLoc),
		ZFar:                      uniforms.NewFloat(program, zFarLoc),
		AmbientLightColor:         uniforms.NewVector3(program, ambientLightColorLoc),
		Diffuse:                   uniforms.NewSampler2D(program, diffuseLoc),
		LightBuffer:               buffers.NewBinding(0),
		VisibleLightIndicesBuffer: buffers.NewBinding(1),
		DirectionalLightBuffer:    buffers.NewBinding(2),
		ShadowMap1:                uniforms.NewSampler2D(program, shadowMap1Loc),
		ShadowMap2:                uniforms.NewSampler2D(program, shadowMap2Loc),
		ShadowMap3:                uniforms.NewSampler2D(program, shadowMap3Loc),
	}

	messagebus.RegisterType("key", func(m *messagebus.Message) {
		pressedKeys := m.Data1.([]glfw.Key)
		for _, key := range pressedKeys {
			if key >= glfw.KeyF1 && key <= glfw.KeyF25 {
				fs.RenderMode.Set(int32(key - glfw.KeyF1))
			}
		}
	})

	return fs, nil
}

// AddToPipeline adds this shader to the provided pipeline.
func (s *ColorFragmentShader) AddToPipeline(pipeline uint32) {
	gl.UseProgramStages(pipeline, gl.FRAGMENT_SHADER_BIT, s.uint32)
}
