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

const int NUMBER_OF_CASCADES = 4;

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;
uniform mat4 lightViewProjs[NUMBER_OF_CASCADES];

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
	vec4 lightPositions[NUMBER_OF_CASCADES];
} vertex_out;

void main() {
	gl_Position = projection * view * model * vec4(vert, 1);
	vertex_out.position = projection * view * model * vec4(vert, 1);
	vertex_out.worldPosition = vec3(model * vec4(vert, 1));
	vertex_out.normal = norm;
	vertex_out.uv = uv;

	for (int i=0;i < NUMBER_OF_CASCADES; i++) {
		vertex_out.lightPositions[i] = lightViewProjs[i] * model * vec4(vert, 1);
	}
}` + "\x00"
	colorShaderOriginalFragmentSourceFile = `colorshader.frag`
	colorShaderFragSrc                    = `
#version 450

const int NUMBER_OF_CASCADES = 4;

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
uniform float shadowMapSize;
uniform vec3 ambientLightColor;
uniform float cascadeDepthLimits[NUMBER_OF_CASCADES + 1];
uniform sampler2D diffuse;
uniform sampler2D shadowMap1;
uniform sampler2D shadowMap2;
uniform sampler2D shadowMap3;
uniform sampler2D shadowMap4;

// Portion of the depth to consider to prevent shadow acne. This is proportional to the depth, to prevent peter panning.
const float shadowBias = 0.9995;

in VERTEX_OUT
{
	vec4 gl_FragCoord;
	vec4 position;
	vec3 worldPosition;
	vec3 normal;
	vec2 uv;	
	vec4 lightPositions[NUMBER_OF_CASCADES];
} fragment_in;

out vec4 outputColor;

float linearize(float depth)
{
	return (2 * zNear) / (zFar + zNear - depth * (zFar - zNear));
}

vec3 saturate(vec3 v) {
	return vec3(clamp(v.x, 0.0, 1.0), clamp(v.y, 0.0, 1.0), clamp(v.z, 0.0, 1.0));
}

float saturatef(float f) {
	return clamp(f, 0.0, 1.0);
}

float getShadowFactor(int index, vec3 projCoords)
{		
	float texelSize = 1.0 / shadowMapSize;
	float currentDepth = projCoords.z;
	float shadowFactor = 1.0f;
	for (int i=-1; i<=1; i++) {
		for (int j=-1; j<=1; j++) {
			float shadowMapDepth = 0.0f;
			if(index == 0) {
				shadowMapDepth = texture(shadowMap1, projCoords.xy + vec2(i,j) * texelSize).x;
			} else if(index == 1) {
				shadowMapDepth = texture(shadowMap2, projCoords.xy + vec2(i,j) * texelSize).x;
			} else if(index == 2) {
				shadowMapDepth = texture(shadowMap3, projCoords.xy + vec2(i,j) * texelSize).x;
			} else {
				shadowMapDepth = texture(shadowMap4, projCoords.xy + vec2(i,j) * texelSize).x;
			}			
			if (linearize(currentDepth*shadowBias) > linearize(shadowMapDepth)) {
				shadowFactor -= 0.1f;
			}
		}
	}
	return shadowFactor;
}

void main() {
	ivec2 location = ivec2(gl_FragCoord.xy);
	// TODO: Put this 16 somewhere constant.
	ivec2 tileID = location / ivec2(16, 16);
	uint index = tileID.y * numTilesX + tileID.x;

	// TODO 1024 should be somewhere constant.
	uint offset = index * 1024;
	
	if (renderMode == 0 || renderMode == 5) {		
		vec4 diffuseColor = texture(diffuse, fragment_in.uv);
		if (diffuseColor.a != 1) {
			discard;
		} 
		
		vec3 pointLightColor = vec3(0, 0, 0);
		uint i=0;
		for (i=0; i < 1024 && visibleLightIndicesBuffer.data[offset + i].index != -1; i++) {
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
		float depthTest = fragment_in.position.z;

		vec3 shadowCoords[4] = vec3[](
			fragment_in.lightPositions[0] * 0.5 + 0.5, 
			fragment_in.lightPositions[1] * 0.5 + 0.5, 
			fragment_in.lightPositions[2] * 0.5 + 0.5,
			fragment_in.lightPositions[3] * 0.5 + 0.5
		);

		int shadowIndex = 4;
		vec3 shadowIndexColor = vec3(1, 1, 1);
		if ((saturatef(shadowCoords[0].x) == shadowCoords[0].x) && (saturatef(shadowCoords[0].y) == shadowCoords[0].y) && depthTest < cascadeDepthLimits[1]) {			
			shadowIndex = 0;
			shadowIndexColor = vec3(1, .5, .5);
		} else if((saturatef(shadowCoords[1].x) == shadowCoords[1].x) && (saturatef(shadowCoords[1].y) == shadowCoords[1].y) && depthTest < cascadeDepthLimits[2]) {
			shadowIndex = 1;
			shadowIndexColor = vec3(.5, 1, .5);
		} else if((saturatef(shadowCoords[2].x) == shadowCoords[2].x) && (saturatef(shadowCoords[2].y) == shadowCoords[2].y) && depthTest < cascadeDepthLimits[3]) {
			shadowIndex = 2;
			shadowIndexColor = vec3(.5, .5, 1);
		} else if((saturatef(shadowCoords[3].x) == shadowCoords[3].x) && (saturatef(shadowCoords[3].y) == shadowCoords[3].y) && depthTest < cascadeDepthLimits[4]){
			shadowIndex = 3;
			shadowIndexColor = vec3(1, 1, .5);
		}

		if (renderMode == 0) {
			shadowIndexColor = vec3(1, 1, 1);
		}
		
		float shadowFactor = 1.0f;	
		if (shadowIndex != 4) {
			shadowFactor = getShadowFactor(shadowIndex, shadowCoords[shadowIndex]);
		}		
		float sunHeight = dot(directionalLight.direction * -1, vec3(0, 1, 0));
		if (sunHeight < 0) {
			shadowFactor = 0.0;
		}	

		outputColor = diffuseColor * vec4(shadowIndexColor, 1.0) * vec4(directionalLightColor*shadowFactor, 1.0) + diffuseColor * vec4(shadowIndexColor, 1.0) * vec4(pointLightColor, 1.0);
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

	Projection, View, Model *uniforms.Matrix4
	LightViewProjs          *uniforms.Matrix4Array
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
	lightViewProjsLoc := gl.GetUniformLocation(program, gl.Str("lightViewProjs\x00"))

	gl.DeleteShader(shader)

	return &ColorVertexShader{
		uint32:         program,
		Projection:     uniforms.NewMatrix4(program, projectionLoc),
		View:           uniforms.NewMatrix4(program, viewLoc),
		Model:          uniforms.NewMatrix4(program, modelLoc),
		LightViewProjs: uniforms.NewMatrix4Array(program, lightViewProjsLoc),
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

	RenderMode         *uniforms.Int
	NumTilesX          *uniforms.UInt
	ZNear              *uniforms.Float
	ZFar               *uniforms.Float
	ShadowMapSize      *uniforms.Float
	AmbientLightColor  *uniforms.Vector3
	CascadeDepthLimits *uniforms.FloatArray
	Diffuse            *uniforms.Sampler2D

	LightBuffer, VisibleLightIndicesBuffer, DirectionalLightBuffer *buffers.Binding

	ShadowMap1, ShadowMap2, ShadowMap3, ShadowMap4 *uniforms.Sampler2D
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
	shadowMapSizeLoc := gl.GetUniformLocation(program, gl.Str("shadowMapSize\x00"))
	ambientLightColorLoc := gl.GetUniformLocation(program, gl.Str("ambientLightColor\x00"))
	cascadeDepthLimitsLoc := gl.GetUniformLocation(program, gl.Str("cascadeDepthLimits\x00"))
	diffuseLoc := gl.GetUniformLocation(program, gl.Str("diffuse\x00"))
	shadowMap1Loc := gl.GetUniformLocation(program, gl.Str("shadowMap1\x00"))
	shadowMap2Loc := gl.GetUniformLocation(program, gl.Str("shadowMap2\x00"))
	shadowMap3Loc := gl.GetUniformLocation(program, gl.Str("shadowMap3\x00"))
	shadowMap4Loc := gl.GetUniformLocation(program, gl.Str("shadowMap4\x00"))

	gl.DeleteShader(shader)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	fs := &ColorFragmentShader{
		uint32:                    program,
		RenderMode:                uniforms.NewInt(program, renderModeLoc),
		NumTilesX:                 uniforms.NewUInt(program, numTilesXLoc),
		ZNear:                     uniforms.NewFloat(program, zNearLoc),
		ZFar:                      uniforms.NewFloat(program, zFarLoc),
		ShadowMapSize:             uniforms.NewFloat(program, shadowMapSizeLoc),
		AmbientLightColor:         uniforms.NewVector3(program, ambientLightColorLoc),
		CascadeDepthLimits:        uniforms.NewFloatArray(program, cascadeDepthLimitsLoc),
		Diffuse:                   uniforms.NewSampler2D(program, diffuseLoc),
		LightBuffer:               buffers.NewBinding(0),
		VisibleLightIndicesBuffer: buffers.NewBinding(1),
		DirectionalLightBuffer:    buffers.NewBinding(2),
		ShadowMap1:                uniforms.NewSampler2D(program, shadowMap1Loc),
		ShadowMap2:                uniforms.NewSampler2D(program, shadowMap2Loc),
		ShadowMap3:                uniforms.NewSampler2D(program, shadowMap3Loc),
		ShadowMap4:                uniforms.NewSampler2D(program, shadowMap4Loc),
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
