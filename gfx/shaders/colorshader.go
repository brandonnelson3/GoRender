package shaders

import (
	"fmt"
	"strings"

	"github.com/brandonnelson3/GoRender/gfx/buffers"
	"github.com/brandonnelson3/GoRender/gfx/uniforms"

	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	colorShaderOriginalVertexSourceFile = `colorshader.vert`
	colorShaderVertSrc                  = `
#version 450

const int NUMBER_OF_CASCADES = 5;

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;
uniform mat4 lightViewProjs[NUMBER_OF_CASCADES];

layout(location = 0) in vec3 vert;
layout(location = 1) in vec3 norm;
layout(location = 2) in vec2 uv;

out vec4 position;
out vec3 worldPosition;
out vec3 norm_out;
out vec2 uv_out;	
out vec4 lightPositions[NUMBER_OF_CASCADES];

void main() {
	gl_Position = projection * view * model * vec4(vert, 1);
	position = projection * view * model * vec4(vert, 1);
	worldPosition = vec3(model * vec4(vert, 1));
	norm_out = normalize(mat3(transpose(inverse(model))) * norm);
	uv_out = uv;

	for (int i=0;i < NUMBER_OF_CASCADES; i++) {
		lightPositions[i] = lightViewProjs[i] * model * vec4(vert, 1);
	}
}` + "\x00"
	colorShaderOriginalFragmentSourceFile = `colorshader.frag`
	colorShaderFragSrc                    = `
#version 450

const int NUMBER_OF_CASCADES = 5;

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
uniform vec3 firstPersonPosition;
uniform vec3 firstPersonForward;
uniform sampler2D diffuse;
uniform sampler2DShadow shadowMap1;
uniform sampler2DShadow shadowMap2;
uniform sampler2DShadow shadowMap3;
uniform sampler2DShadow shadowMap4;
uniform sampler2DShadow shadowMap5;

// Point light shadows
const int MAX_POINT_SHADOW_LIGHTS = 4;
uniform samplerCubeArrayShadow pointShadowMaps;
uniform int   numPointShadowLights;
uniform vec3  pointShadowLightPositions[MAX_POINT_SHADOW_LIGHTS];
uniform float pointShadowFarPlane;

in vec4 position;
in vec3 worldPosition;
in vec3 norm_out;
in vec2 uv_out;	
in vec4 lightPositions[NUMBER_OF_CASCADES];

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

float getShadowFactor(int index, vec3 projCoords, int radius)
{		
	float texelSize = 1.0 / shadowMapSize;
	float shadowFactor = 0.0f;
	float count = 0.0;
	
	// Hardware PCF (2x2) is already applied by texture() for sampler2DShadow.
	// We still loop a bit for even smoother results (softening the 2x2 edges).
	for (int i=-radius; i<=radius; i++) {
		for (int j=-radius; j<=radius; j++) {
			vec3 shadowUV = vec3(projCoords.xy + vec2(i,j) * texelSize, projCoords.z);
			if(index == 0) shadowFactor += texture(shadowMap1, shadowUV);
			else if(index == 1) shadowFactor += texture(shadowMap2, shadowUV);
			else if(index == 2) shadowFactor += texture(shadowMap3, shadowUV);
			else if(index == 3) shadowFactor += texture(shadowMap4, shadowUV);
			else shadowFactor += texture(shadowMap5, shadowUV);
			count += 1.0;
		}
	}
	return shadowFactor / count;
}

// Poisson disk samples for softening point light shadows (12 samples).
vec3 poissonDisk[12] = vec3[]
(
   vec3(-0.5212691, -0.4013232,  0.5125319),
   vec3(-0.7924651,  0.1578255, -0.2209583),
   vec3(-0.3851416,  0.7363506,  0.3151242),
   vec3( 0.1652817,  0.1313845,  0.6511982),
   vec3( 0.4033304,  0.4752404, -0.5239012),
   vec3(-0.8358865, -0.3427192, -0.1158563),
   vec3( 0.2826883, -0.0479542, -0.6661914),
   vec3( 0.8525992,  0.0285867,  0.2790367),
   vec3( 0.4161608, -0.7312921,  0.1224869),
   vec3(-0.1114824, -0.7151128, -0.4411124),
   vec3( 0.1924824,  0.7711284, -0.1411124),
   vec3( 0.7114824,  0.4151128,  0.2411124)
);

// Returns a value in [0,1] where 0.0 is full shadow and 1.0 is full light.
float getPointShadowFactor(int slot, vec3 worldPos, vec3 normal) {
	vec3 fragToLight = worldPos - pointShadowLightPositions[slot];
	float currentDepth = length(fragToLight);
	
	vec3 lightDir = normalize(-fragToLight);
	float bias = max(0.15 * (1.0 - dot(normal, lightDir)), 0.05); 
	
	float shadow = 0.0;
	// Distant point shadows use fewer samples.
	int samples = (currentDepth > 30.0) ? 4 : 12;
	float diskRadius = 0.05;
	
	for(int i = 0; i < samples; ++i) {
		vec4 shadowUV = vec4(fragToLight + poissonDisk[i] * diskRadius, slot);
		// samplerCubeArrayShadow texture() takes vec4(dir, layer) and a reference depth as the last arg.
		// Wait, for samplerCubeArrayShadow it's texture(sampler, vec4(dir, layer), ref)?
		// Actually it's texture(sampler, vec4(dir, layer), ref).
		shadow += texture(pointShadowMaps, shadowUV, (currentDepth - bias) / pointShadowFarPlane);
	}
	return shadow / float(samples);
}


void main() {
	ivec2 location = ivec2(gl_FragCoord.xy);
	// TODO: Put this 16 somewhere constant.
	ivec2 tileID = location / ivec2(16, 16);
	uint index = tileID.y * numTilesX + tileID.x;

	// TODO 1024 should be somewhere constant.
	uint offset = index * 1024;
	
	if (renderMode == 0 || renderMode == 5) {		
		vec4 diffuseColor = texture(diffuse, uv_out);
		if (diffuseColor.a < 0.5) {
			discard;
		} 
		
		vec3 pointLightColor = vec3(0, 0, 0);
		uint i=0;
		for (i=0; i < 1024 && visibleLightIndicesBuffer.data[offset + i].index != -1; i++) {
			uint lightIndex = visibleLightIndicesBuffer.data[offset + i].index;
			PointLight light = lightBuffer.data[lightIndex];
			vec3 lightVector = light.position - worldPosition;
			float dist = length(lightVector);
			vec3 lightDir = normalize(lightVector);
			float diff = max(dot(norm_out, lightDir), 0.0);
			float attenuation = max(1.0 - (dist / light.radius), 0.0);

			// Point light shadow
			float plShadow = 1.0;
			for (int s = 0; s < numPointShadowLights && s < MAX_POINT_SHADOW_LIGHTS; s++) {
				if (distance(pointShadowLightPositions[s], light.position) < 0.1) {
					plShadow = getPointShadowFactor(s, worldPosition, norm_out);
					break;
				}
			}

			pointLightColor += plShadow * light.color * light.intensity * diff * attenuation;
		}
		
		DirectionalLight directionalLight = directionalLightBuffer.data;
		float NdL = max(0.0f, dot(norm_out, -1*directionalLight.direction));
		vec3 directionalLightColor = (NdL) * directionalLight.color * directionalLight.brightness;
		float depthTest = dot(worldPosition - firstPersonPosition, firstPersonForward);

		vec3 shadowCoords[5] = vec3[](
			lightPositions[0] * 0.5 + 0.5, 
			lightPositions[1] * 0.5 + 0.5, 
			lightPositions[2] * 0.5 + 0.5,
			lightPositions[3] * 0.5 + 0.5,
			lightPositions[4] * 0.5 + 0.5
		);

		int shadowIndex = 5;
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
		} else if((saturatef(shadowCoords[4].x) == shadowCoords[4].x) && (saturatef(shadowCoords[4].y) == shadowCoords[4].y) && depthTest < cascadeDepthLimits[5]){
			shadowIndex = 4;
			shadowIndexColor = vec3(.5, 1, 1);
		}

		if (renderMode == 0) {
			shadowIndexColor = vec3(1, 1, 1);
		}
		
		float shadowFactor = 1.0f;	
		if (shadowIndex != 5) {
			// Distant cascades use fewer samples (radius 0 or 1).
			int radius = (shadowIndex > 2) ? 0 : 1;
			shadowFactor = getShadowFactor(shadowIndex, shadowCoords[shadowIndex], radius);
		}		
		
		vec3 ambientLight = directionalLight.color * directionalLight.brightness * 0.2f;

		outputColor = diffuseColor * vec4(shadowIndexColor, 1.0) * vec4(directionalLightColor*shadowFactor + ambientLight, 1.0) + diffuseColor * vec4(shadowIndexColor, 1.0) * vec4(pointLightColor, 1.0);
	} else if (renderMode == 1) {
		uint i=0;
		for (i; i < 1024 && visibleLightIndicesBuffer.data[offset + i].index != -1; i++) {}
		outputColor = vec4(vec3(float(i)/256)+vec3(0.1), 1.0);
	} else if (renderMode == 2) {
		outputColor = vec4(abs(norm_out), 1.0);
	} else if (renderMode == 3) {
		outputColor = vec4(uv_out, 0, 1.0);
	} else if (renderMode == 4) {
		vec4 diffuseColor = texture(diffuse, uv_out);
		if (diffuseColor.a < 0.5) {
			discard;
		} 
		outputColor = diffuseColor;
	} 
}
` + "\x00"
)

// ColorShader is a Shader.
type ColorShader struct {
	shader

	Projection, View, Model *uniforms.Matrix4
	LightViewProjs          *uniforms.Matrix4Array

	RenderMode         *uniforms.Int
	NumTilesX          *uniforms.UInt
	ZNear              *uniforms.Float
	ZFar               *uniforms.Float
	ShadowMapSize      *uniforms.Float
	AmbientLightColor  *uniforms.Vector3
	CascadeDepthLimits *uniforms.FloatArray
	FirstPersonPosition *uniforms.Vector3
	FirstPersonForward  *uniforms.Vector3
	Diffuse            *uniforms.Sampler2D

	LightBuffer, VisibleLightIndicesBuffer, DirectionalLightBuffer *buffers.Binding

	ShadowMap1, ShadowMap2, ShadowMap3, ShadowMap4, ShadowMap5 *uniforms.Sampler2D

	// Point light shadow cubemaps
	PointShadowMaps           *uniforms.SamplerCubeArrayTexture
	NumPointShadowLights      *uniforms.Int
	PointShadowLightPositions *uniforms.Vector3Array
	PointShadowFarPlane       *uniforms.Float
}

// NewColorShader instantiates and initializes a shader object.
func NewColorShader() (*ColorShader, error) {
	program := gl.CreateProgram()

	// VertexShader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	vertexSrc, freeVertexSrc := gl.Strs(colorShaderVertSrc)
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
		return nil, fmt.Errorf("failed to compile %v: %v", colorShaderOriginalVertexSourceFile, log)
	}
	gl.AttachShader(program, vertexShader)

	// FragmentShader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentSrc, freeFragmentSrc := gl.Strs(colorShaderFragSrc)
	gl.ShaderSource(fragmentShader, 1, fragmentSrc, nil)
	freeFragmentSrc()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile %v: %v", colorShaderOriginalFragmentSourceFile, log)
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
		return nil, fmt.Errorf("failed to link %v: %v", colorShaderOriginalVertexSourceFile, log)
	}

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))
	modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))
	lightViewProjsLoc := gl.GetUniformLocation(program, gl.Str("lightViewProjs\x00"))
	renderModeLoc := gl.GetUniformLocation(program, gl.Str("renderMode\x00"))
	numTilesXLoc := gl.GetUniformLocation(program, gl.Str("numTilesX\x00"))
	zNearLoc := gl.GetUniformLocation(program, gl.Str("zNear\x00"))
	zFarLoc := gl.GetUniformLocation(program, gl.Str("zFar\x00"))
	shadowMapSizeLoc := gl.GetUniformLocation(program, gl.Str("shadowMapSize\x00"))
	ambientLightColorLoc := gl.GetUniformLocation(program, gl.Str("ambientLightColor\x00"))
	cascadeDepthLimitsLoc := gl.GetUniformLocation(program, gl.Str("cascadeDepthLimits\x00"))
	firstPersonPositionLoc := gl.GetUniformLocation(program, gl.Str("firstPersonPosition\x00"))
	firstPersonForwardLoc := gl.GetUniformLocation(program, gl.Str("firstPersonForward\x00"))
	diffuseLoc := gl.GetUniformLocation(program, gl.Str("diffuse\x00"))
	shadowMap1Loc := gl.GetUniformLocation(program, gl.Str("shadowMap1\x00"))
	shadowMap2Loc := gl.GetUniformLocation(program, gl.Str("shadowMap2\x00"))
	shadowMap3Loc := gl.GetUniformLocation(program, gl.Str("shadowMap3\x00"))
	shadowMap4Loc := gl.GetUniformLocation(program, gl.Str("shadowMap4\x00"))
	shadowMap5Loc := gl.GetUniformLocation(program, gl.Str("shadowMap5\x00"))
	pointShadowMapsLoc := gl.GetUniformLocation(program, gl.Str("pointShadowMaps\x00"))
	numPointShadowLightsLoc := gl.GetUniformLocation(program, gl.Str("numPointShadowLights\x00"))
	pointShadowLightPositionsLoc := gl.GetUniformLocation(program, gl.Str("pointShadowLightPositions\x00"))
	pointShadowFarPlaneLoc := gl.GetUniformLocation(program, gl.Str("pointShadowFarPlane\x00"))

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	return &ColorShader{
		shader:                    shader{program},
		Projection:                uniforms.NewMatrix4(program, projectionLoc),
		View:                      uniforms.NewMatrix4(program, viewLoc),
		Model:                     uniforms.NewMatrix4(program, modelLoc),
		LightViewProjs:            uniforms.NewMatrix4Array(program, lightViewProjsLoc),
		RenderMode:                uniforms.NewInt(program, renderModeLoc),
		NumTilesX:                 uniforms.NewUInt(program, numTilesXLoc),
		ZNear:                     uniforms.NewFloat(program, zNearLoc),
		ZFar:                      uniforms.NewFloat(program, zFarLoc),
		ShadowMapSize:             uniforms.NewFloat(program, shadowMapSizeLoc),
		AmbientLightColor:         uniforms.NewVector3(program, ambientLightColorLoc),
		CascadeDepthLimits:        uniforms.NewFloatArray(program, cascadeDepthLimitsLoc),
		FirstPersonPosition:       uniforms.NewVector3(program, firstPersonPositionLoc),
		FirstPersonForward:        uniforms.NewVector3(program, firstPersonForwardLoc),
		Diffuse:                   uniforms.NewSampler2D(program, diffuseLoc),
		LightBuffer:               buffers.NewBinding(0),
		VisibleLightIndicesBuffer: buffers.NewBinding(1),
		DirectionalLightBuffer:    buffers.NewBinding(2),
		ShadowMap1:                uniforms.NewSampler2D(program, shadowMap1Loc),
		ShadowMap2:                uniforms.NewSampler2D(program, shadowMap2Loc),
		ShadowMap3:                uniforms.NewSampler2D(program, shadowMap3Loc),
		ShadowMap4:                uniforms.NewSampler2D(program, shadowMap4Loc),
		ShadowMap5:                uniforms.NewSampler2D(program, shadowMap5Loc),
		PointShadowMaps:           uniforms.NewSamplerCubeArrayTexture(program, pointShadowMapsLoc),
		NumPointShadowLights:      uniforms.NewInt(program, numPointShadowLightsLoc),
		PointShadowLightPositions: uniforms.NewVector3Array(program, pointShadowLightPositionsLoc),
		PointShadowFarPlane:       uniforms.NewFloat(program, pointShadowFarPlaneLoc),
	}, nil
}
