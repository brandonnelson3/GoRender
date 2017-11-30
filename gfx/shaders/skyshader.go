package shaders

import (
	"fmt"
	"strings"

	"github.com/brandonnelson3/GoRender/gfx/buffers"
	"github.com/brandonnelson3/GoRender/gfx/uniforms"

	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	skyShaderOriginalVertexSourceFile = `skyshader.vert`
	skyShaderVertSrc                  = `
#version 450

uniform mat4 projection;
uniform mat4 view;

in vec3 vert;

out vec3 position;

void main() {
	gl_Position = projection * view * vec4(vert, 1);
	position = vert;
}` + "\x00"
	skyShaderOriginalFragmentSourceFile = `skyshader.frag`
	skyShaderFragSrc                    = `
// Significant credit here to: https://github.com/wwwtyro/glsl-atmosphere pretty much copy pasta.
#version 450

precision highp float;

struct DirectionalLight {
	vec3 color;
	float brightness;
	vec3 direction;
};

layout(std430, binding = 2) readonly buffer DirectionalLightBuffer {
	DirectionalLight data;
} directionalLightBuffer;

in vec3 position;

out vec4 outputColor;

#define PI 3.141592
#define iSteps 16
#define jSteps 8

vec2 rsi(vec3 r0, vec3 rd, float sr) {
    // ray-sphere intersection that assumes
    // the sphere is centered at the origin.
    // No intersection when result.x > result.y
    float a = dot(rd, rd);
    float b = 2.0 * dot(rd, r0);
    float c = dot(r0, r0) - (sr * sr);
    float d = (b*b) - 4.0*a*c;
    if (d < 0.0) return vec2(1e5,-1e5);
    return vec2(
        (-b - sqrt(d))/(2.0*a),
        (-b + sqrt(d))/(2.0*a)
    );
}

vec3 atmosphere(vec3 r, vec3 r0, vec3 pSun, float iSun, float rPlanet, float rAtmos, vec3 kRlh, float kMie, float shRlh, float shMie, float g) {
    // Normalize the sun and view directions.
    pSun = normalize(pSun);
    r = normalize(r);

    // Calculate the step size of the primary ray.
    vec2 p = rsi(r0, r, rAtmos);
    if (p.x > p.y) return vec3(0,0,0);
    p.y = min(p.y, rsi(r0, r, rPlanet).x);
    float iStepSize = (p.y - p.x) / float(iSteps);

    // Initialize the primary ray time.
    float iTime = 0.0;

    // Initialize accumulators for Rayleigh and Mie scattering.
    vec3 totalRlh = vec3(0,0,0);
    vec3 totalMie = vec3(0,0,0);

    // Initialize optical depth accumulators for the primary ray.
    float iOdRlh = 0.0;
    float iOdMie = 0.0;

    // Calculate the Rayleigh and Mie phases.
    float mu = dot(r, pSun);
    float mumu = mu * mu;
    float gg = g * g;
    float pRlh = 3.0 / (16.0 * PI) * (1.0 + mumu);
    float pMie = 3.0 / (8.0 * PI) * ((1.0 - gg) * (mumu + 1.0)) / (pow(1.0 + gg - 2.0 * mu * g, 1.5) * (2.0 + gg));

    // Sample the primary ray.
    for (int i = 0; i < iSteps; i++) {

        // Calculate the primary ray sample position.
        vec3 iPos = r0 + r * (iTime + iStepSize * 0.5);

        // Calculate the height of the sample.
        float iHeight = length(iPos) - rPlanet;

        // Calculate the optical depth of the Rayleigh and Mie scattering for this step.
        float odStepRlh = exp(-iHeight / shRlh) * iStepSize;
        float odStepMie = exp(-iHeight / shMie) * iStepSize;

        // Accumulate optical depth.
        iOdRlh += odStepRlh;
        iOdMie += odStepMie;

        // Calculate the step size of the secondary ray.
        float jStepSize = rsi(iPos, pSun, rAtmos).y / float(jSteps);

        // Initialize the secondary ray time.
        float jTime = 0.0;

        // Initialize optical depth accumulators for the secondary ray.
        float jOdRlh = 0.0;
        float jOdMie = 0.0;

        // Sample the secondary ray.
        for (int j = 0; j < jSteps; j++) {

            // Calculate the secondary ray sample position.
            vec3 jPos = iPos + pSun * (jTime + jStepSize * 0.5);

            // Calculate the height of the sample.
            float jHeight = length(jPos) - rPlanet;

            // Accumulate the optical depth.
            jOdRlh += exp(-jHeight / shRlh) * jStepSize;
            jOdMie += exp(-jHeight / shMie) * jStepSize;

            // Increment the secondary ray time.
            jTime += jStepSize;
        }

        // Calculate attenuation.
        vec3 attn = exp(-(kMie * (iOdMie + jOdMie) + kRlh * (iOdRlh + jOdRlh)));

        // Accumulate scattering.
        totalRlh += odStepRlh * attn;
        totalMie += odStepMie * attn;

        // Increment the primary ray time.
        iTime += iStepSize;

    }

    // Calculate and return the final color.
    return iSun * (pRlh * kRlh * totalRlh + pMie * kMie * totalMie);
}

void main() {	
	DirectionalLight directionalLight = directionalLightBuffer.data;
	
	vec3 color = atmosphere(
        normalize(position),// normalized ray direction
        vec3(0,6372e3,0),               // ray origin
        directionalLight.direction * -1,// position of the sun
        22.0,                           // intensity of the sun
        6371e3,                         // radius of the planet in meters
        6471e3,                         // radius of the atmosphere in meters
        vec3(5.5e-6, 13.0e-6, 22.4e-6), // Rayleigh scattering coefficient
        21e-6,                          // Mie scattering coefficient
        8e3,                            // Rayleigh scale height
        1.2e3,                          // Mie scale height
        0.758                           // Mie preferred scattering direction
    );

    // Apply exposure.
    color = 1.0 - exp(-1.0 * color);

    outputColor = vec4(color, 1);
}
` + "\x00"
)

// SkyShader is a Shader.
type SkyShader struct {
	shader

	Projection, View *uniforms.Matrix4

	DirectionalLightBuffer *buffers.Binding
}

// NewSkyShader instantiates and initializes a shader object.
func NewSkyShader() (*SkyShader, error) {
	program := gl.CreateProgram()

	// VertexShader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	vertexSrc, freeVertexSrc := gl.Strs(skyShaderVertSrc)
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
		return nil, fmt.Errorf("failed to compile %v: %v", skyShaderOriginalFragmentSourceFile, log)
	}
	gl.AttachShader(program, vertexShader)

	// FragmentShader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentSrc, freeFragmentSrc := gl.Strs(skyShaderFragSrc)
	gl.ShaderSource(fragmentShader, 1, fragmentSrc, nil)
	freeFragmentSrc()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to compile %v: %v", skyShaderOriginalFragmentSourceFile, log)
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
		return nil, fmt.Errorf("failed to link %v: %v", skyShaderOriginalVertexSourceFile, log)
	}

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	return &SkyShader{
		shader:     shader{program},
		Projection: uniforms.NewMatrix4(program, projectionLoc),
		View:       uniforms.NewMatrix4(program, viewLoc),
		DirectionalLightBuffer: buffers.NewBinding(2),
	}, nil
}
